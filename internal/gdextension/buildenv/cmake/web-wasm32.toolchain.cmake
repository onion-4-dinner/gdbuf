# Toolchain for building 32-bit Linux on a 64-bit host
set(CMAKE_SYSTEM_NAME Emscripten)

include("$ENV{EMSDK}/upstream/emscripten/cmake/Modules/Platform/Emscripten.cmake")
