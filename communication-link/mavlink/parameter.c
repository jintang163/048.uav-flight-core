#include "parameter.h"
#include "encoder.h"
#include <string.h>
#include <stdlib.h>

void parameter_init(parameter_t* param, const char* name, param_type_t type)
{
    memset(param, 0, sizeof(parameter_t));
    strncpy(param->name, name, PARAMETER_NAME_LEN - 1);
    param->name[PARAMETER_NAME_LEN - 1] = '\0';
    param->type = type;
}

bool parameter_set_value(parameter_t* param, float value)
{
    switch (param->type) {
    case PARAM_TYPE_INT8:
        param->value.int8_val = (int8_t)value;
        break;
    case PARAM_TYPE_UINT8:
        param->value.uint8_val = (uint8_t)value;
        break;
    case PARAM_TYPE_INT16:
        param->value.int16_val = (int16_t)value;
        break;
    case PARAM_TYPE_UINT16:
        param->value.uint16_val = (uint16_t)value;
        break;
    case PARAM_TYPE_INT32:
        param->value.int32_val = (int32_t)value;
        break;
    case PARAM_TYPE_UINT32:
        param->value.uint32_val = (uint32_t)value;
        break;
    case PARAM_TYPE_FLOAT:
        param->value.float_val = value;
        break;
    case PARAM_TYPE_DOUBLE:
        param->value.double_val = value;
        break;
    default:
        return false;
    }
    return true;
}

float parameter_get_value(const parameter_t* param)
{
    switch (param->type) {
    case PARAM_TYPE_INT8:    return (float)param->value.int8_val;
    case PARAM_TYPE_UINT8:   return (float)param->value.uint8_val;
    case PARAM_TYPE_INT16:   return (float)param->value.int16_val;
    case PARAM_TYPE_UINT16:  return (float)param->value.uint16_val;
    case PARAM_TYPE_INT32:   return (float)param->value.int32_val;
    case PARAM_TYPE_UINT32:  return (float)param->value.uint32_val;
    case PARAM_TYPE_FLOAT:   return param->value.float_val;
    case PARAM_TYPE_DOUBLE:  return (float)param->value.double_val;
    default:                 return 0.0f;
    }
}

void param_manager_init(param_manager_t* manager, uint8_t target_system, uint8_t target_component)
{
    memset(manager, 0, sizeof(param_manager_t));
    manager->target_system = target_system;
    manager->target_component = target_component;
    manager->count = 0;
    manager->request_all_in_progress = false;
}

int param_manager_add(param_manager_t* manager, const parameter_t* param)
{
    if (manager->count >= PARAMETER_MAX_COUNT) {
        return -1;
    }
    memcpy(&manager->params[manager->count], param, sizeof(parameter_t));
    manager->count++;
    return 0;
}

parameter_t* param_manager_find(param_manager_t* manager, const char* name)
{
    for (uint16_t i = 0; i < manager->count; i++) {
        if (strncmp(manager->params[i].name, name, PARAMETER_NAME_LEN) == 0) {
            return &manager->params[i];
        }
    }
    return NULL;
}

uint16_t param_manager_get_count(const param_manager_t* manager) { return manager->count; }

bool param_manager_get_by_index(const param_manager_t* manager, uint16_t index, parameter_t* out)
{
    if (index >= manager->count) {
        return false;
    }
    memcpy(out, &manager->params[index], sizeof(parameter_t));
    return true;
}

int param_request_list(param_manager_t* manager, mavlink_encoder_t* encoder,
                       param_value_cb_t callback, void* user_data, uint64_t now_ms)
{
    (void)encoder;
    manager->request_all_in_progress = true;
    manager->request_index = 0;
    manager->retries = 0;
    manager->last_request_time = now_ms;
    manager->value_cb = callback;
    manager->user_data = user_data;
    manager->request_count++;
    return 0;
}

int param_request_read(param_manager_t* manager, mavlink_encoder_t* encoder,
                       const char* name, int16_t index,
                       param_value_cb_t callback, void* user_data, uint64_t now_ms)
{
    (void)encoder;
    (void)name;
    (void)index;
    manager->retries = 0;
    manager->last_request_time = now_ms;
    manager->value_cb = callback;
    manager->user_data = user_data;
    manager->request_count++;
    return 0;
}

int param_set(param_manager_t* manager, mavlink_encoder_t* encoder,
              const char* name, float value, param_type_t type,
              param_set_result_cb_t callback, void* user_data, uint64_t now_ms)
{
    (void)encoder;
    parameter_t* param = param_manager_find(manager, name);
    if (!param) {
        parameter_t new_param;
        parameter_init(&new_param, name, type);
        parameter_set_value(&new_param, value);
        if (param_manager_add(manager, &new_param) < 0) {
            return -1;
        }
    } else {
        parameter_set_value(param, value);
    }
    manager->retries = 0;
    manager->last_request_time = now_ms;
    manager->set_cb = callback;
    manager->user_data = user_data;
    manager->set_count++;
    if (callback) {
        callback(name, true, user_data);
    }
    return 0;
}

bool param_process_message(param_manager_t* manager, const mavlink_message_t* msg, uint64_t now_ms)
{
    (void)manager;
    (void)msg;
    (void)now_ms;
    return false;
}

void param_manager_update(param_manager_t* manager, mavlink_encoder_t* encoder, uint64_t now_ms)
{
    (void)manager;
    (void)encoder;
    (void)now_ms;
}
