# Development Guidelines

## Build & Test
- **Build:** `make build` compiles the CLI tool to `bin/gdbuf`.
- **Unit Test:** `go test ./...` runs all Go unit tests.
  - Single test: `go test -run TestName ./...`
- **Integration Test:** `make test-full` runs the end-to-end suite (compile protos -> run gdbuf -> compile GDExtension).
  - Requires: `go`, `protoc` (v25.1+), `make`, `cmake`, `ninja`, C++ compiler.

## Code Style & Conventions
- **Go:** Standard formatting (`go fmt`). Imports grouped: stdlib, 3rd-party, internal (`github.com/LJ-Software/gdbuf/...`).
- **C++:** Follow Godot GDExtension conventions. Templates in `internal/codegen/templates`.
- **Logging:** Use `log/slog` exclusively.
- **Errors:** Wrap with context: `fmt.Errorf("context: %w", err)`.
- **Protobuf:** Use `google.golang.org/protobuf`. Map WKTs to native Godot types in `godot.go`.

## Project Structure
- `internal/codegen`: Logic for generating C++ code. `codegen.go` is the entry point.
- `internal/gdextension`: Builds the C++ library.
- `internal/protoc`: Wrapper for `protoc` CLI.
- `internal/codegen/templates`: C++ templates for generated files.
