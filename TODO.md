# Project Refactoring & Improvement Tasks

## 1. Refactoring & Code Structure

### **Refactor Type Resolution Logic**
- **File:** `internal/codegen/codegen.go`
- **Issue:** The `extractProtoData` function is excessively long (~160 lines) and contains deeply nested logic for parsing Protobuf descriptors and determining corresponding Godot types. This violates the Single Responsibility Principle.
- **Task:**
  1. Create a new function `resolveGodotType(field *descriptorpb.FieldDescriptorProto, fileToMsgs map[string][]string, fileToEnum map[string][]string) (godotType string, isCustom bool, err error)`.
  2. Move the entire `switch fieldType` block (including WKT detection for Timestamp, Struct, etc.) from `codegen.go` into this new function.
  3. Place this function in `internal/codegen/godot.go` to centralize type mapping logic.

### **Centralize Godot Type Mapping**
- **File:** `internal/codegen/godot.go`, `internal/codegen/codegen.go`
- **Issue:** Primitive type mappings exist in `godot.go`, while complex Well-Known Type (WKT) mappings exist in `codegen.go`.
- **Task:** Ensure all "Proto -> Godot" type mapping logic resides in `godot.go`.

### **Remove Dead Code**
- **Files:** `internal/codegen/codegen.go`, `internal/codegen/godot.go`, `test/proto/gdbuf_test.proto`
- **Issue:** Several files contain commented-out code that clutter the codebase.
- **Task:**
  - Remove commented-out `if` block in `codegen.go` (lines 269-275).
  - Remove commented-out map keys in `godot.go` (lines 25-26).
  - Remove commented-out message definitions in `gdbuf_test.proto`.

## 2. Robustness & Error Handling

### **Improve Protoc Version Parsing**
- **File:** `internal/protoc/protoc.go`
- **Issue:** `getProtocExecutableVersion` manually splits the version string (e.g., "libprotoc 29.5"), which is fragile and prone to runtime panics if the output format changes slightly.
- **Task:** Replace string splitting with a Regular Expression (e.g., `libprotoc (\d+\.\d+)`) to safely extract the version number.

### **Safe Temporary Directories**
- **File:** `internal/protoc/protoc.go`
- **Issue:** `CompileCpp` creates a fixed directory `gdbuf-build` inside the system temp directory. This causes collisions if multiple instances run concurrently.
- **Task:** Use `os.MkdirTemp` to generate a unique build directory for every execution.

## 3. Cross-Platform Support

### **Support Non-Linux Builds**
- **File:** `internal/gdextension/gdextension.go`
- **Issue:** The build command is hardcoded to `exec.Command("make", "build-linux")`.
- **Task:**
  - Detect the host OS using `runtime.GOOS`.
  - Dynamically select the target (e.g., `build-linux`, `build-macos`, `build-windows`) based on the OS.
  - Alternatively, add a CLI flag to `main.go` to specify the target platform.

## 4. Testing

### **Add Unit Tests for Type Resolution**
- **Location:** `internal/codegen/godot_test.go` (New File)
- **Issue:** The project relies entirely on integration tests (`make test-full`). Fast feedback loops are missing for core logic.
- **Task:**
  - Create unit tests for the (newly refactored) `resolveGodotType` function.
  - Verify mappings for:
    - Primitives (int32 -> int32_t)
    - WKTs (Timestamp -> int64_t, Struct -> godot::Dictionary)
    - Enums (MyEnum -> int32_t)
