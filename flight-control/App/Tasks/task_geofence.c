#include "task_geofence.h"
#include "geofence_manager.h"
#include "flight_controller.h"
#include "sensor_manager.h"
#include "mavlink_handler.h"
#include <string.h>

static TaskHandle_t task_handle = NULL;
static bool armed_blocked = false;
static char block_reason[64] = {0};

void task_geofence_init(void)
{
    geofence_manager_init();
    armed_blocked = false;
    memset(block_reason, 0, sizeof(block_reason));
    
    xTaskCreate(task_geofence_main,
                "Geofence",
                TASK_GEOFENCE_STACK_SIZE,
                NULL,
                TASK_GEOFENCE_PRIORITY,
                &task_handle);
}

static uint8_t determine_action(uint8_t severity, uint8_t fence_action)
{
    if (severity == VIOLATION_SEVERITY_FATAL) {
        return (fence_action == GEOFENCE_ACTION_LAND) ? GEOFENCE_ACTION_LAND : GEOFENCE_ACTION_RTL;
    }
    
    if (severity == VIOLATION_SEVERITY_CRITICAL) {
        return (fence_action == GEOFENCE_ACTION_WARN) ? GEOFENCE_ACTION_HOVER : fence_action;
    }
    
    return fence_action;
}

static void execute_action(uint8_t action)
{
    switch (action) {
        case GEOFENCE_ACTION_HOVER:
            flight_controller_set_mode(FLIGHT_MODE_POS_HOLD);
            break;
            
        case GEOFENCE_ACTION_RTL:
            flight_controller_set_mode(FLIGHT_MODE_RTL);
            break;
            
        case GEOFENCE_ACTION_LAND:
            flight_controller_set_mode(FLIGHT_MODE_LAND);
            break;
            
        case GEOFENCE_ACTION_WARN:
        default:
            break;
    }
}

void task_geofence_main(void *argument)
{
    UNUSED(argument);
    
    TickType_t last_wake_time = xTaskGetTickCount();
    const TickType_t period = pdMS_TO_TICKS(1000 / TASK_GEOFENCE_FREQ);
    
    while (1) {
        vTaskDelayUntil(&last_wake_time, period);
        
        PositionState pos_state;
        sensor_manager_get_position(&pos_state);
        
        if (pos_state.fix_type < 3) {
            continue;
        }
        
        float lat = pos_state.position.lat / 1e7f;
        float lon = pos_state.position.lon / 1e7f;
        float alt = pos_state.altitude;
        
        bool violation = geofence_check_position(lat, lon, alt);
        
        if (violation) {
            GeofenceViolation v;
            geofence_get_last_violation(&v);
            
            uint8_t action = determine_action(v.severity, v.action_taken);
            
            if (flight_controller_is_armed() && !flight_controller_is_failsafe_active()) {
                if (v.severity >= VIOLATION_SEVERITY_CRITICAL) {
                    execute_action(action);
                    flight_controller_trigger_failsafe();
                } else if (v.severity == VIOLATION_SEVERITY_WARNING) {
                    if (v.action_taken == GEOFENCE_ACTION_HOVER ||
                        v.action_taken == GEOFENCE_ACTION_RTL ||
                        v.action_taken == GEOFENCE_ACTION_LAND) {
                        execute_action(action);
                    }
                }
            }
            
            mavlink_send_geofence_violation(v.geofence_id, v.violation_type,
                                            v.severity, v.distance,
                                            v.latitude, v.longitude, v.altitude);
        }
        
        if (!flight_controller_is_armed()) {
            bool can_arm = geofence_check_takeoff(lat, lon, alt);
            if (!can_arm) {
                armed_blocked = true;
                GeofenceViolation v;
                geofence_get_last_violation(&v);
                snprintf(block_reason, sizeof(block_reason),
                        "禁飞区限制: 围栏ID=%d, 类型=%d", v.geofence_id, v.violation_type);
            } else {
                armed_blocked = false;
                memset(block_reason, 0, sizeof(block_reason));
            }
        }
    }
}

bool task_geofence_is_armed_blocked(void)
{
    return armed_blocked;
}

const char* task_geofence_get_block_reason(void)
{
    return block_reason;
}
