#include "4g_driver.h"
#include "main.h"
#include <string.h>
#include <stdlib.h>
#include <stdio.h>

#define LTE_UART huart4

#define AT_CMD_BUFFER_SIZE 128
#define AT_RESPONSE_BUFFER_SIZE 512
#define RING_BUFFER_SIZE 1024

#define LTE_MIN_RSSI -120
#define LTE_MAX_RSSI -51

#define AT_TIMEOUT_MS 3000
#define AT_RETRY_MAX 3
#define CSQ_INTERVAL_MS 5000

typedef enum {
    LTE_INIT = 0,
    LTE_CHECK_ECHO,
    LTE_SET_CPIN,
    LTE_SET_CREG,
    LTE_SET_CSQ,
    LTE_SET_COPS,
    LTE_SET_NETWORK,
    LTE_SET_PDP,
    LTE_CONNECT,
    LTE_CONNECTED,
    LTE_ERROR
} LTEInternalState;

typedef struct {
    uint8_t buffer[RING_BUFFER_SIZE];
    uint16_t head;
    uint16_t tail;
} RingBuffer;

static LTEStatus lte_status;
static LTEInternalState internal_state;
static uint32_t last_update_time;
static uint32_t last_csq_time;
static bool initialized = false;

static char at_cmd_buffer[AT_CMD_BUFFER_SIZE];
static char at_response_buffer[AT_RESPONSE_BUFFER_SIZE];
static uint16_t response_index = 0;

static RingBuffer rx_ring_buffer;

static uint8_t retry_count = 0;
static uint32_t cmd_send_time = 0;
static bool cmd_pending = false;

static void ring_buffer_init(RingBuffer *rb)
{
    rb->head = 0;
    rb->tail = 0;
}

static bool ring_buffer_put(RingBuffer *rb, uint8_t byte)
{
    uint16_t next = (rb->head + 1) % RING_BUFFER_SIZE;
    if (next == rb->tail) {
        return false;
    }
    rb->buffer[rb->head] = byte;
    rb->head = next;
    return true;
}

static bool ring_buffer_get(RingBuffer *rb, uint8_t *byte)
{
    if (rb->head == rb->tail) {
        return false;
    }
    *byte = rb->buffer[rb->tail];
    rb->tail = (rb->tail + 1) % RING_BUFFER_SIZE;
    return true;
}

static void _4g_driver_send_at_cmd(const char *cmd)
{
    uint16_t len = strlen(cmd);
    if (len > AT_CMD_BUFFER_SIZE - 4) {
        return;
    }

    snprintf(at_cmd_buffer, sizeof(at_cmd_buffer), "%s\r\n", cmd);
    HAL_UART_Transmit(&LTE_UART, (uint8_t *)at_cmd_buffer, strlen(at_cmd_buffer), 100);
    cmd_send_time = HAL_GetTick();
    cmd_pending = true;
}

static void _4g_driver_parse_csq(const char *response)
{
    int rssi, ber;
    if (sscanf(response, "+CSQ: %d,%d", &rssi, &ber) == 2) {
        lte_status.csq = (uint8_t)rssi;
        if (rssi == 99) {
            lte_status.rssi = LTE_MIN_RSSI;
        } else {
            lte_status.rssi = (int8_t)(-113 + rssi * 2);
        }
        lte_status.ber = (int8_t)ber;
    }
}

static void _4g_driver_parse_creg(const char *response)
{
    int stat;
    if (sscanf(response, "+CREG: %*d,%d", &stat) == 1) {
        lte_status.registered = (stat == 1 || stat == 5);
        if (lte_status.registered) {
            lte_status.network_type = NETWORK_TYPE_4G;
        } else {
            lte_status.network_type = NETWORK_TYPE_NONE;
        }
    }
}

static void _4g_driver_parse_cops(const char *response)
{
    char *p = strstr(response, "+COPS:");
    if (p != NULL) {
        char *start = strchr(p, '"');
        if (start != NULL) {
            start++;
            char *end = strchr(start, '"');
            if (end != NULL) {
                int len = end - start;
                if (len > 15) len = 15;
                strncpy(lte_status.operator_name, start, len);
                lte_status.operator_name[len] = '\0';
            }
        }
    }
}

static void _4g_driver_process_response_line(const char *line)
{
    if (strstr(line, "+CSQ:") != NULL) {
        _4g_driver_parse_csq(line);
    } else if (strstr(line, "+CREG:") != NULL) {
        _4g_driver_parse_creg(line);
    } else if (strstr(line, "+COPS:") != NULL) {
        _4g_driver_parse_cops(line);
    } else if (strstr(line, "OK") != NULL) {
        cmd_pending = false;
        retry_count = 0;

        switch (internal_state) {
            case LTE_CHECK_ECHO:
                internal_state = LTE_SET_CPIN;
                _4g_driver_send_at_cmd("AT+CPIN?");
                break;
            case LTE_SET_CPIN:
                internal_state = LTE_SET_CREG;
                _4g_driver_send_at_cmd("AT+CREG?");
                break;
            case LTE_SET_CREG:
                internal_state = LTE_SET_NETWORK;
                _4g_driver_send_at_cmd("AT+CNMP=2");
                break;
            case LTE_SET_NETWORK:
                internal_state = LTE_SET_COPS;
                _4g_driver_send_at_cmd("AT+COPS?");
                break;
            case LTE_SET_COPS:
                internal_state = LTE_SET_PDP;
                _4g_driver_send_at_cmd("AT+CGDCONT=1,\"IP\",\"CMNET\"");
                break;
            case LTE_SET_PDP:
                internal_state = LTE_CONNECT;
                _4g_driver_send_at_cmd("AT+CGACT=1,1");
                break;
            case LTE_CONNECT:
                internal_state = LTE_CONNECTED;
                last_csq_time = HAL_GetTick();
                break;
            case LTE_CONNECTED:
                break;
            default:
                break;
        }
    } else if (strstr(line, "ERROR") != NULL) {
        cmd_pending = false;
        retry_count++;

        if (retry_count >= AT_RETRY_MAX) {
            if (internal_state != LTE_CONNECTED) {
                internal_state = LTE_ERROR;
            }
            retry_count = 0;
        } else {
            uint32_t now = HAL_GetTick();
            cmd_send_time = now;
            cmd_pending = true;

            switch (internal_state) {
                case LTE_CHECK_ECHO:
                    _4g_driver_send_at_cmd("ATE0");
                    break;
                case LTE_SET_CPIN:
                    _4g_driver_send_at_cmd("AT+CPIN?");
                    break;
                case LTE_SET_CREG:
                    _4g_driver_send_at_cmd("AT+CREG?");
                    break;
                case LTE_SET_NETWORK:
                    _4g_driver_send_at_cmd("AT+CNMP=2");
                    break;
                case LTE_SET_COPS:
                    _4g_driver_send_at_cmd("AT+COPS?");
                    break;
                case LTE_SET_PDP:
                    _4g_driver_send_at_cmd("AT+CGDCONT=1,\"IP\",\"CMNET\"");
                    break;
                case LTE_CONNECT:
                    _4g_driver_send_at_cmd("AT+CGACT=1,1");
                    break;
                default:
                    break;
            }
        }
    }
}

static void _4g_driver_process_ring_buffer(void)
{
    uint8_t byte;
    while (ring_buffer_get(&rx_ring_buffer, &byte)) {
        if (response_index < AT_RESPONSE_BUFFER_SIZE - 1) {
            at_response_buffer[response_index++] = byte;
            at_response_buffer[response_index] = '\0';

            if (byte == '\n' && response_index > 2) {
                char line[256];
                strncpy(line, at_response_buffer, sizeof(line) - 1);
                line[sizeof(line) - 1] = '\0';

                char *start = line;
                while (*start == '\r' || *start == '\n') start++;
                char *end = start + strlen(start) - 1;
                while (end > start && (*end == '\r' || *end == '\n' || *end == ' ')) {
                    *end = '\0';
                    end--;
                }

                if (strlen(start) > 0) {
                    _4g_driver_process_response_line(start);
                }

                response_index = 0;
                memset(at_response_buffer, 0, sizeof(at_response_buffer));
            }
        } else {
            response_index = 0;
            memset(at_response_buffer, 0, sizeof(at_response_buffer));
        }
    }
}

static void _4g_driver_check_timeout(void)
{
    if (!cmd_pending) {
        return;
    }

    uint32_t now = HAL_GetTick();
    if (now - cmd_send_time > AT_TIMEOUT_MS) {
        cmd_pending = false;
        retry_count++;

        if (retry_count >= AT_RETRY_MAX) {
            if (internal_state != LTE_CONNECTED) {
                internal_state = LTE_ERROR;
            }
            retry_count = 0;
        } else {
            switch (internal_state) {
                case LTE_INIT:
                    _4g_driver_send_at_cmd("AT");
                    break;
                case LTE_CHECK_ECHO:
                    _4g_driver_send_at_cmd("ATE0");
                    break;
                case LTE_SET_CPIN:
                    _4g_driver_send_at_cmd("AT+CPIN?");
                    break;
                case LTE_SET_CREG:
                    _4g_driver_send_at_cmd("AT+CREG?");
                    break;
                case LTE_SET_NETWORK:
                    _4g_driver_send_at_cmd("AT+CNMP=2");
                    break;
                case LTE_SET_COPS:
                    _4g_driver_send_at_cmd("AT+COPS?");
                    break;
                case LTE_SET_PDP:
                    _4g_driver_send_at_cmd("AT+CGDCONT=1,\"IP\",\"CMNET\"");
                    break;
                case LTE_CONNECT:
                    _4g_driver_send_at_cmd("AT+CGACT=1,1");
                    break;
                default:
                    break;
            }
        }
    }
}

bool _4g_driver_init(void)
{
    memset(&lte_status, 0, sizeof(LTEStatus));
    lte_status.rssi = LTE_MIN_RSSI;
    lte_status.ber = 99;
    lte_status.csq = 99;
    lte_status.network_type = NETWORK_TYPE_NONE;
    lte_status.registered = false;
    memset(lte_status.operator_name, 0, sizeof(lte_status.operator_name));

    internal_state = LTE_INIT;
    last_update_time = 0;
    last_csq_time = 0;
    response_index = 0;
    retry_count = 0;
    cmd_pending = false;

    ring_buffer_init(&rx_ring_buffer);

    initialized = true;
    return true;
}

void _4g_driver_update(void)
{
    if (!initialized) {
        return;
    }

    uint32_t now = HAL_GetTick();

    _4g_driver_process_ring_buffer();
    _4g_driver_check_timeout();

    switch (internal_state) {
        case LTE_INIT:
            if (!cmd_pending) {
                retry_count = 0;
                _4g_driver_send_at_cmd("AT");
                internal_state = LTE_CHECK_ECHO;
            }
            break;

        case LTE_CHECK_ECHO:
        case LTE_SET_CPIN:
        case LTE_SET_CREG:
        case LTE_SET_NETWORK:
        case LTE_SET_COPS:
        case LTE_SET_PDP:
        case LTE_CONNECT:
            break;

        case LTE_CONNECTED:
            if (now - last_csq_time >= CSQ_INTERVAL_MS) {
                if (!cmd_pending) {
                    _4g_driver_send_at_cmd("AT+CSQ");
                    last_csq_time = now;
                }
            }
            break;

        case LTE_ERROR:
            if (now - last_update_time > 5000) {
                internal_state = LTE_INIT;
                last_update_time = now;
            }
            break;

        default:
            break;
    }
}

bool _4g_driver_get_status(LTEStatus *status)
{
    if (status == NULL) {
        return false;
    }
    *status = lte_status;
    return true;
}

bool _4g_driver_is_connected(void)
{
    return lte_status.registered && internal_state == LTE_CONNECTED;
}

int8_t _4g_driver_get_rssi(void)
{
    return lte_status.rssi;
}

void _4g_driver_process_byte(uint8_t byte)
{
    ring_buffer_put(&rx_ring_buffer, byte);
}
