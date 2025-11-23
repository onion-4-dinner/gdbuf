# Development Guidelines

## Build & Test
- **Build:** `make build` compiles the CLI tool to `bin/gdbuf`.
- **Unit Test:** `go test ./...` runs Go unit tests (specifically type resolution logic).
- **Integration Test:** `make test-full` runs the end-to-end suite: compiles protos -> runs gdbuf -> compiles GDExtension.
  - *Prerequisites:* `go`, `protoc` (v25.1+), `make`, `cmake`, `ninja`, C++ compiler (gcc/clang).

## Code Style & Conventions
- **Go:** Idiomatic Go. Use `go fmt`. Group imports: stdlib, 3rd-party, internal.
- **C++:** Follow Godot GDExtension conventions. Generated code is in `internal/codegen/templates`.
- **Logging:** Use `log/slog` exclusively.
- **Error Handling:** Wrap errors with context: `fmt.Errorf("context: %w", err)`.
- **Protobuf:** Use `google.golang.org/protobuf`. Map WKTs (Timestamp, Struct) to native Godot types (int64, Dictionary) in `godot.go`.

## Project Structure
- `internal/codegen`: Core logic. `godot.go` handles type mapping. `codegen.go` drives generation.
- `internal/gdextension`: Builds the C++ library. `gdextension.go` selects OS-specific build targets.
- `internal/protoc`: Wrapper for `protoc` CLI operations. Uses `os.MkdirTemp` for safety.
