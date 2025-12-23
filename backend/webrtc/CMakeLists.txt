cmake_minimum_required(VERSION 3.15)
project(WebRTCService LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Dependencies
find_package(Threads REQUIRED)
find_package(LibDataChannel CONFIG REQUIRED)
find_package(OpenSSL REQUIRED)
find_package(Protobuf CONFIG REQUIRED)
find_package(gRPC CONFIG REQUIRED)

# Find the gRPC plugin
find_program(GRPC_CPP_PLUGIN grpc_cpp_plugin)
if(NOT GRPC_CPP_PLUGIN)
    # Try common locations
    find_program(GRPC_CPP_PLUGIN 
        NAMES grpc_cpp_plugin
        PATHS 
            /usr/local/bin
            /usr/bin
            ${CMAKE_INSTALL_PREFIX}/bin
        REQUIRED
    )
endif()

message(STATUS "Found gRPC plugin: ${GRPC_CPP_PLUGIN}")

# Get protoc
find_program(PROTOC_EXECUTABLE protoc REQUIRED)
message(STATUS "Found protoc: ${PROTOC_EXECUTABLE}")

# Proto file
set(PROTO_FILE "${CMAKE_CURRENT_SOURCE_DIR}/proto/webrtc_service.proto")

# Generated files
set(PROTO_GEN_DIR "${CMAKE_CURRENT_BINARY_DIR}")
set(PROTO_SRCS "${CMAKE_CURRENT_BINARY_DIR}/webrtc_service.pb.cc")
set(PROTO_HDRS "${CMAKE_CURRENT_BINARY_DIR}/webrtc_service.pb.h")
set(GRPC_SRCS "${CMAKE_CURRENT_BINARY_DIR}/webrtc_service.grpc.pb.cc")
set(GRPC_HDRS "${CMAKE_CURRENT_BINARY_DIR}/webrtc_service.grpc.pb.h")

add_custom_command(
    OUTPUT ${PROTO_SRCS} ${PROTO_HDRS} ${GRPC_SRCS} ${GRPC_HDRS}
    COMMAND ${PROTOC_EXECUTABLE}
        --grpc_out=${CMAKE_CURRENT_BINARY_DIR}
        --cpp_out=${CMAKE_CURRENT_BINARY_DIR}
        -I ${CMAKE_CURRENT_SOURCE_DIR}/proto
        --plugin=protoc-gen-grpc=${GRPC_CPP_PLUGIN}
        ${PROTO_FILE}
    DEPENDS ${PROTO_FILE}
    COMMENT "Generating protobuf + gRPC files"
    VERBATIM
)

include_directories(${CMAKE_CURRENT_BINARY_DIR})

add_executable(${PROJECT_NAME}
    ${PROTO_SRCS}
    ${GRPC_SRCS}
    src/main.cpp
    src/session/PeerSession.cpp
    src/session/SessionManager.cpp
    src/signaling/SignalStream.cpp
    src/dispatcher/Dispatcher.cpp
    src/rtc/RtcPeer.cpp
    src/grpc/WebRTCServiceImpl.cpp
)

target_include_directories(${PROJECT_NAME} PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/src
    ${PROTO_GEN_DIR}
)

# Link libraries
target_link_libraries(${PROJECT_NAME} PRIVATE
    gRPC::grpc++
    gRPC::grpc++_reflection
    protobuf::libprotobuf
    LibDataChannel::LibDataChannel
    Threads::Threads
    OpenSSL::SSL
    OpenSSL::Crypto
)