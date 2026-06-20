#ifndef FRAGMENT_H
#define FRAGMENT_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

#define FRAGMENT_MAX_DATA_SIZE    4096
#define FRAGMENT_MAX_FRAGMENTS    64
#define FRAGMENT_DEFAULT_MTU    255
#define FRAGMENT_HEADER_SIZE    4

typedef enum {
    FRAGMENT_STATUS_INIT = 0,
    FRAGMENT_STATUS_IN_PROGRESS = 1,
    FRAGMENT_STATUS_COMPLETE = 2,
    FRAGMENT_STATUS_TIMEOUT = 3,
    FRAGMENT_STATUS_ERROR = 4
} fragment_status_t;

typedef struct __attribute__((packed)) {
    uint16_t packet_id;
    uint8_t fragment_index;
    uint8_t fragment_total;
    uint8_t data[251];
} fragment_packet_t;

typedef struct {
    uint8_t buffer[FRAGMENT_MAX_DATA_SIZE];
    size_t total_size;
    size_t received_size;
    uint16_t packet_id;
    uint8_t fragment_total;
    uint8_t fragment_mask[FRAGMENT_MAX_FRAGMENTS / 8 + 1];
    uint32_t first_fragment_time;
    uint32_t last_fragment_time;
    uint32_t timeout_ms;
    fragment_status_t status;
    uint8_t mtu;
} fragment_reassembly_t;

typedef struct {
    uint16_t next_packet_id;
    uint8_t mtu;
    uint32_t timeout_ms;
    uint64_t packets_fragmented;
    uint64_t packets_reassembled;
    uint64_t fragments_sent;
    uint64_t fragments_received;
    uint32_t error_count;
} fragment_manager_t;

int fragment_manager_init(fragment_manager_t* manager, uint8_t mtu, uint32_t timeout_ms);
int fragment_manager_set_mtu(fragment_manager_t* manager, uint8_t mtu);
int fragment_manager_set_timeout(fragment_manager_t* manager, uint32_t timeout_ms);
int fragment_manager_split(fragment_manager_t* manager, const uint8_t* data, size_t len,
                           uint8_t* fragments, size_t* fragment_count, size_t max_fragments);
int fragment_reassembly_init(fragment_reassembly_t* reassembly, uint8_t mtu, uint32_t timeout_ms);
int fragment_reassembly_add(fragment_reassembly_t* reassembly, const fragment_packet_t* fragment);
int fragment_reassembly_is_complete(const fragment_reassembly_t* reassembly);
int fragment_reassembly_get_data(const fragment_reassembly_t* reassembly, uint8_t* data, size_t* len);
fragment_status_t fragment_reassembly_get_status(const fragment_reassembly_t* reassembly);
int fragment_reassembly_reset(fragment_reassembly_t* reassembly);
bool fragment_manager_is_fragment_packet(const uint8_t* data, size_t len);

#endif
