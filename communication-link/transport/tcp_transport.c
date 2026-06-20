#include "tcp_transport.h"
#include <string.h>
#include <stdlib.h>

static int tcp_transport_connect_impl(transport_t* transport)
{
    tcp_transport_t* tcp = (tcp_transport_t*)transport;
    if (!tcp) return -1;
    tcp->base.state = TRANSPORT_STATE_CONNECTED;
    tcp->is_connected = true;
    if (tcp->base.on_state_change) {
        tcp->base.on_state_change(transport, TRANSPORT_STATE_CONNECTED, tcp->base.user_data);
    }
    return 0;
}

static int tcp_transport_disconnect_impl(transport_t* transport)
{
    tcp_transport_t* tcp = (tcp_transport_t*)transport;
    if (!tcp) return -1;
    tcp->base.state = TRANSPORT_STATE_DISCONNECTED;
    tcp->is_connected = false;
    if (tcp->base.on_state_change) {
        tcp->base.on_state_change(transport, TRANSPORT_STATE_DISCONNECTED, tcp->base.user_data);
    }
    return 0;
}

static int tcp_transport_send_impl(transport_t* transport, const uint8_t* data, size_t len, uint32_t timeout_ms)
{
    tcp_transport_t* tcp = (tcp_transport_t*)transport;
    (void)timeout_ms;
    if (!tcp || !tcp->is_connected || !data || len == 0) return -1;
    if (len > TRANSPORT_MAX_PACKET_SIZE) return -1;
    return (int)len;
}

static int tcp_transport_recv_impl(transport_t* transport, uint8_t* data, size_t len, size_t* recv_len, uint32_t timeout_ms)
{
    tcp_transport_t* tcp = (tcp_transport_t*)transport;
    (void)timeout_ms;
    if (!tcp || !tcp->is_connected || !data || len == 0 || !recv_len) return -1;
    *recv_len = 0;
    return 0;
}

int tcp_transport_init(tcp_transport_t* transport, const tcp_config_t* config)
{
    if (!transport) return -1;
    memset(transport, 0, sizeof(tcp_transport_t));
    transport_init(&transport->base, TRANSPORT_TYPE_TCP);
    transport->base.connect = tcp_transport_connect_impl;
    transport->base.disconnect = tcp_transport_disconnect_impl;
    transport->base.send = tcp_transport_send_impl;
    transport->base.recv = tcp_transport_recv_impl;
    transport->fd = -1;
    transport->is_server = false;
    transport->is_connected = false;
    transport->tls_context = NULL;
    if (config) {
        memcpy(&transport->config, config, sizeof(tcp_config_t));
    } else {
        tcp_transport_default_config(&transport->config);
    }
    return 0;
}

int tcp_transport_set_config(tcp_transport_t* transport, const tcp_config_t* config)
{
    if (!transport || !config) return -1;
    memcpy(&transport->config, config, sizeof(tcp_config_t));
    return 0;
}

int tcp_transport_get_config(const tcp_transport_t* transport, tcp_config_t* config)
{
    if (!transport || !config) return -1;
    memcpy(config, &transport->config, sizeof(tcp_config_t));
    return 0;
}

int tcp_transport_set_server_mode(tcp_transport_t* transport, bool enable)
{
    if (!transport) return -1;
    transport->is_server = enable;
    return 0;
}

int tcp_transport_accept(tcp_transport_t* server, tcp_transport_t* client)
{
    if (!server || !client || !server->is_server) return -1;
    memcpy(&client->config, &server->config, sizeof(tcp_config_t));
    client->is_connected = true;
    client->base.state = TRANSPORT_STATE_CONNECTED;
    return 0;
}

void tcp_transport_default_config(tcp_config_t* config)
{
    if (!config) return;
    memset(config, 0, sizeof(tcp_config_t));
    strcpy(config->host, "127.0.0.1");
    config->port = TCP_DEFAULT_PORT;
    config->connect_timeout_ms = 5000;
    config->recv_timeout_ms = 1000;
    config->send_timeout_ms = 1000;
    config->keepalive = true;
    config->keepalive_idle = 30;
    config->keepalive_interval = 10;
    config->keepalive_count = 3;
    config->use_tls = false;
}
