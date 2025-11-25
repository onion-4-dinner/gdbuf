# `gdbuf`
**Seamlessly use Google Protobuf messages as native Resources in Godot.**

`gdbuf` compiles your `.proto` files into a high-performance GDExtension (C++), automatically bridging the gap between your network/data layer and the Godot Engine.

---

## Why `gdbuf`?

- **Native Resources:** Your Protobuf messages become first-class Godot `Resource` objects. Create, edit, and save them (`.tres`) directly in the Godot Editor.
- **Inspector Integration:** View and modify message fields in the Inspector with full type support.
- **Documentation:** Comments in your `.proto` files are automatically converted into Godot Editor tooltips and documentation.
- **Performance:** Built on GDExtension and C++ for maximum speed.
- **Zero Boilerplate:** No manual C++ coding required. Just run the tool.

## Installation

### Prerequisites
- **Go** (1.21+)
- **protoc** (The Protocol Buffers compiler) installed and in your system PATH.
- **C++ Compiler** (gcc/clang/msvc) & **CMake** (for building the extension).

### Build `gdbuf`
```bash
git clone --recursive https://github.com/LJ-Software/gdbuf.git
cd gdbuf
make build
# The binary will be in bin/gdbuf
```

## Usage

Run `gdbuf` pointing to your Protobuf definitions directory. It will handle parsing, code generation, and compiling the GDExtension library for you.

```bash
./bin/gdbuf \
  --proto ./path/to/your/protos \
  --include ./path/to/your/protos/public \
  --include ./path/to/your/protos/private \
  --out ./path/to/godot_project/addons/my_proto_extension \
  --name MyProtoLib
```

### Arguments
- `--proto`: Path to the directory containing your `.proto` files (Required).
- `--include`: Additional directories to include for resolving imports. Can be specified multiple times.
- `--out`: Directory where the compiled GDExtension (library + `.gdextension` file) will be placed (Default: `./out`).
- `--genout`: Directory where the intermediate C++ source code will be generated (Default: `.`).
- `--generate-only`: Only generate the C++ source code, skipping the GDExtension compilation step (Default: `false`).
- `--name`: Name of the GDExtension library (Default: `gdbufgen`).

## In Godot

Once the extension is generated and placed in your project:

1. **Restart Godot** to load the new extension.
2. **Use in GDScript:**
   ```gdscript
   # Create a message
   var msg = MyMessage.new()
   msg.health = 100
   msg.name = "Player 1"

   # Serialize
   var bytes = msg.to_byte_array()
   
   # Save to disk
   ResourceSaver.save(msg, "res://player_data.tres")
   ```
3. **Use in Editor:** Right-click in FileSystem -> New Resource -> Search for your message name.

## Documentation

- **[Features & Supported Types](docs/FEATURES.md)**: Detailed list of supported Protobuf features and how they map to Godot.
- **[GDScript API Reference](docs/API.md)**: Detailed guide on using the generated GDScript API.
- **[Developer Guide](docs/DEV.md)**: Architecture overview and guide for contributors.
- **[Agent Guidelines](docs/AGENTS.md)**: Instructions for AI agents working on this codebase.
- **[Testing](test/README.md)**: How to run the test suite.

