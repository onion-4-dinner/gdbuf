.PHONY: build-linux build-windows build-web build-android deps-update

# Linux
build-linux: build-linux-debug-single build-linux-release-single

build-linux-debug-single:
	cmake -B build/linux/debug_single -G Ninja -DCMAKE_CXX_COMPILER=clang++ -DCMAKE_SYSTEM_NAME=linux -DCMAKE_SYSTEM_PROCESSOR=x86_64 -DGODOTCPP_TARGET=template_debug -DGODOTCPP_PRECISION=single -DCMAKE_TOOLCHAIN_FILE=cmake/x86_64-linux.toolchain.cmake
	cmake --build build/linux/debug_single
	mkdir -p build/linux/bin
	cp build/linux/debug_single/bin/* build/linux/bin/
	cp build/linux/debug_single/compile_commands.json ./

build-linux-release-single:
	cmake -B build/linux/release_single -G Ninja -DCMAKE_CXX_COMPILER=clang++ -DCMAKE_SYSTEM_NAME=linux -DCMAKE_SYSTEM_PROCESSOR=x86_64 -DGODOTCPP_TARGET=template_release -DGODOTCPP_PRECISION=single -DCMAKE_TOOLCHAIN_FILE=cmake/x86_64-linux.toolchain.cmake
	cmake --build build/linux/release_single
	mkdir -p build/linux/bin
	cp build/linux/release_single/bin/* build/linux/bin/

build-linux-debug-double:
	cmake -B build/linux/debug_double -G Ninja -DCMAKE_CXX_COMPILER=clang++ -DCMAKE_SYSTEM_NAME=linux -DCMAKE_SYSTEM_PROCESSOR=x86_64 -DGODOTCPP_TARGET=template_debug -DGODOTCPP_PRECISION=double -DCMAKE_TOOLCHAIN_FILE=cmake/x86_64-linux.toolchain.cmake
	cmake --build build/linux/debug_double
	mkdir -p build/linux/bin
	cp build/linux/debug_double/bin/* build/linux/bin/

build-linux-release-double:
	cmake -B build/linux/release_double -G Ninja -DCMAKE_CXX_COMPILER=clang++ -DCMAKE_SYSTEM_NAME=linux -DCMAKE_SYSTEM_PROCESSOR=x86_64 -DGODOTCPP_TARGET=template_release -DGODOTCPP_PRECISION=double -DCMAKE_TOOLCHAIN_FILE=cmake/x86_64-linux.toolchain.cmake
	cmake --build build/linux/release_double
	mkdir -p build/linux/bin
	cp build/linux/release_double/bin/* build/linux/bin/

# Windows
build-windows: build-windows-debug-single build-windows-release-single

build-windows-debug-single:
	cmake -B build/windows/debug_single -DCMAKE_SYSTEM_NAME=Windows -DCMAKE_SYSTEM_PROCESSOR=x86_64 -DGODOTCPP_TARGET=template_debug -DGODOTCPP_PRECISION=single -DCMAKE_TOOLCHAIN_FILE=cmake/x86_64-windows.toolchain.cmake
	cmake --build build/windows/debug_single
	mkdir -p build/windows/bin
	cp build/windows/debug_single/bin/* build/windows/bin/

build-windows-release-single:
	cmake -B build/windows/release_single -DCMAKE_SYSTEM_NAME=Windows -DCMAKE_SYSTEM_PROCESSOR=x86_64 -DGODOTCPP_TARGET=template_release -DGODOTCPP_PRECISION=single -DCMAKE_TOOLCHAIN_FILE=cmake/x86_64-windows.toolchain.cmake
	cmake --build build/windows/release_single
	mkdir -p build/windows/bin
	cp build/windows/release_single/bin/* build/windows/bin/

build-windows-debug-double:
	cmake -B build/windows/debug_double -DCMAKE_SYSTEM_NAME=Windows -DCMAKE_SYSTEM_PROCESSOR=x86_64 -DGODOTCPP_TARGET=template_debug -DGODOTCPP_PRECISION=double -DCMAKE_TOOLCHAIN_FILE=cmake/x86_64-windows.toolchain.cmake
	cmake --build build/windows/debug_double
	mkdir -p build/windows/bin
	cp build/windows/debug_double/bin/* build/windows/bin/

build-windows-release-double:
	cmake -B build/windows/release_double -DCMAKE_SYSTEM_NAME=Windows -DCMAKE_SYSTEM_PROCESSOR=x86_64 -DGODOTCPP_TARGET=template_release -DGODOTCPP_PRECISION=double -DCMAKE_TOOLCHAIN_FILE=cmake/x86_64-windows.toolchain.cmake
	cmake --build build/windows/release_double
	mkdir -p build/windows/bin
	cp build/windows/release_double/bin/* build/windows/bin/

# Web
build-web: build-web-debug-single build-web-release-single

build-web-debug-single:
	cmake -B build/web/debug_single -DCMAKE_TOOLCHAIN_FILE=cmake/web-wasm32.toolchain.cmake -DGODOTCPP_TARGET=template_debug -DGODOTCPP_PRECISION=single -DGODOTCPP_THREADS=OFF
	cmake --build build/web/debug_single
	mkdir -p build/web/bin
	cp build/web/debug_single/bin/* build/web/bin/

build-web-release-single:
	cmake -B build/web/release_single -DCMAKE_TOOLCHAIN_FILE=cmake/web-wasm32.toolchain.cmake -DGODOTCPP_TARGET=template_release -DGODOTCPP_PRECISION=single -DGODOTCPP_THREADS=OFF
	cmake --build build/web/release_single
	mkdir -p build/web/bin
	cp build/web/release_single/bin/* build/web/bin/

build-web-debug-double:
	cmake -B build/web/debug_double -DCMAKE_TOOLCHAIN_FILE=cmake/web-wasm32.toolchain.cmake -DGODOTCPP_TARGET=template_debug -DGODOTCPP_PRECISION=double -DGODOTCPP_THREADS=OFF
	cmake --build build/web/debug_double
	mkdir -p build/web/bin
	cp build/web/debug_double/bin/* build/web/bin/

build-web-release-double:
	cmake -B build/web/release_double -DCMAKE_TOOLCHAIN_FILE=cmake/web-wasm32.toolchain.cmake -DGODOTCPP_TARGET=template_release -DGODOTCPP_PRECISION=double -DGODOTCPP_THREADS=OFF
	cmake --build build/web/release_double
	mkdir -p build/web/bin
	cp build/web/release_double/bin/* build/web/bin/

# Android
build-android: build-android-debug-single build-android-release-single

build-android-debug-single:
	cmake -B build/android/debug_single -DCMAKE_TOOLCHAIN_FILE=cmake/android.toolchain.cmake -DGODOTCPP_TARGET=template_debug -DGODOTCPP_PRECISION=single -DANDROID_ABI=arm64-v8a -DANDROID_PLATFORM_LEVEL=24
	cmake --build build/android/debug_single
	mkdir -p build/android/bin
	cp build/android/debug_single/bin/* build/android/bin/

build-android-release-single:
	cmake -B build/android/release_single -DCMAKE_TOOLCHAIN_FILE=cmake/android.toolchain.cmake -DGODOTCPP_TARGET=template_release -DGODOTCPP_PRECISION=single -DANDROID_ABI=arm64-v8a -DANDROID_PLATFORM_LEVEL=24
	cmake --build build/android/release_single
	mkdir -p build/android/bin
	cp build/android/release_single/bin/* build/android/bin/

build-android-debug-double:
	cmake -B build/android/debug_double -DCMAKE_TOOLCHAIN_FILE=cmake/android.toolchain.cmake -DGODOTCPP_TARGET=template_debug -DGODOTCPP_PRECISION=double -DANDROID_ABI=arm64-v8a -DANDROID_PLATFORM_LEVEL=24
	cmake --build build/android/debug_double
	mkdir -p build/android/bin
	cp build/android/debug_double/bin/* build/android/bin/

build-android-release-double:
	cmake -B build/android/release_double -DCMAKE_TOOLCHAIN_FILE=cmake/android.toolchain.cmake -DGODOTCPP_TARGET=template_release -DGODOTCPP_PRECISION=double -DANDROID_ABI=arm64-v8a -DANDROID_PLATFORM_LEVEL=24
	cmake --build build/android/release_double
	mkdir -p build/android/bin
	cp build/android/release_double/bin/* build/android/bin/

deps-update:
	git submodule update --remote --recursive
