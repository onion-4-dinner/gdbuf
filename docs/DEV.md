# Developer Guide

`gdbuf` is a tool designed to bridge the gap between **Google Protocol Buffers** and the **Godot Engine**. Its primary goal is to automate the generation of high-performance, native C++ GDExtension code from `.proto` files, treating Protobuf messages as first-class Godot Resources.

## Architecture Overview

The program operates in a linear pipeline:
1.  **Input Parsing**: Accepts a directory of `.proto` files and optional include directories for import resolution.
2.  **Protoc Compilation**: Uses `protoc` to generate standard C++ headers/sources and a binary descriptor set (`.desc.binpb`).
3.  **Code Generation**: Parses the descriptor set to understand the message structure, then executes Go `text/template` templates to generate Godot-specific C++ wrappers.
4.  **GDExtension Compilation**: Orchestrates a CMake build to compile the generated wrappers + the embedded **Nanopb** library into a shared library. It uses a persistent build cache (in `~/.cache/gdbuf` or local `.gdbuf_cache`) to enable incremental builds.

## Codebase Structure

### `cmd` / Root
-   **`main.go`**: The entry point. It parses command-line flags, sets up logging, and orchestrates the three main internal packages (`protoc`, `codegen`, `gdextension`).

### `internal/protoc`
-   **Responsibility**: Wraps the `protoc` command-line tool.
-   **Key Functions**:
    -   `BuildDescriptorSet`: Runs `protoc --descriptor_set_out` to get a machine-readable definition of the proto files.
    -   `CompileNanopb`: Runs `protoc` with the `protoc-gen-nanopb` plugin to generate Nanopb C headers/sources (`.pb.h`, `.pb.c`).
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
-   **`buildenv/`**: Contains the "build environment" (CMake lists, godot-cpp submodule, nanopb submodule) embedded into the Go binary.
-   **`gdextension.go`**: Manages the build process. It sets up a persistent build cache to avoid recompiling `godot-cpp` on every run. It copies generated sources into this cache and invokes `make`.

## Key Concepts

### Type Mapping (`godot.go`)
We map standard Proto types to Godot Variants:
-   `string` -> `godot::String`
-   `map<K,V>` -> `godot::Dictionary` (with custom serialization logic in templates)
-   `oneof` -> Properties for each option, handled via explicit getters/setters.

### Resource Wrapper (`resource.h.tmpl`)
Every Protobuf message is wrapped in a class inheriting from `godot::Resource`.
-   **Properties**: Fields are registered with `ADD_PROPERTY`, making them visible in the Godot Inspector.
-   **Serialization**: We generate `to_byte_array()` and `from_byte_array()` methods that internally use **Nanopb** functions (`pb_encode`, `pb_decode`) to serialize to/from the standard Protobuf binary format.
-   **Memory Management**: Custom message fields (nested messages) are stored using `godot::Ref<T>`. This ensures that the Godot RefCounted system correctly manages the lifecycle of nested resources, preventing dangling pointers and memory leaks.

## Development Workflow

1.  **Make Changes**: Edit Go code or Templates.
2.  **Rebuild Tool**: Run `make build` to update `bin/gdbuf`.
3.  **Test**: Run `make test-full`. This:
    -   Cleans previous outputs.
    -   Builds the GDExtension for **all supported platforms** (Linux, Windows, Web, Android) using `test-all-platforms`.
    -   Installs the artifacts into a test Godot project.
    -   Runs a headless Godot instance (on the host) to execute GDScript integration tests (`test-godot`).
    -   *Note*: This requires all cross-compilers to be present. For a faster, host-only test (Linux), use `make test-linux`.
    -   **Platform Specific Tests**:
        -   `make test-linux`: Builds and tests for Linux.
        -   `make test-web`: Builds for Web (wasm32).
        -   `make test-windows`: Builds for Windows.
        -   `make test-android`: Builds for Android.

## Future Improvements
-   **Nested Enums**: Currently top-level enums work best. Nested enums map to `int` but don't generate C++ enum definitions in the wrapper namespace.
-   **Platform Support**: The Go code supports detecting platforms.
    -   **Linux/Windows/macOS**: Supported via CMake toolchains.
    -   **Android**: Fully supported. `gdbuf` automatically downloads and manages the Android NDK if `ANDROID_NDK_HOME` is not set.
    -   **Web**: Fully supported. `gdbuf` automatically downloads and manages the Emscripten SDK (emsdk) if `EMSDK` is not set. Requires Python 3 to be installed on the host system.
-   **Map Values (Structs)**: Support for `map<Key, Message>` with Nanopb needs verification for complex nested types.
