package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func loadWasm(release, build bool) []byte {
	path := debugPath
	if release {
		path = releasePath
	}
	// if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
	if build {
		cmd := exec.Command("cargo", "build", "--target=wasm32-wasi")
		if release {
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

func startWz(wasm []byte, ctx context.Context) api.Module {
	cache := wazero.NewCompilationCache()
	cfg := wazero.NewRuntimeConfig().WithCompilationCache(cache)
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
	debugPath   = "target/wasm32-wasi/debug/wazero-crash-lib.wasm"
	releasePath = "target/wasm32-wasi/release/wazero-crash-lib.wasm"
)

func writeStr(mod api.Module, str string) (ptr uint64, cleanup func()) {
	alloc := mod.ExportedFunction("allocate")
	if alloc == nil {
		log.Fatal("alloc not found")
	}
	dealloc := mod.ExportedFunction("deallocate")
	if dealloc == nil {
		log.Fatal("dealloc not found")
	}
	ctx := context.Background()
	res, err := alloc.Call(ctx, uint64(len(str)))
	if err != nil {
		log.Fatalf("calling alloc: %s", err)
	}
	if len(res) != 1 || res[0] == 0 {
		log.Fatalf("want pointer, got %v", res)
	}

	if ok := mod.Memory().WriteString(uint32(ptr), str); !ok {
		// return fmt.Errorf("failed writing %d bytes (%q) to wasm memory at %d; mem size %d", len(alCfg.logDir), alCfg.logDir, dirp, pw.mod.Memory().Size())
		log.Fatalf("failed writing %d bytes (%q) to wasm memory at %d; mem size %d", len(str), str, ptr, mod.Memory().Size())
	}
	cleanup = func() {
		dealloc.Call(ctx, ptr)
	}
	return ptr, cleanup
}

func main() {
	release := flag.Bool("release", false, "use release build rather than debug")
	skipBuild := flag.Bool("skipbuild", false, "skip cargo wasm (re)build")
	flag.Parse()

	// load
	wasm := loadWasm(*release, !*skipBuild)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mod := startWz(wasm, ctx)
	listFns(mod)

	s := "mello world vectors"
	reCheck(mod, ctx, s)
	acCheck(mod, ctx, s)
}

func reCheck(mod api.Module, ctx context.Context, s string) {
	m := mod.ExportedFunction("regex")
	if m == nil {
		log.Fatal("no such exported function regex")
	}
	ptr, cleanup := writeStr(mod, s)
	defer cleanup()
	res, err := m.Call(ctx, ptr, uint64(len(s)))
	if err != nil {
		log.Fatalf("call regex(): %s", err)
	}
	log.Println(res)
}

func acCheck(mod api.Module, ctx context.Context, s string) {
	m := mod.ExportedFunction("ac")
	if m == nil {
		log.Fatal("no such exported function ac")
	}
	ptr, cleanup := writeStr(mod, s)
	defer cleanup()

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
		log.Printf("  %s: %d params", f, len(d.ParamTypes()))
	}
}
