#include "uart_transport.h"
#include <string.h>
#include <stdlib.h>

static int uart_transport_connect_impl(transport_t* transport)
{
    uart_transport_t* uart = (uart_transport_t*)transport;
    if (!uart) return -1;
    uart->base.state = TRANSPORT_STATE_CONNECTED;
    uart->is_open = true;
    if (uart->base.on_state_change) {
        uart->base.on_state_change(transport, TRANSPORT_STATE_CONNECTED, uart->base.user_data);
    }
    return 0;
}

static int uart_transport_disconnect_impl(transport_t* transport)
{
    uart_transport_t* uart = (uart_transport_t*)transport;
    if (!uart) return -1;
    uart->base.state = TRANSPORT_STATE_DISCONNECTED;
    uart->is_open = false;
    if (uart->base.on_state_change) {
        uart->base.on_state_change(transport, TRANSPORT_STATE_DISCONNECTED, uart->base.user_data);
    }
    return 0;
}

static int uart_transport_send_impl(transport_t* transport, const uint8_t* data, size_t len, uint32_t timeout_ms)
{
    uart_transport_t* uart = (uart_transport_t*)transport;
    (void)timeout_ms;
    if (!uart || !uart->is_open || !data || len == 0) return -1;
    if (len > TRANSPORT_MAX_PACKET_SIZE) return -1;
    return (int)len;
}

static int uart_transport_recv_impl(transport_t* transport, uint8_t* data, size_t len, size_t* recv_len, uint32_t timeout_ms)
{
    uart_transport_t* uart = (uart_transport_t*)transport;
    (void)timeout_ms;
    if (!uart || !uart->is_open || !data || len == 0 || !recv_len) return -1;
    *recv_len = 0;
    return 0;
}

int uart_transport_init(uart_transport_t* transport, const uart_config_t* config)
{
    if (!transport) return -1;
    memset(transport, 0, sizeof(uart_transport_t));
    transport_init(&transport->base, TRANSPORT_TYPE_UART);
    transport->base.connect = uart_transport_connect_impl;
    transport->base.disconnect = uart_transport_disconnect_impl;
    transport->base.send = uart_transport_send_impl;
    transport->base.recv = uart_transport_recv_impl;
    transport->fd = -1;
    transport->is_open = false;
    if (config) {
        memcpy(&transport->config, config, sizeof(uart_config_t));
    } else {
        uart_transport_default_config(&transport->config);
    }
    return 0;
}

int uart_transport_set_config(uart_transport_t* transport, const uart_config_t* config)
{
    if (!transport || !config) return -1;
    memcpy(&transport->config, config, sizeof(uart_config_t));
    return 0;
}

int uart_transport_get_config(const uart_transport_t* transport, uart_config_t* config)
{
    if (!transport || !config) return -1;
    memcpy(config, &transport->config, sizeof(uart_config_t));
    return 0;
}

void uart_transport_default_config(uart_config_t* config)
{
    if (!config) return;
    memset(config, 0, sizeof(uart_config_t));
    strcpy(config->port_name, "/dev/ttyUSB0");
    config->baud_rate = UART_DEFAULT_BAUD_RATE;
    config->data_bits = 8;
    config->parity = UART_PARITY_NONE;
    config->stop_bits = UART_STOP_BITS_1;
    config->flow_control = UART_FLOW_CONTROL_NONE;
    config->read_timeout_ms = 100;
    config->write_timeout_ms = 100;
}
