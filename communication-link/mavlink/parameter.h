#ifndef MAVLINK_PARAMETER_H
#define MAVLINK_PARAMETER_H

#include "mavlink_v2.h"
#include <stdint.h>
#include <stdbool.h>

#define PARAMETER_NAME_LEN           16
#define PARAMETER_MAX_COUNT          256
#define PARAMETER_TIMEOUT_MS         5000
#define PARAMETER_MAX_RETRIES        3

typedef enum {
    PARAM_TYPE_INT8 = 0,
    PARAM_TYPE_UINT8 = 1,
    PARAM_TYPE_INT16 = 2,
    PARAM_TYPE_UINT16 = 3,
    PARAM_TYPE_INT32 = 4,
    PARAM_TYPE_UINT32 = 5,
    PARAM_TYPE_FLOAT = 9,
    PARAM_TYPE_DOUBLE = 10
} param_type_t;

typedef struct {
    char name[PARAMETER_NAME_LEN];
    param_type_t type;
    union {
        int8_t int8_val;
        uint8_t uint8_val;
        int16_t int16_val;
        uint16_t uint16_val;
        int32_t int32_val;
        uint32_t uint32_val;
        float float_val;
        double double_val;
    } value;
} parameter_t;

typedef void (*param_value_cb_t)(const parameter_t* param, void* user_data);
typedef void (*param_set_result_cb_t)(const char* name, bool success, void* user_data);

typedef struct {
    parameter_t params[PARAMETER_MAX_COUNT];
    uint16_t count;
    uint8_t target_system;
    uint8_t target_component;
    uint64_t last_request_time;
    uint8_t retries;
    uint16_t request_index;
    param_value_cb_t value_cb;
    param_set_result_cb_t set_cb;
    void* user_data;
    bool request_all_in_progress;
    uint32_t request_count;
    uint32_t set_count;
} param_manager_t;

void parameter_init(parameter_t* param, const char* name, param_type_t type);
bool parameter_set_value(parameter_t* param, float value);
float parameter_get_value(const parameter_t* param);

void param_manager_init(param_manager_t* manager, uint8_t target_system, uint8_t target_component);
int param_manager_add(param_manager_t* manager, const parameter_t* param);
parameter_t* param_manager_find(param_manager_t* manager, const char* name);
uint16_t param_manager_get_count(const param_manager_t* manager);
bool param_manager_get_by_index(const param_manager_t* manager, uint16_t index, parameter_t* out);

int param_request_list(param_manager_t* manager, mavlink_encoder_t* encoder,
                       param_value_cb_t callback, void* user_data, uint64_t now_ms);
int param_request_read(param_manager_t* manager, mavlink_encoder_t* encoder,
                       const char* name, int16_t index,
                       param_value_cb_t callback, void* user_data, uint64_t now_ms);
int param_set(param_manager_t* manager, mavlink_encoder_t* encoder,
              const char* name, float value, param_type_t type,
              param_set_result_cb_t callback, void* user_data, uint64_t now_ms);

bool param_process_message(param_manager_t* manager, const mavlink_message_t* msg, uint64_t now_ms);
void param_manager_update(param_manager_t* manager, mavlink_encoder_t* encoder, uint64_t now_ms);

#endif
