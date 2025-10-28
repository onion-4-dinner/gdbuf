# Toolchain for building 32-bit Linux on a 64-bit host
set(CMAKE_SYSTEM_NAME Emscripten)

set(VCPKG_CHAINLOAD_TOOLCHAIN_FILE "$ENV{EMSDK}/upstream/emscripten/cmake/Modules/Platform/Emscripten.cmake")

# Include the vcpkg toolchain file
if(DEFINED ENV{WORKSPACE})
  include("$ENV{WORKSPACE}/vcpkg/scripts/buildsystems/vcpkg.cmake")
endif()
