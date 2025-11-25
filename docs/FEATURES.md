# gdbuf Features

`gdbuf` turns your Google Protobuf definitions into native Godot C++ classes, allowing you to use them seamlessly in GDScript and the Godot Editor.

## Core Features

### 1. Native Resources
Every `message` in your `.proto` file is compiled into a native **Godot Resource** class.
- **Inheritance:** All messages inherit from `Resource`.
- **Usage:** You can create them using `.new()`, save them as `.tres` files, and view them in the Inspector.
- **Memory Management:** Godot handles memory automatically (Reference Counting).

### 2. Inspector Integration
Fields in your messages become **Properties** in Godot.
- **Editor Support:** View and edit message fields directly in the Inspector.
- **Tweening:** Use standard `tween_property` calls on your messages.
- **Access:** Access fields using dot notation: `msg.my_field = 10`.

### 3. Type Mapping
Protobuf types are mapped to their most natural Godot equivalents:

| Protobuf Type | Godot Type | Note |
| :--- | :--- | :--- |
| `int32`, `int64` | `int` | |
| `float`, `double` | `float` | |
| `string` | `String` | |
| `bytes` | `PackedByteArray` | |
| `repeated` field | `Array` | Typed array (e.g. `Array[int]`) where possible |
| `map` | `Dictionary` | |
| **Enums** | `int` | Constants are registered in the class |
| **Oneof** | *various* | `get_..._case()` helpers available |

#### Google Well-Known Types (WKT)
Common Google types are automatically converted to native Godot types for ease of use:
- **Timestamp** → `int` (Unix timestamp in milliseconds)
- **Duration** → `float` (Seconds)
- **Struct** → `Dictionary`
- **Value** → `Variant`
- **ListValue** → `Array`

### 4. Editor Documentation
Comments in your `.proto` files are converted into **Godot Editor Documentation**.
- **Tooltips:** Hover over a property in the Inspector or use code completion in the script editor to see your comments.
- **Formatting:** Standard protobuf comments (`// ...`) are captured and formatted.

### 5. Serialization
Classes include helper methods for binary serialization compatible with standard Protobuf libraries.
- `to_byte_array() -> PackedByteArray`
- `from_byte_array(data: PackedByteArray)`

### 6. Debugging
Messages implement `_to_string()`, allowing you to print them in GDScript for a human-readable text representation (using Protobuf's text format).
```gdscript
print(my_message)
# Output:
# name: "Hero"
# health: 100
```

### 7. Oneof Support
Oneof fields are supported with automatic state management.
- **Automatic Clearing:** Setting a field in a `oneof` group automatically clears the others.
- **Case Helpers:** Use `get_<oneof_name>_case()` to check which field is set.

```gdscript
# In your proto:
# oneof payload {
#   string text = 1;
#   int32 number = 2;
# }

var msg = MyMessage.new()
msg.text = "Hello"
print(msg.get_payload_case()) # Prints enum value for 'text' (1)

msg.number = 42 
# msg.text is now automatically cleared!
print(msg.text) # Prints "" (default string)
```

## Example

**Input (`player.proto`):**
```protobuf
// A player in the game world.
message Player {
  // The name displayed above the head.
  string name = 1;
  int32 health = 2;
}
```

**Usage (GDScript):**
```gdscript
var p = Player.new()
p.name = "Hero"
p.health = 100

# Save to disk
ResourceSaver.save(p, "res://hero.tres")

# Serialize for network
var data = p.to_byte_array()
```
