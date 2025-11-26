.PHONY: build
build:
	mkdir -p bin
	go build -o bin/gdbuf .

.PHONY: test-clean
test-clean:
	rm -rf test/out
	rm -rf test/genout
	rm -rf test/out-linux
	rm -rf test/genout-linux
	rm -rf test/out-web
	rm -rf test/genout-web
	rm -rf test/out-android
	rm -rf test/genout-android
	rm -rf test/out-windows
	rm -rf test/genout-windows
	rm -rf test/out-all
	rm -rf test/genout-all
	rm -rf test/out-hyphen
	rm -rf test/genout-hyphen
	rm -rf test/godot_project/addons/gdbufgen

.PHONY: test-build
test-build: test/test.desc.binpb

test/test.desc.binpb:
	protoc --descriptor_set_out=test/test.desc.binpb test/proto/gdbuf_test.proto

.PHONY: test-build-linux
test-build-linux: test-clean test-build
	mkdir -p test/out-linux
	mkdir -p test/genout-linux
	go run main.go --proto test/proto --include . --genout test/genout-linux --out test/out-linux --platform linux

.PHONY: test-godot
test-godot: test-build-linux
	mkdir -p test/godot_project/addons/gdbufgen
	cp -r test/out-linux/* test/godot_project/addons/gdbufgen/
	# Run editor briefly to import
	godot --headless --path test/godot_project --editor --quit
	godot --headless --verbose --path test/godot_project -s test_runner.gd

.PHONY: test-linux
test-linux: test-build test-godot

.PHONY: test-full
test-full: test-all-platforms
	mkdir -p test/godot_project/addons/gdbufgen
	cp -r test/out-all/* test/godot_project/addons/gdbufgen/
	# Run editor briefly to import
	godot --headless --path test/godot_project --editor --quit
	godot --headless --verbose --path test/godot_project -s test_runner.gd

.PHONY: test-hyphen
test-hyphen: test-clean test-build
	mkdir -p test/out-hyphen
	mkdir -p test/genout-hyphen
	go run main.go --proto test/proto --include . --genout test/genout-hyphen --out test/out-hyphen --name "my-hyphenated-extension"

.PHONY: test-web
test-web: test-clean test-build
	mkdir -p test/out-web
	mkdir -p test/genout-web
	go run main.go --proto test/proto --include . --genout test/genout-web --out test/out-web --platform web

.PHONY: test-android
test-android: test-clean test-build
	mkdir -p test/out-android
	mkdir -p test/genout-android
	go run main.go --proto test/proto --include . --genout test/genout-android --out test/out-android --platform android

.PHONY: test-windows
test-windows: test-clean test-build
	mkdir -p test/out-windows
	mkdir -p test/genout-windows
	go run main.go --proto test/proto --include . --genout test/genout-windows --out test/out-windows --platform windows

.PHONY: test-all-platforms
test-all-platforms: test-clean test-build
	mkdir -p test/out-all
	mkdir -p test/genout-all
	go run main.go --proto test/proto --include . --genout test/genout-all --out test/out-all --platform all
