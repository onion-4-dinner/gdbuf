# Toolchain file for cross-compiling to 64-bit Windows using MinGW-w64

# Set the target system name
set(CMAKE_SYSTEM_NAME Windows)

# Set the target architecture
set(CMAKE_SYSTEM_PROCESSOR x86_64)

# Specify the cross-compilers
# On Ubuntu runners, the mingw-w64 package provides these executables.
set(THREADS_PREFER_PTHREAD_FLAG TRUE)
set(CMAKE_C_COMPILER x86_64-w64-mingw32-gcc CACHE STRING "C compiler")
set(CMAKE_CXX_COMPILER x86_64-w64-mingw32-g++ CACHE STRING "C++ compiler")
set(CMAKE_RC_COMPILER x86_64-w64-mingw32-windres CACHE STRING "Resource compiler")
