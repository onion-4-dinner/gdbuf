# Toolchain for cross-compiling to AArch64 Linux
set(CMAKE_SYSTEM_NAME Linux)
set(CMAKE_SYSTEM_PROCESSOR aarch64)

# Specify the cross-compilers
set(CMAKE_C_COMPILER aarch64-linux-gnu-gcc)
set(CMAKE_CXX_COMPILER aarch64-linux-gnu-g++)

# Include the vcpkg toolchain file
if(DEFINED ENV{WORKSPACE})
  include("$ENV{WORKSPACE}/vcpkg/scripts/buildsystems/vcpkg.cmake")
endif()
