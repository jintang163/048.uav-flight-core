#ifndef TCP_TRANSPORT_H
#define TCP_TRANSPORT_H

#include "transport.h"

#define TCP_MAX_HOST_LEN    256
#define TCP_DEFAULT_PORT    5760
#define TCP_DEFAULT_KEEPALIVE    30

typedef struct {
    char host[TCP_MAX_HOST_LEN];
    uint16_t port;
    uint32_t connect_timeout_ms;
    uint32_t recv_timeout_ms;
    uint32_t send_timeout_ms;
    bool keepalive;
    uint32_t keepalive_idle;
    uint32_t keepalive_interval;
    uint32_t keepalive_count;
    bool use_tls;
} tcp_config_t;

typedef struct {
    transport_t base;
    tcp_config_t config;
    int fd;
    bool is_server;
    bool is_connected;
    void* tls_context;
} tcp_transport_t;

int tcp_transport_init(tcp_transport_t* transport, const tcp_config_t* config);
int tcp_transport_set_config(tcp_transport_t* transport, const tcp_config_t* config);
int tcp_transport_get_config(const tcp_transport_t* transport, tcp_config_t* config);
int tcp_transport_set_server_mode(tcp_transport_t* transport, bool enable);
int tcp_transport_accept(tcp_transport_t* server, tcp_transport_t* client);
void tcp_transport_default_config(tcp_config_t* config);

#endif
