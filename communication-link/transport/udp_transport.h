#ifndef UDP_TRANSPORT_H
#define UDP_TRANSPORT_H

#include "transport.h"

#define UDP_MAX_HOST_LEN    256
#define UDP_DEFAULT_PORT    14550
#define UDP_MAX_BROADCAST_ADDRS    8
#define UDP_DEFAULT_TTL    64

typedef struct {
    char local_host[UDP_MAX_HOST_LEN];
    uint16_t local_port;
    char remote_host[UDP_MAX_HOST_LEN];
    uint16_t remote_port;
    bool broadcast;
    char broadcast_addrs[UDP_MAX_BROADCAST_ADDRS][UDP_MAX_HOST_LEN];
    uint16_t broadcast_ports[UDP_MAX_BROADCAST_ADDRS];
    uint8_t broadcast_count;
    uint32_t recv_timeout_ms;
    uint8_t ttl;
    bool reuse_addr;
} udp_config_t;

typedef struct {
    transport_t base;
    udp_config_t config;
    int fd;
    bool is_bound;
    char remote_addr[UDP_MAX_HOST_LEN];
    uint16_t remote_port;
} udp_transport_t;

int udp_transport_init(udp_transport_t* transport, const udp_config_t* config);
int udp_transport_set_config(udp_transport_t* transport, const udp_config_t* config);
int udp_transport_get_config(const udp_transport_t* transport, udp_config_t* config);
int udp_transport_add_broadcast(udp_transport_t* transport, const char* addr, uint16_t port);
int udp_transport_send_broadcast(udp_transport_t* transport, const uint8_t* data, size_t len);
int udp_transport_recv_from(udp_transport_t* transport, uint8_t* data, size_t len, size_t* recv_len,
                            char* from_addr, size_t addr_len, uint16_t* from_port, uint32_t timeout_ms);
void udp_transport_default_config(udp_config_t* config);

#endif
