# Developer Guide

`gdbuf` is a tool designed to bridge the gap between **Google Protocol Buffers** and the **Godot Engine**. Its primary goal is to automate the generation of high-performance, native C++ GDExtension code from `.proto` files, treating Protobuf messages as first-class Godot Resources.

## Architecture Overview

The program operates in a linear pipeline:
1.  **Input Parsing**: Accepts a directory of `.proto` files.
2.  **Protoc Compilation**: Uses `protoc` to generate standard C++ headers/sources and a binary descriptor set (`.desc.binpb`).
3.  **Code Generation**: Parses the descriptor set to understand the message structure, then executes Go `text/template` templates to generate Godot-specific C++ wrappers.
4.  **GDExtension Compilation**: Orchestrates a CMake build to compile the generated wrappers + standard Protobuf C++ code into a shared library (`.so`, `.dll`, `.dylib`).

## Codebase Structure

### `cmd` / Root
-   **`main.go`**: The entry point. It parses command-line flags, sets up logging, and orchestrates the three main internal packages (`protoc`, `codegen`, `gdextension`).

### `internal/protoc`
-   **Responsibility**: Wraps the `protoc` command-line tool.
-   **Key Functions**:
    -   `BuildDescriptorSet`: Runs `protoc --descriptor_set_out` to get a machine-readable definition of the proto files.
    -   `CompileCpp`: Runs `protoc --cpp_out` to generate the standard C++ implementation of the messages.
-   **Notes**: Enforces strict include paths (relative to current directory) to avoid aliasing issues.

### `internal/codegen`
-   **Responsibility**: Logic for mapping Protobuf concepts to Godot concepts and generating C++ code.
-   **`codegen.go`**: The main driver. It extracts data from the `FileDescriptorProto`s, populates the `templateData` struct, and executes templates. It handles recursive message flattening (for nested types) and doc comment extraction.
-   **`godot.go`**: Contains the `resolveGodotType` logic. This is the "brain" that decides that `google.protobuf.Timestamp` becomes `int64_t`, or `repeated MyMsg` becomes `godot::Array`.
-   **`templates/`**: Contains the `.tmpl` files.
    -   `gdextension/`: CMake and config files.
    -   `src/gen_once/`: Files generated once per run (e.g., `register_types.cpp`).
    -   `src/gen_per_proto_file/`: Files generated for every proto file (e.g., `resource.h` which defines the wrapper classes).
    -   `doc/`: Templates for generating Godot XML documentation.

### `internal/gdextension`
-   **Responsibility**: Builds the final binary.
-   **`buildenv/`**: Contains the "build environment" (CMake lists, vcpkg config, godot-cpp submodule) embedded into the Go binary.
-   **`gdextension.go`**: Extracts the embedded build environment to a temporary directory, copies the generated code into it, and runs `cmake` / `make`.

## Key Concepts

### Type Mapping (`godot.go`)
We map standard Proto types to Godot Variants:
-   `string` -> `godot::String`
-   `map<K,V>` -> `godot::Dictionary` (with custom serialization logic in templates)
-   `oneof` -> Properties for each option, handled via explicit getters/setters.

### Resource Wrapper (`resource.h.tmpl`)
Every Protobuf message is wrapped in a class inheriting from `godot::Resource`.
-   **Properties**: Fields are registered with `ADD_PROPERTY`, making them visible in the Godot Inspector.
-   **Serialization**: We generate `to_byte_array()` and `from_byte_array()` methods that internally use the standard Protobuf `SerializeAsString()` and `ParseFromArray()`.

## Development Workflow

1.  **Make Changes**: Edit Go code or Templates.
2.  **Rebuild Tool**: Run `make build` to update `bin/gdbuf`.
3.  **Test**: Run `make test-full`. This:
    -   Cleans previous outputs.
    -   Runs `gdbuf` on `test/proto`.
    -   Compiles the resulting GDExtension.
    -   *Note*: This is an integration test. If it passes, the C++ code is valid.

## Future Improvements
-   **Nested Enums**: Currently top-level enums work best. Nested enums map to `int` but don't generate C++ enum definitions in the wrapper namespace.
-   **Platform Support**: The Go code supports detecting platforms, but the `Makefile` in `buildenv` primarily targets Linux (`x64-linux`). Windows/macOS targets need to be fully fleshed out in the embedded Makefile.
