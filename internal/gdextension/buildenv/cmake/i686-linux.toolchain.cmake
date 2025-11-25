# Toolchain for building 32-bit Linux on a 64-bit host
set(CMAKE_SYSTEM_NAME Linux)
set(CMAKE_SYSTEM_PROCESSOR i686)

# Use the host compilers but force 32-bit output
set(CMAKE_C_COMPILER gcc)
set(CMAKE_CXX_COMPILER g++)
set(CMAKE_C_FLAGS "-m32" CACHE STRING "C compiler flags for 32-bit")
set(CMAKE_CXX_FLAGS "-m32" CACHE STRING "C++ compiler flags for 32-bit")
