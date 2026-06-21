#include "4g_driver.h"
#include "main.h"
#include <string.h>
#include <stdlib.h>

#define LTE_UART huart3

#define AT_CMD_BUFFER_SIZE 128
#define AT_RESPONSE_BUFFER_SIZE 256

#define LTE_MIN_RSSI -120
#define LTE_MAX_RSSI -30

typedef enum {
    LTE_STATE_INIT = 0,
    LTE_STATE_CHECK_MODULE,
    LTE_STATE_SET_FUNCTIONALITY,
    LTE_STATE_CHECK_SIM,
    LTE_STATE_CHECK_REGISTRATION,
    LTE_STATE_SET_PDP_CONTEXT,
    LTE_STATE_ACTIVATE_PDP,
    LTE_STATE_CONNECTED,
    LTE_STATE_ERROR
} LTEInternalState;

static LTEStatus lte_status;
static LTEInternalState internal_state;
static uint32_t last_update_time;
static bool initialized = false;

static char at_cmd_buffer[AT_CMD_BUFFER_SIZE];
static char at_response_buffer[AT_RESPONSE_BUFFER_SIZE];
static uint16_t response_index = 0;

static int8_t simulated_rssi = -85;
static bool simulated_registered = true;

static void _4g_driver_send_at_cmd(const char *cmd)
{
    uint16_t len = strlen(cmd);
    if (len > AT_CMD_BUFFER_SIZE - 4) {
        return;
    }

    snprintf(at_cmd_buffer, sizeof(at_cmd_buffer), "%s\r\n", cmd);
    HAL_UART_Transmit(&LTE_UART, (uint8_t *)at_cmd_buffer, strlen(at_cmd_buffer), 100);
}

static void _4g_driver_parse_csq(const char *response)
{
    int rssi, ber;
    if (sscanf(response, "+CSQ: %d,%d", &rssi, &ber) == 2) {
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

static void _4g_driver_process_response(void)
{
    if (strstr(at_response_buffer, "+CSQ:") != NULL) {
        _4g_driver_parse_csq(at_response_buffer);
    } else if (strstr(at_response_buffer, "+CREG:") != NULL) {
        _4g_driver_parse_creg(at_response_buffer);
    } else if (strstr(at_response_buffer, "OK") != NULL) {
        switch (internal_state) {
            case LTE_STATE_CHECK_MODULE:
                internal_state = LTE_STATE_SET_FUNCTIONALITY;
                _4g_driver_send_at_cmd("AT+CFUN=1");
                break;
            case LTE_STATE_SET_FUNCTIONALITY:
                internal_state = LTE_STATE_CHECK_SIM;
                _4g_driver_send_at_cmd("AT+CPIN?");
                break;
            case LTE_STATE_CHECK_SIM:
                internal_state = LTE_STATE_CHECK_REGISTRATION;
                _4g_driver_send_at_cmd("AT+CREG?");
                break;
            case LTE_STATE_CHECK_REGISTRATION:
                internal_state = LTE_STATE_SET_PDP_CONTEXT;
                _4g_driver_send_at_cmd("AT+CGDCONT=1,\"IP\",\"CMNET\"");
                break;
            case LTE_STATE_SET_PDP_CONTEXT:
                internal_state = LTE_STATE_ACTIVATE_PDP;
                _4g_driver_send_at_cmd("AT+CGACT=1,1");
                break;
            case LTE_STATE_ACTIVATE_PDP:
                internal_state = LTE_STATE_CONNECTED;
                break;
            default:
                break;
        }
    } else if (strstr(at_response_buffer, "ERROR") != NULL) {
        internal_state = LTE_STATE_ERROR;
    }
}

static void _4g_driver_simulate_update(void)
{
    uint32_t now = HAL_GetTick();
    if (now - last_update_time < 1000) {
        return;
    }
    last_update_time = now;

    int16_t delta = (rand() % 11) - 5;
    simulated_rssi = (int8_t)CONSTRAIN(simulated_rssi + delta, LTE_MIN_RSSI, LTE_MAX_RSSI);

    lte_status.rssi = simulated_rssi;
    lte_status.ber = (int8_t)(rand() % 10);
    lte_status.registered = simulated_registered;
    lte_status.network_type = simulated_registered ? NETWORK_TYPE_4G : NETWORK_TYPE_NONE;

    if (internal_state < LTE_STATE_CONNECTED) {
        internal_state = (LTEInternalState)(internal_state + 1);
        if (internal_state > LTE_STATE_CONNECTED) {
            internal_state = LTE_STATE_CONNECTED;
        }
    }
}

bool _4g_driver_init(void)
{
    memset(&lte_status, 0, sizeof(LTEStatus));
    lte_status.rssi = LTE_MIN_RSSI;
    lte_status.ber = 99;
    lte_status.network_type = NETWORK_TYPE_NONE;
    lte_status.registered = false;

    internal_state = LTE_STATE_INIT;
    last_update_time = 0;
    response_index = 0;

    initialized = true;
    return true;
}

void _4g_driver_update(void)
{
    if (!initialized) {
        return;
    }

    _4g_driver_simulate_update();

    switch (internal_state) {
        case LTE_STATE_INIT:
            internal_state = LTE_STATE_CHECK_MODULE;
            break;
        case LTE_STATE_CHECK_MODULE:
        case LTE_STATE_SET_FUNCTIONALITY:
        case LTE_STATE_CHECK_SIM:
        case LTE_STATE_CHECK_REGISTRATION:
        case LTE_STATE_SET_PDP_CONTEXT:
        case LTE_STATE_ACTIVATE_PDP:
            break;
        case LTE_STATE_CONNECTED:
            break;
        case LTE_STATE_ERROR:
            internal_state = LTE_STATE_CHECK_MODULE;
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
    return lte_status.registered && internal_state == LTE_STATE_CONNECTED;
}

int8_t _4g_driver_get_rssi(void)
{
    return lte_status.rssi;
}

void _4g_driver_uart_rx_callback(uint8_t byte)
{
    if (response_index < AT_RESPONSE_BUFFER_SIZE - 1) {
        at_response_buffer[response_index++] = byte;
        at_response_buffer[response_index] = '\0';

        if (byte == '\n' && response_index > 2) {
            _4g_driver_process_response();
            response_index = 0;
            memset(at_response_buffer, 0, sizeof(at_response_buffer));
        }
    } else {
        response_index = 0;
        memset(at_response_buffer, 0, sizeof(at_response_buffer));
    }
}
