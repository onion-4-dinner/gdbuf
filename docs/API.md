# GDScript API Reference

This guide documents the GDScript API available for your generated Protobuf messages.

## Overview

Every `message` defined in your `.proto` files is compiled into a native C++ class that inherits from `godot::Resource`. This means you can use them just like any other Resource in Godot (e.g., `Texture`, `Material`).

## Base Methods

All generated message classes include the following methods:

### `to_byte_array() -> PackedByteArray`
Serializes the message content into a standard Protobuf binary format.
- **Returns:** A `PackedByteArray` containing the serialized data.
- **Usage:**
  ```gdscript
  var bytes = my_msg.to_byte_array()
  ```

### `from_byte_array(bytes: PackedByteArray) -> Error`
Deserializes data from a `PackedByteArray` into the current message object.
- **Parameters:**
  - `bytes`: The binary data to parse (must be valid Protobuf wire format).
- **Returns:** `OK` (0) if successful, `ERR_PARSE_ERROR` if parsing failed.
- **Usage:**
  ```gdscript
  var err = my_msg.from_byte_array(data)
  if err != OK:
      printerr("Failed to parse message")
  ```

### `get_proto_file_name() -> String`
Returns the name of the source `.proto` file this message was generated from (without the extension).
- **Usage:** `print(my_msg.get_proto_file_name())`

### `_to_string() -> String`
Returns a human-readable string representation of the message (debug string).
- **Usage:** `print(my_msg)`

## Fields (Properties)

Message fields are exposed as standard Godot properties. You can read and write them directly.

| Protobuf Field | GDScript Property | Type | Note |
| :--- | :--- | :--- | :--- |
| `int32 id = 1;` | `msg.id` | `int` | |
| `string name = 2;` | `msg.name` | `String` | |
| `bool active = 3;` | `msg.active` | `bool` | |
| `repeated int32 scores = 4;` | `msg.scores` | `Array[int]` | |
| `map<string, int32> items = 5;` | `msg.items` | `Dictionary` | |
| `MyNestedMsg nested = 6;` | `msg.nested` | `MyNestedMsg` | Inherits `Resource` |

### Nested Messages
Nested message fields use `Ref<Resource>` semantics.
- If a field is unset, it might be `null`.
- You can assign a new instance: `msg.nested = MyNestedMsg.new()`.

## Enums

Protobuf `enum` definitions are exposed as constants within the class or namespace.

```protobuf
enum Status {
  UNKNOWN = 0;
  ACTIVE = 1;
}
```

**GDScript:**
```gdscript
msg.status = MyMessage.Status.ACTIVE
# or if defined at file level
msg.status = MyProtoFile.Status.ACTIVE
```

## Oneof Fields

`oneof` fields allow only one of the fields in the group to be set at a time.

```protobuf
oneof payload {
  string text = 1;
  int32 number = 2;
}
```

### Methods
- **`get_<oneof_name>_case()`**: Returns the enum value corresponding to the currently set field tag number. Returns `0` (NOT_SET) if none are set.

### Usage
```gdscript
# Set text
msg.text = "Hello"
print(msg.get_payload_case()) # Prints 1 (tag number of 'text')

# Set number (automatically clears text)
msg.number = 42
print(msg.get_payload_case()) # Prints 2 (tag number of 'number')
print(msg.text) # Prints ""
```

## Well-Known Types (WKT)

| Protobuf WKT | GDScript Type | Details |
| :--- | :--- | :--- |
| `google.protobuf.Timestamp` | `int` | Unix timestamp in milliseconds. |
| `google.protobuf.Duration` | `float` | Duration in seconds. |
| `google.protobuf.Struct` | `Dictionary` | JSON-like object. |
| `google.protobuf.Value` | `Variant` | Any simple value. |
| `google.protobuf.ListValue` | `Array` | List of values. |
| `google.protobuf.Empty` | `null` | Effectively unused. |
