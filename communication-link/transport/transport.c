#include "transport.h"
#include <string.h>

int transport_init(transport_t* transport, transport_type_t type)
{
    if (!transport) return -1;
    memset(transport, 0, sizeof(transport_t));
    transport->type = type;
    transport->state = TRANSPORT_STATE_DISCONNECTED;
    transport->is_open = false;
    return 0;
}

int transport_connect(transport_t* transport)
{
    if (!transport || !transport->connect) return -1;
    return transport->connect(transport);
}

int transport_disconnect(transport_t* transport)
{
    if (!transport || !transport->disconnect) return -1;
    return transport->disconnect(transport);
}

int transport_send(transport_t* transport, const uint8_t* data, size_t len, uint32_t timeout_ms)
{
    if (!transport || !transport->send || !data || len == 0) return -1;
    int ret = transport->send(transport, data, len, timeout_ms);
    if (ret > 0) {
        transport->bytes_sent += (uint64_t)ret;
        transport->packet_count_sent++;
    } else if (ret < 0) {
        transport->error_count++;
    }
    return ret;
}

int transport_recv(transport_t* transport, uint8_t* data, size_t len, size_t* recv_len, uint32_t timeout_ms)
{
    if (!transport || !transport->recv || !data || len == 0 || !recv_len) return -1;
    int ret = transport->recv(transport, data, len, recv_len, timeout_ms);
    if (ret == 0 && *recv_len > 0) {
        transport->bytes_received += (uint64_t)*recv_len;
        transport->packet_count_received++;
    } else if (ret < 0) {
        transport->error_count++;
    }
    return ret;
}

int transport_set_state_callback(transport_t* transport, transport_state_change_cb_t cb, void* user_data)
{
    if (!transport) return -1;
    transport->on_state_change = cb;
    transport->user_data = user_data;
    return 0;
}

transport_state_t transport_get_state(const transport_t* transport)
{
    return transport ? transport->state : TRANSPORT_STATE_ERROR;
}

transport_type_t transport_get_type(const transport_t* transport)
{
    return transport ? transport->type : TRANSPORT_TYPE_UNKNOWN;
}

uint64_t transport_get_bytes_sent(const transport_t* transport)
{
    return transport ? transport->bytes_sent : 0;
}

uint64_t transport_get_bytes_received(const transport_t* transport)
{
    return transport ? transport->bytes_received : 0;
}

uint32_t transport_get_packet_count_sent(const transport_t* transport)
{
    return transport ? transport->packet_count_sent : 0;
}

uint32_t transport_get_packet_count_received(const transport_t* transport)
{
    return transport ? transport->packet_count_received : 0;
}

uint32_t transport_get_error_count(const transport_t* transport)
{
    return transport ? transport->error_count : 0;
}

void transport_reset_stats(transport_t* transport)
{
    if (!transport) return;
    transport->bytes_sent = 0;
    transport->bytes_received = 0;
    transport->packet_count_sent = 0;
    transport->packet_count_received = 0;
    transport->error_count = 0;
}

void transport_cleanup(transport_t* transport)
{
    if (!transport) return;
    if (transport->state == TRANSPORT_STATE_CONNECTED) {
        transport_disconnect(transport);
    }
    memset(transport, 0, sizeof(transport_t));
}
