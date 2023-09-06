cmake_minimum_required(VERSION 3.22)
project(testcpp CXX)
set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED True)

add_executable(testcpp ${PROJECT_SOURCE_DIR}/main.cpp)
