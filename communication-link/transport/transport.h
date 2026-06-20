#ifndef TRANSPORT_H
#define TRANSPORT_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

#define TRANSPORT_MAX_PACKET_SIZE    4096
#define TRANSPORT_DEFAULT_TIMEOUT    5000

typedef enum {
    TRANSPORT_TYPE_UNKNOWN = 0,
    TRANSPORT_TYPE_UART = 1,
    TRANSPORT_TYPE_TCP = 2,
    TRANSPORT_TYPE_UDP = 3,
    TRANSPORT_TYPE_MUX = 4
} transport_type_t;

typedef enum {
    TRANSPORT_STATE_DISCONNECTED = 0,
    TRANSPORT_STATE_CONNECTING = 1,
    TRANSPORT_STATE_CONNECTED = 2,
    TRANSPORT_STATE_DISCONNECTING = 3,
    TRANSPORT_STATE_ERROR = 4
} transport_state_t;

typedef struct transport_s transport_t;

typedef int (*transport_send_cb_t)(transport_t* transport, const uint8_t* data, size_t len, uint32_t timeout_ms);
typedef int (*transport_recv_cb_t)(transport_t* transport, uint8_t* data, size_t len, size_t* recv_len, uint32_t timeout_ms);
typedef int (*transport_connect_cb_t)(transport_t* transport);
typedef int (*transport_disconnect_cb_t)(transport_t* transport);
typedef void (*transport_state_change_cb_t)(transport_t* transport, transport_state_t new_state, void* user_data);

struct transport_s {
    transport_type_t type;
    transport_state_t state;
    transport_send_cb_t send;
    transport_recv_cb_t recv;
    transport_connect_cb_t connect;
    transport_disconnect_cb_t disconnect;
    transport_state_change_cb_t on_state_change;
    void* user_data;
    uint64_t bytes_sent;
    uint64_t bytes_received;
    uint32_t packet_count_sent;
    uint32_t packet_count_received;
    uint32_t error_count;
    bool is_open;
};

int transport_init(transport_t* transport, transport_type_t type);
int transport_connect(transport_t* transport);
int transport_disconnect(transport_t* transport);
int transport_send(transport_t* transport, const uint8_t* data, size_t len, uint32_t timeout_ms);
int transport_recv(transport_t* transport, uint8_t* data, size_t len, size_t* recv_len, uint32_t timeout_ms);
int transport_set_state_callback(transport_t* transport, transport_state_change_cb_t cb, void* user_data);
transport_state_t transport_get_state(const transport_t* transport);
transport_type_t transport_get_type(const transport_t* transport);
uint64_t transport_get_bytes_sent(const transport_t* transport);
uint64_t transport_get_bytes_received(const transport_t* transport);
uint32_t transport_get_packet_count_sent(const transport_t* transport);
uint32_t transport_get_packet_count_received(const transport_t* transport);
uint32_t transport_get_error_count(const transport_t* transport);
void transport_reset_stats(transport_t* transport);
void transport_cleanup(transport_t* transport);

#endif
