# `gdbuf` Testing

---

## Quick Start

0. Prerequisites
You will need the following software installed:
- `Go` >=1.25
- `protoc`
- gnu `make`

1. Run the full test
```
make test-full
```

---

## Testing Strategy

In this directory there is a `proto` directory that contains protobuf message definitions spanning much of the feature set. To test `gdbuf` first we build a protobuf description file with `protoc` using this test definition. Then, we run the test description file through `gdbuf` to get the generated Golang gdextension C++ source code. Finally, we can try to compile the source code to ensure that we (at least) have some compile-time gauruntee that we made something valid.

### Improvements

Some improvements to testing that could be added:

- Some type of runtime test suite to confirm the compiled gdextension works as intended
