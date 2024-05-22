package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func loadWasm(opts wasmOpts) []byte {
	path := debugPath
	if opts.release {
		path = releasePath
	}
	if opts.proprietary {
		opts.skipBuild = true
		path = dbgProprietary
		if _, err := os.Stat(path); err != nil {
			opts.release = true
			log.Printf("debug build not found, trying release")
		}
		if opts.release {
			path = relProprietary
		}
	}
	if !opts.skipBuild {
		cmd := exec.Command("cargo", "build", "--target=wasm32-wasi")
		if opts.release {
			cmd.Args = append(cmd.Args, "--release")
		}
		out, err := cmd.CombinedOutput()
		if err != nil {
			buildHelp := "========\ndo you have rust and its target wasm32-wasi installed?"
			log.Fatalf("running %v: %s\noutput:\n%s\n%s", cmd.Args, err, string(out), buildHelp)
		}
	}
	wasm, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("reading wasm failed: %s", err)
	}
	return wasm
}

func startWz(wasm []byte, ctx context.Context, interp bool) api.Module {
	var cfg wazero.RuntimeConfig
	if interp {
		cfg = wazero.NewRuntimeConfigInterpreter()
	} else {
		cache := wazero.NewCompilationCache()
		cfg = wazero.NewRuntimeConfig().WithCompilationCache(cache)
	}
	wz := wazero.NewRuntimeWithConfig(ctx, cfg)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, wz); err != nil {
		log.Fatalf("snapshot instantiation failed: %s", err)
	}

	// compile once and cache
	compiled, err := wz.CompileModule(ctx, wasm)
	if err != nil {
		log.Fatalf("compiling: %s", err)
	}
	modcfg := wazero.NewModuleConfig().WithName("") //.WithEnv("RUST_BACKTRACE", "full")

	mod, err := wz.InstantiateModule(ctx, compiled, modcfg)
	if err != nil {
		log.Fatalf("instantiating: %s", err)
	}
	return mod
}

const (
	debugPath      = "target/wasm32-wasi/debug/wazero-crash-lib.wasm"
	releasePath    = "target/wasm32-wasi/release/wazero-crash-lib.wasm"
	dbgProprietary = "./proprietary_dbg.wasm"
	relProprietary = "./proprietary_rel.wasm"
)

func alloc(mod api.Module, ctx context.Context, n uint64) (ptr uint64, cleanup func()) {
	alloc := auditedFn{Function: mod.ExportedFunction("allocate")}
	if alloc.Function == nil {
		log.Fatal("alloc not found")
	}
	dealloc := auditedFn{Function: mod.ExportedFunction("deallocate")}
	if dealloc.Function == nil {
		log.Fatal("dealloc not found")
	}
	res, err := alloc.Call(ctx, n)
	if err != nil {
		log.Fatalf("calling alloc: %s", err)
	}
	if len(res) != 1 || res[0] == 0 {
		log.Fatalf("want pointer, got %v", res)
	}
	return res[0], func() {
		dealloc.Call(ctx, ptr, n)
	}
}

func writeData[T []byte | string](mod api.Module, ctx context.Context, data T) (ptr uint64, cleanup func()) {
	ptr, cleanup = alloc(mod, ctx, uint64(len(data)))
	if ok := mod.Memory().WriteString(uint32(ptr), string(data)); !ok {
		log.Fatalf("failed writing %d bytes (%q) to wasm memory at %d; mem size %d", len(data), data, ptr, mod.Memory().Size())
	}
	return ptr, cleanup
}

var nocleanup bool

type wasmOpts struct {
	release     bool
	skipBuild   bool
	proprietary bool
	nomem       bool
}

func main() {
	s := "mello world vectors"
	var wOpts wasmOpts
	flag.BoolVar(&wOpts.release, "release", false, "use release build rather than debug")
	flag.BoolVar(&wOpts.skipBuild, "skipbuild", false, "skip cargo wasm (re)build")
	flag.BoolVar(&nocleanup, "nocleanup", false, "skips dealloc's")
	flag.StringVar(&s, "str", s, "specify an arbitrary string to match re (orld) and/or ac (short, vectors)")
	flag.BoolVar(&wOpts.proprietary, "proprietary", false, "if true, exercise fn in proprietary wasm. obviates -skipbuild, -str.")
	flag.BoolVar(&wOpts.nomem, "nomem", false, "if true, do not grow memory")
	interp := flag.Bool("interp", false, "if true, use interpreter rather than compiling")
	flag.Parse()

	// load
	wasm := loadWasm(wOpts)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mod := startWz(wasm, ctx, *interp)

	if !wOpts.nomem {
		growMemory(mod)
	}

	runInit(mod, ctx)

	if !wOpts.proprietary {
		listFns(mod)
		reCheck(mod, ctx, s)
		acCheck(mod, ctx, s)
	}
}

func reCheck(mod api.Module, ctx context.Context, s string) {
	m := auditedFn{Function: mod.ExportedFunction("regex")}
	if m.Function == nil {
		log.Fatal("no such exported function regex")
	}
	ptr, cleanup := writeData(mod, ctx, s)
	if !nocleanup {
		defer cleanup()
	}
	res, err := m.Call(ctx, ptr, uint64(len(s)))
	if err != nil {
		log.Fatalf("call regex(): %s", err)
	}
	log.Println(res)
}

func acCheck(mod api.Module, ctx context.Context, s string) {
	m := auditedFn{Function: mod.ExportedFunction("ac")}
	if m.Function == nil {
		log.Fatal("no such exported function ac")
	}
	ptr, cleanup := writeData(mod, ctx, s)
	if !nocleanup {
		defer cleanup()
	}

	res, err := m.Call(ctx, ptr, uint64(len(s)))
	if err != nil {
		log.Fatalf("call ac(): %s", err)
	}
	log.Println(res)
}

func listFns(mod api.Module) {
	fns := mod.ExportedFunctionDefinitions()
	log.Printf("%d funcs exported:", len(fns))
	for f, d := range fns {
		log.Printf("  %s([%d]) -> %d", f, len(d.ParamTypes()), len(d.ResultTypes()))
	}
}

type auditedFn struct {
	api.Function
}

// must implement api.Function
var _ api.Function = (*auditedFn)(nil)

func (af auditedFn) Call(ctx context.Context, params ...uint64) (res []uint64, err error) {
	defer func() {
		errstr := "(nil)"
		if err != nil {
			errstr = err.Error()
		}
		log.Printf("%s(%v)=%v,%s", af.Definition().Name(), params, res, errstr)
	}()
	return af.Function.Call(ctx, params...)
}

func growMemory(mod api.Module) {
	const (
		page        = 65536 // page size from wazero documentation
		mb          = 1024 * 1024
		desiredMegs = 8
		desiredPgs  = desiredMegs * mb / page
	)
	currentPgs := mod.Memory().Size() / page
	delta := desiredPgs - currentPgs
	if delta > 0 {
		mod.Memory().Grow(delta)
	}
}

// - module memory unaffected: succeeds
// - interpreted: succeeds
// - module memory grown _and_ compiled: fails
func runInit(mod api.Module, ctx context.Context) {
	ini := auditedFn{Function: mod.ExportedFunction("init")}
	if ini.Function == nil {
		log.Fatalf("init function not found")
	}
	res, err := ini.Call(ctx)
	if err != nil {
		s := err.Error()
		if !strings.HasPrefix(s, "wasm error: unreachable") {
			redfatal("error does not begin with unreachable")
			os.Exit(1)
		}
		log.Fatal("init() error")
	}
	if len(res) != 1 || res[0] != 0 {
		log.Fatalf("init() did not succeed:\nres=%v", res)
	}
}

func redfatal(s string) {
	blackOnRed := "\x1b[101m"
	reset := "\x1b(B\x1b[m"
	fmt.Fprintf(os.Stderr, "%s%s%s\n", blackOnRed, s, reset)
	os.Exit(1)
}
