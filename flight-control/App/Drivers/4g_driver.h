#ifndef __4G_DRIVER_H__
#define __4G_DRIVER_H__

#include "types.h"
#include "flight_config.h"

typedef enum {
    NETWORK_TYPE_NONE = 0,
    NETWORK_TYPE_2G = 1,
    NETWORK_TYPE_3G = 2,
    NETWORK_TYPE_4G = 3,
    NETWORK_TYPE_5G = 4
} NetworkType;

typedef struct __attribute__((packed)) {
    int8_t rssi;
    int8_t ber;
    NetworkType network_type;
    bool registered;
} LTEStatus;

bool _4g_driver_init(void);
void _4g_driver_update(void);
bool _4g_driver_get_status(LTEStatus *status);
bool _4g_driver_is_connected(void);
int8_t _4g_driver_get_rssi(void);

#endif
