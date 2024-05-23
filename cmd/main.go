package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func loadWasm(opts wasmOpts) []byte {
	path := dbgProprietary
	if _, err := os.Stat(path); err != nil {
		log.Printf("debug build not found, trying release")
		opts.release = true
	}
	if opts.release {
		path = relProprietary
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
	modcfg := wazero.NewModuleConfig().WithName("")

	mod, err := wz.InstantiateModule(ctx, compiled, modcfg)
	if err != nil {
		log.Fatalf("instantiating: %s", err)
	}
	return mod
}

const (
	dbgProprietary = "./proprietary_dbg.wasm"
	relProprietary = "./proprietary_rel.wasm"
)

type wasmOpts struct {
	release bool
	nomem   bool
}

func main() {
	var wOpts wasmOpts
	flag.BoolVar(&wOpts.release, "release", false, "use release build rather than debug")
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
}

type auditedFn struct {
	api.Function
}

// must implement api.Function
var _ api.Function = (*auditedFn)(nil)

// logs args and results for each call
func (af auditedFn) Call(ctx context.Context, params ...uint64) (res []uint64, err error) {
	defer func() {
		errstr := "(nil)"
		if err != nil {
			errstr = err.Error()
		}
		log.Printf("call %s(%v) ->\n  res=%v\n  err= \\\n%s", af.Definition().Name(), params, res, errstr)
	}()
	return af.Function.Call(ctx, params...)
}

// grows module memory to 8 mb
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
		log.Printf("growing memory by %d pages", delta)
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
			// returned an error, but it differs from that expected
			redfatal("error does not begin with unreachable")
			os.Exit(1)
		}
		// actual error logged by auditedFn, no need to log again
		log.Fatal("init() error")
	}
	if len(res) != 1 || res[0] != 0 {
		log.Fatalf("init() did not succeed:\nres=%v", res)
	}
}

// uses ansi escape sequences to color message
func redfatal(s string) {
	blackOnRed := "\x1b[101m"
	reset := "\x1b(B\x1b[m"
	fmt.Fprintf(os.Stderr, "%s%s%s\n", blackOnRed, s, reset)
	os.Exit(1)
}
