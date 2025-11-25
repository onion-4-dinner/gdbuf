.PHONY: build-linux
build-linux:
	cmake -B build -G Ninja -DCMAKE_CXX_COMPILER=clang++ -DCMAKE_SYSTEM_NAME=linux -DCMAKE_SYSTEM_PROCESSOR=x86_64 -DGODOTCPP_TARGET=template_debug -DGODOTCPP_PRECISION=single -DCMAKE_TOOLCHAIN_FILE=cmake/x86_64-linux.toolchain.cmake
	cmake --build build
	cp build/compile_commands.json ./

.PHONY: build-windows
build-windows:
	cmake -B build -DCMAKE_SYSTEM_NAME=Windows -DCMAKE_SYSTEM_PROCESSOR=x86_64 -DGODOTCPP_TARGET=template_debug -DGODOTCPP_PRECISION=single -DCMAKE_TOOLCHAIN_FILE=cmake/x86_64-windows.toolchain.cmake
	cmake --build build

.PHONY: deps-update
deps-update:
	git submodule update --remote --recursive
