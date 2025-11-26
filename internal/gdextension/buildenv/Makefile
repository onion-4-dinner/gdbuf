.PHONY: build-linux
build-linux:
	cmake -B build/linux -G Ninja -DCMAKE_CXX_COMPILER=clang++ -DCMAKE_SYSTEM_NAME=linux -DCMAKE_SYSTEM_PROCESSOR=x86_64 -DGODOTCPP_TARGET=template_debug -DGODOTCPP_PRECISION=single -DCMAKE_TOOLCHAIN_FILE=cmake/x86_64-linux.toolchain.cmake
	cmake --build build/linux
	cp build/linux/compile_commands.json ./

.PHONY: build-windows
build-windows:
	cmake -B build/windows -DCMAKE_SYSTEM_NAME=Windows -DCMAKE_SYSTEM_PROCESSOR=x86_64 -DGODOTCPP_TARGET=template_debug -DGODOTCPP_PRECISION=single -DCMAKE_TOOLCHAIN_FILE=cmake/x86_64-windows.toolchain.cmake
	cmake --build build/windows

.PHONY: build-web
build-web:
	cmake -B build/web -DCMAKE_TOOLCHAIN_FILE=cmake/web-wasm32.toolchain.cmake -DGODOTCPP_TARGET=template_debug -DGODOTCPP_PRECISION=single -DGODOTCPP_THREADS=OFF
	cmake --build build/web

.PHONY: build-android
build-android:
	cmake -B build/android -DCMAKE_TOOLCHAIN_FILE=cmake/android.toolchain.cmake -DGODOTCPP_TARGET=template_debug -DGODOTCPP_PRECISION=single -DANDROID_ABI=arm64-v8a -DANDROID_PLATFORM_LEVEL=24
	cmake --build build/android

.PHONY: deps-update
deps-update:
	git submodule update --remote --recursive
