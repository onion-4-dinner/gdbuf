# Toolchain file for cross-compiling to 32-bit Windows using MinGW-w64

# Set the target system name
set(CMAKE_SYSTEM_NAME Windows)

# Set the target architecture
set(CMAKE_SYSTEM_PROCESSOR x86)
set(THREADS_PREFER_PTHREAD_FLAG TRUE)

# Specify the cross-compilers
# On Ubuntu runners, the mingw-w64 package provides these executables.
set(CMAKE_C_COMPILER i686-w64-mingw32-gcc)
set(CMAKE_CXX_COMPILER i686-w64-mingw32-g++)
set(CMAKE_RC_COMPILER i686-w64-mingw32-windres)

# Include the vcpkg toolchain file
if(DEFINED ENV{WORKSPACE})
  include("$ENV{WORKSPACE}/vcpkg/scripts/buildsystems/vcpkg.cmake")
endif()

