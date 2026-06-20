#ifndef TRANSPORT_MUX_H
#define TRANSPORT_MUX_H

#include "transport.h"

#define MUX_MAX_LINKS    8
#define MUX_DEFAULT_PRIMARY    0
#define MUX_MAX_QUEUE_SIZE    64

typedef enum {
    MUX_MODE_PRIMARY = 0,
    MUX_MODE_REDUNDANT = 1,
    MUX_MODE_LOAD_BALANCE = 2,
    MUX_MODE_FAILOVER = 3
} mux_mode_t;

typedef enum {
    MUX_LINK_STATUS_UNKNOWN = 0,
    MUX_LINK_STATUS_ACTIVE = 1,
    MUX_LINK_STATUS_STANDBY = 2,
    MUX_LINK_STATUS_DEGRADED = 3,
    MUX_LINK_STATUS_FAILED = 4
} mux_link_status_t;

typedef struct {
    transport_t* transport;
    mux_link_status_t status;
    int32_t quality_score;
    uint32_t priority;
    uint64_t bytes_sent;
    uint64_t bytes_received;
    uint32_t failures;
    uint32_t last_used_ms;
    bool enabled;
} mux_link_t;

typedef struct {
    transport_t base;
    mux_mode_t mode;
    mux_link_t links[MUX_MAX_LINKS];
    uint8_t link_count;
    int32_t primary_link;
    int32_t active_link;
    uint32_t failover_threshold;
    uint32_t failback_delay_ms;
    uint32_t last_failover_ms;
    void* user_data;
} transport_mux_t;

int transport_mux_init(transport_mux_t* mux, mux_mode_t mode);
int transport_mux_add_link(transport_mux_t* mux, transport_t* transport, uint32_t priority);
int transport_mux_remove_link(transport_mux_t* mux, uint8_t index);
int transport_mux_set_primary(transport_mux_t* mux, uint8_t index);
int transport_mux_set_mode(transport_mux_t* mux, mux_mode_t mode);
int transport_mux_enable_link(transport_mux_t* mux, uint8_t index, bool enable);
int transport_mux_get_link_status(const transport_mux_t* mux, uint8_t index, mux_link_status_t* status);
int transport_mux_get_link_quality(const transport_mux_t* mux, uint8_t index, int32_t* quality);
int32_t transport_mux_get_active_link(const transport_mux_t* mux);
int transport_mux_set_failover_threshold(transport_mux_t* mux, uint32_t threshold);
int transport_mux_set_failback_delay(transport_mux_t* mux, uint32_t delay_ms);
int transport_mux_manual_switch(transport_mux_t* mux, uint8_t index);

#endif
