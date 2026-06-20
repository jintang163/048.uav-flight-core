#ifndef UART_TRANSPORT_H
#define UART_TRANSPORT_H

#include "transport.h"

#define UART_MAX_PORT_NAME_LEN    64
#define UART_DEFAULT_BAUD_RATE    57600

typedef enum {
    UART_PARITY_NONE = 0,
    UART_PARITY_ODD = 1,
    UART_PARITY_EVEN = 2
} uart_parity_t;

typedef enum {
    UART_STOP_BITS_1 = 1,
    UART_STOP_BITS_2 = 2
} uart_stop_bits_t;

typedef enum {
    UART_FLOW_CONTROL_NONE = 0,
    UART_FLOW_CONTROL_HARDWARE = 1,
    UART_FLOW_CONTROL_SOFTWARE = 2
} uart_flow_control_t;

typedef struct {
    char port_name[UART_MAX_PORT_NAME_LEN];
    uint32_t baud_rate;
    uint8_t data_bits;
    uart_parity_t parity;
    uart_stop_bits_t stop_bits;
    uart_flow_control_t flow_control;
    uint32_t read_timeout_ms;
    uint32_t write_timeout_ms;
} uart_config_t;

typedef struct {
    transport_t base;
    uart_config_t config;
    int fd;
    bool is_open;
} uart_transport_t;

int uart_transport_init(uart_transport_t* transport, const uart_config_t* config);
int uart_transport_set_config(uart_transport_t* transport, const uart_config_t* config);
int uart_transport_get_config(const uart_transport_t* transport, uart_config_t* config);
void uart_transport_default_config(uart_config_t* config);

#endif
