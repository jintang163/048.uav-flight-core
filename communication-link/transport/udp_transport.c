#include "udp_transport.h"
#include <string.h>
#include <stdlib.h>

static int udp_transport_connect_impl(transport_t* transport)
{
    udp_transport_t* udp = (udp_transport_t*)transport;
    if (!udp) return -1;
    udp->base.state = TRANSPORT_STATE_CONNECTED;
    udp->is_bound = true;
    if (udp->base.on_state_change) {
        udp->base.on_state_change(transport, TRANSPORT_STATE_CONNECTED, udp->base.user_data);
    }
    return 0;
}

static int udp_transport_disconnect_impl(transport_t* transport)
{
    udp_transport_t* udp = (udp_transport_t*)transport;
    if (!udp) return -1;
    udp->base.state = TRANSPORT_STATE_DISCONNECTED;
    udp->is_bound = false;
    if (udp->base.on_state_change) {
        udp->base.on_state_change(transport, TRANSPORT_STATE_DISCONNECTED, udp->base.user_data);
    }
    return 0;
}

static int udp_transport_send_impl(transport_t* transport, const uint8_t* data, size_t len, uint32_t timeout_ms)
{
    udp_transport_t* udp = (udp_transport_t*)transport;
    (void)timeout_ms;
    if (!udp || !udp->is_bound || !data || len == 0) return -1;
    if (len > TRANSPORT_MAX_PACKET_SIZE) return -1;
    return (int)len;
}

static int udp_transport_recv_impl(transport_t* transport, uint8_t* data, size_t len, size_t* recv_len, uint32_t timeout_ms)
{
    udp_transport_t* udp = (udp_transport_t*)transport;
    (void)timeout_ms;
    if (!udp || !udp->is_bound || !data || len == 0 || !recv_len) return -1;
    *recv_len = 0;
    return 0;
}

int udp_transport_init(udp_transport_t* transport, const udp_config_t* config)
{
    if (!transport) return -1;
    memset(transport, 0, sizeof(udp_transport_t));
    transport_init(&transport->base, TRANSPORT_TYPE_UDP);
    transport->base.connect = udp_transport_connect_impl;
    transport->base.disconnect = udp_transport_disconnect_impl;
    transport->base.send = udp_transport_send_impl;
    transport->base.recv = udp_transport_recv_impl;
    transport->fd = -1;
    transport->is_bound = false;
    if (config) {
        memcpy(&transport->config, config, sizeof(udp_config_t));
    } else {
        udp_transport_default_config(&transport->config);
    }
    return 0;
}

int udp_transport_set_config(udp_transport_t* transport, const udp_config_t* config)
{
    if (!transport || !config) return -1;
    memcpy(&transport->config, config, sizeof(udp_config_t));
    return 0;
}

int udp_transport_get_config(const udp_transport_t* transport, udp_config_t* config)
{
    if (!transport || !config) return -1;
    memcpy(config, &transport->config, sizeof(udp_config_t));
    return 0;
}

int udp_transport_add_broadcast(udp_transport_t* transport, const char* addr, uint16_t port)
{
    if (!transport || !addr) return -1;
    if (transport->config.broadcast_count >= UDP_MAX_BROADCAST_ADDRS) return -1;
    strncpy(transport->config.broadcast_addrs[transport->config.broadcast_count],
            addr, UDP_MAX_HOST_LEN - 1);
    transport->config.broadcast_ports[transport->config.broadcast_count] = port;
    transport->config.broadcast_count++;
    transport->config.broadcast = true;
    return 0;
}

int udp_transport_send_broadcast(udp_transport_t* transport, const uint8_t* data, size_t len)
{
    if (!transport || !data || len == 0) return -1;
    if (!transport->config.broadcast || transport->config.broadcast_count == 0) return -1;
    if (len > TRANSPORT_MAX_PACKET_SIZE) return -1;
    transport->base.bytes_sent += (uint64_t)len * transport->config.broadcast_count;
    transport->base.packet_count_sent += transport->config.broadcast_count;
    return (int)(len * transport->config.broadcast_count);
}

int udp_transport_recv_from(udp_transport_t* transport, uint8_t* data, size_t len, size_t* recv_len,
                            char* from_addr, size_t addr_len, uint16_t* from_port, uint32_t timeout_ms)
{
    (void)timeout_ms;
    if (!transport || !data || len == 0 || !recv_len) return -1;
    *recv_len = 0;
    if (from_addr && addr_len > 0) {
        from_addr[0] = '\0';
    }
    if (from_port) {
        *from_port = 0;
    }
    return 0;
}

void udp_transport_default_config(udp_config_t* config)
{
    if (!config) return;
    memset(config, 0, sizeof(udp_config_t));
    strcpy(config->local_host, "0.0.0.0");
    config->local_port = UDP_DEFAULT_PORT;
    strcpy(config->remote_host, "127.0.0.1");
    config->remote_port = UDP_DEFAULT_PORT;
    config->broadcast = false;
    config->broadcast_count = 0;
    config->recv_timeout_ms = 100;
    config->ttl = UDP_DEFAULT_TTL;
    config->reuse_addr = true;
}
