# `gdbuf`
## Use your Protobuf messages in Godot

---
Note that this software is early and development; don't expect everything to work.

## Quick Start

0. Prerequisites
You will need the following software installed:
- `Go` >=1.25
- `protoc`

1. Generate a protobuf description file from your protobuf definition files.
```
protoc --descriptor_set_out=my_proto.desc.binpb <PATH_TO_YOUR_PROTO_FILES>
```
Note: replace `my_proto` with your proto project name.

2. Clone this repository.
```
git clone https://github.com/LJ-Software/gdbuf.git
```


3. Bring your protobuf description file into the gdbuf directory.
```
mv my_proto.desc.binpb gdbuf && cd gdbuf
```

4. Run `gdbuf` on your protobuf description file to generate the c++ gdextension code
```
mkdir -p gen
go run main.go --proto-desc my_proto.desc.binpb --out gen
```

5. Compile the gdextension with your newly generated code
```
COMING SOON
```
---

## More Info

- [Testing `gdbuf`](test/README.md)
