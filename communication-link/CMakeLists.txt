cmake_minimum_required(VERSION 3.10)
project(communication_link C)

set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)

if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release)
endif()

add_compile_options(-Wall -Wextra -Werror)

include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}/include
    ${CMAKE_CURRENT_SOURCE_DIR}/mavlink
    ${CMAKE_CURRENT_SOURCE_DIR}/crypto
    ${CMAKE_CURRENT_SOURCE_DIR}/transport
)

set(MAVLINK_SOURCES
    mavlink/mavlink_v2.c
    mavlink/parser.c
    mavlink/encoder.c
    mavlink/heartbeat.c
    mavlink/command.c
    mavlink/mission.c
    mavlink/parameter.c
)

set(CRYPTO_SOURCES
    crypto/sm4.c
    crypto/tls_manager.c
    crypto/crypto_manager.c
    crypto/key_exchange.c
    crypto/auth.c
)

set(TRANSPORT_SOURCES
    transport/transport.c
    transport/uart_transport.c
    transport/tcp_transport.c
    transport/udp_transport.c
    transport/transport_mux.c
    transport/link_monitor.c
    transport/fragment.c
)

add_library(communication_link STATIC
    ${MAVLINK_SOURCES}
    ${CRYPTO_SOURCES}
    ${TRANSPORT_SOURCES}
)

find_package(MBEDTLS QUIET)
if(MBEDTLS_FOUND)
    include_directories(${MBEDTLS_INCLUDE_DIRS})
    target_link_libraries(communication_link ${MBEDTLS_LIBRARIES})
    add_definitions(-DHAVE_MBEDTLS)
endif()

target_link_libraries(communication_link pthread)

enable_testing()
add_subdirectory(tests)
add_subdirectory(examples)

install(TARGETS communication_link
    ARCHIVE DESTINATION lib
    LIBRARY DESTINATION lib
)

install(DIRECTORY include/ DESTINATION include/communication_link)
install(DIRECTORY mavlink/ DESTINATION include/communication_link/mavlink
    FILES_MATCHING PATTERN "*.h")
install(DIRECTORY crypto/ DESTINATION include/communication_link/crypto
    FILES_MATCHING PATTERN "*.h")
install(DIRECTORY transport/ DESTINATION include/communication_link/transport
    FILES_MATCHING PATTERN "*.h")
