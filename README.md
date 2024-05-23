# eh?

Just trying to get to the bottom of a crash in a wasm module built from rust, when run via wazero.

## quick start
```sh
go run ./cmd -release
```
 - wasm is compiled, and additional memory is alloc'd for wasm
 - wasm func exits with `unreachable` error and stack trace

```sh
go run ./cmd -release -interp
```
 - interpreted rather than compiled
 - wasm func is successful
```sh
go run ./cmd -release -nomem
```
 - no extra memory is allocated
 - wasm func is successful

## versions
```

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
