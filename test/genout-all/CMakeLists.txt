# Project metadata
cmake_minimum_required(VERSION 3.16)

project(gdbufgen VERSION 1.0 LANGUAGES CXX C)

# Emscripten hack for shared libraries (must be before add_library shared)
if(EMSCRIPTEN)
    set_property(GLOBAL PROPERTY TARGET_SUPPORTS_SHARED_LIBS TRUE)
    set(CMAKE_SHARED_LIBRARY_CREATE_C_FLAGS "-sSIDE_MODULE")
    set(CMAKE_SHARED_LIBRARY_CREATE_CXX_FLAGS "-sSIDE_MODULE")
    set(CMAKE_STRIP FALSE)
    set(CMAKE_SHARED_LIBRARY_SUFFIX ".wasm")
endif()

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# ensure things are ready to go
if(DEFINED CMAKE_TOOLCHAIN_FILE)
  message(STATUS "CMAKE_TOOLCHAIN_FILE is set to: ${CMAKE_TOOLCHAIN_FILE}")
endif()

message(STATUS "Running with CMAKE_SYSTEM_PROCESSOR: ${CMAKE_SYSTEM_PROCESSOR}")
message(STATUS "Running with CMAKE_CXX_FLAGS: ${CMAKE_CXX_FLAGS}")

# Nanopb
set(NANOPB_DIR "${CMAKE_SOURCE_DIR}/nanopb")
set(NANOPB_SOURCES
    "${NANOPB_DIR}/pb_common.c"
    "${NANOPB_DIR}/pb_encode.c"
    "${NANOPB_DIR}/pb_decode.c"
)
include_directories("${NANOPB_DIR}")
add_definitions(-DPB_ENABLE_MALLOC) # Enable malloc for dynamic allocation

# Source files
file(GLOB_RECURSE SOURCES "src/*.cpp" "src/*.h" "src/*.c" "src/*.cc")
add_library(gdbufgen SHARED ${SOURCES} ${NANOPB_SOURCES})
target_include_directories(gdbufgen PUBLIC ${CMAKE_SOURCE_DIR}/src/test/proto)
target_include_directories(gdbufgen PUBLIC ${CMAKE_SOURCE_DIR}/src/test/proto/nested/deeply)
target_include_directories(gdbufgen PUBLIC ${CMAKE_SOURCE_DIR}/src/test/proto)
target_include_directories(gdbufgen PUBLIC ${CMAKE_SOURCE_DIR}/src)

# where to build it to
set(OUT_DIR "${CMAKE_BINARY_DIR}/bin")

# setup godot-cpp
set(GODOTCPP_BUILD_PROFILE "${CMAKE_SOURCE_DIR}/build_profile.json")
add_subdirectory(godot-cpp SYSTEM)
get_target_property(GODOTCPP_SUFFIX godot::cpp GODOTCPP_SUFFIX)
set(GODOTCPP_SUFFIX ${GODOTCPP_SUFFIX} CACHE INTERNAL "Godot CPP Suffix")

target_link_libraries(gdbufgen PRIVATE godot::cpp)
set_target_properties(gdbufgen PROPERTIES 
  OUTPUT_NAME "gdbufgen${GODOTCPP_SUFFIX}"
  # Linux
  LIBRARY_OUTPUT_DIRECTORY "${OUT_DIR}"
  # Windows
  RUNTIME_OUTPUT_DIRECTORY "${OUT_DIR}"
  RUNTIME_OUTPUT_DIRECTORY_DEBUG "${OUT_DIR}"
  RUNTIME_OUTPUT_DIRECTORY_RELEASE "${OUT_DIR}"
  RUNTIME_OUTPUT_DIRECTORY_RELWITHDEBINFO "${OUT_DIR}"
  RUNTIME_OUTPUT_DIRECTORY_MINSIZEREL "${OUT_DIR}"

  CXX_VISIBILITY_PRESET "default"
  VISIBILITY_INLINES_HIDDEN 1
)
