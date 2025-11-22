# Development Guidelines

## Build & Test
- **Build:** `make build` creates the binary in `bin/gdbuf`.
- **Test:** `make test-full` runs the complete integration suite.
  - *Process:* Compiles `test/proto/*.proto`, runs the generator, and compiles the resulting GDExtension.
  - *Note:* No Go unit tests exist; verification relies on successful end-to-end generation and C++ compilation.
  - *Prerequisites:* `go`, `protoc`, `make`, and a C++ compiler.

## Code Style & Conventions
- **Go:** Follow standard `gofmt` and idiomatic Go conventions.
- **Imports:** Grouped: Standard library, then 3rd party, then project packages (`github.com/LJ-Software/gdbuf/...`).
- **Logging:** Use `log/slog` for all application logging.
- **Error Handling:** Always wrap errors using `fmt.Errorf("...: %w", err)`.
- **Templates:** C++ code generation uses `text/template` + `sprig` in `internal/codegen/templates/`.
- **C++:** Generated C++ code follows Godot GDExtension patterns.
- **Project Structure:**
  - `internal/codegen`: Logic for generating C++ files.
  - `internal/gdextension`: Builds the final shared library.
  - `internal/protoc`: Wrapper around `protoc` operations.
