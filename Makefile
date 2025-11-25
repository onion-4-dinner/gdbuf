.PHONY: build
build:
	mkdir -p bin
	go build -o bin/gdbuf .

.PHONY: test-build-extension
test-build-extension: test-clean test-build
	mkdir -p test/out
	mkdir -p test/genout
	go run main.go --proto test/proto --include . --genout test/genout --out test/out

.PHONY: test-clean
test-clean:
	rm -r test/out
	rm -r test/genout

.PHONY: test-build
test-build: test/test.desc.binpb

test/test.desc.binpb:
	protoc --descriptor_set_out=test/test.desc.binpb test/proto/gdbuf_test.proto

.PHONY: test-godot
test-godot: test-build-extension
	mkdir -p test/godot_project/addons/gdbufgen
	cp -r test/out/* test/godot_project/addons/gdbufgen/
	# Run editor briefly to import
	godot --headless --path test/godot_project --editor --quit
	godot --headless --verbose --path test/godot_project -s test_runner.gd

.PHONY: test-full
test-full: test-build test-godot
