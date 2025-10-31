.PHONY: build
build:
	mkdir -p bin
	go build -o bin/gdbuf .

.PHONY: test-full
test-full: test-clean test-build
	mkdir -p test/out
	go run main.go --proto test/proto --genout test/genout --out test/out

.PHONY: test-clean
test-clean:
	rm -r test/out

.PHONY: test-build
test-build: test/test.desc.binpb

test/test.desc.binpb:
	protoc --descriptor_set_out=test/test.desc.binpb test/proto/gdbuf_test.proto
