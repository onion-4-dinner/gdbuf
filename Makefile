.PHONY: build
build:
	mkdir -p bin
	go build -o bin/gdbuf .

.PHONY: test-full
test-full: test-build
	mkdir -p test/out
	go run main.go --proto test/proto --out test/out

.PHONY: test-build
test-build: test/test.desc.binpb

test/test.desc.binpb:
	protoc --descriptor_set_out=test/test.desc.binpb test/proto/gdbuf_test.proto
