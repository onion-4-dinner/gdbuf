# This toolchain file delegates to the Android NDK's toolchain file
cmake_minimum_required(VERSION 3.16)

if(NOT DEFINED ENV{ANDROID_NDK_HOME})
  message(FATAL_ERROR "The ANDROID_NDK_HOME environment variable must be set.")
endif()

set(ANDROID_NDK_HOME "$ENV{ANDROID_NDK_HOME}")

# Set variables expected by the NDK toolchain file
if(NOT DEFINED ANDROID_PLATFORM)
  set(ANDROID_PLATFORM "android-${ANDROID_PLATFORM_LEVEL}")
endif()

if(NOT DEFINED ANDROID_ABI)
  set(ANDROID_ABI "arm64-v8a")
endif()

set(ANDROID_STL "c++_static")

# Include the NDK toolchain file
include("${ANDROID_NDK_HOME}/build/cmake/android.toolchain.cmake")
