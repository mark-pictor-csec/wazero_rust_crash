## quick start

go run ./cmd -skipbuild
  - wasm crashes when regex() func called

go run ./cmd -release -skipbuild
  - no crash

## versions

```
$ rustc --version
rustc 1.78.0 (9b00956e5 2024-04-29)

$ go version
go version go1.22.0 darwin/amd64
```

wasm v1.7.1

## files
* `crates/wazero-crash-lib/src/main.rs`: code for wasm module
  * can also run as native:
    * `cargo run`
    * `cargo run --release`
* `target/wasm32-wasi/*/*.wasm`: debug and release build of module, built with rust `1.78.0`
* `cmd/main.go`: builds and runs the wasm module, calling the exported function `regex()`
  * flags:
    * `-skipbuild`: do not try to build wasm
    * `-release`: build/run release build
