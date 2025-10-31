# This toolchain file configures the build for Android using the NDK
cmake_minimum_required(VERSION 3.16)

# Tell CMake we are cross-compiling for Android
set(CMAKE_SYSTEM_NAME Android)
set(CMAKE_SYSTEM_VERSION ${ANDROID_PLATFORM_LEVEL}) # Will be passed in, e.g., 24

# Let CMake know which ABI you are targeting
set(CMAKE_ANDROID_ARCH_ABI ${ANDROID_ABI}) # This will be passed in from your GitHub Action

# Point CMake to the root of the Android NDK
if(NOT DEFINED ENV{ANDROID_NDK_HOME})
  message(FATAL_ERROR "The ANDROID_NDK_HOME environment variable must be set.")
endif()
set(CMAKE_ANDROID_NDK "$ENV{ANDROID_NDK_HOME}")

# Specify the C++ standard library to use
set(CMAKE_ANDROID_STL_TYPE "c++_static")

# Include the vcpkg toolchain file
if(DEFINED ENV{WORKSPACE})
  include("$ENV{WORKSPACE}/vcpkg/scripts/buildsystems/vcpkg.cmake")
endif()
