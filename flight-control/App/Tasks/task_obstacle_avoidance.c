#include "task_obstacle_avoidance.h"
#include "flight_controller.h"
#include "sensor_manager.h"
#include "mavlink_handler.h"
#include "blackbox_logger.h"
#include "coordinate.h"
#include <string.h>
#include <math.h>

static TaskHandle_t task_handle = NULL;

static oa_config_t config = {
    .enabled = true,
    .sensitivity = OA_SENSITIVITY_MEDIUM,
    .strategy = OA_STRATEGY_ASCEND_BYPASS,
    .sensor_type = OA_SENSOR_MILLIMETER_WAVE,
    .detection_range = OA_DETECTION_RANGE_MEDIUM,
    .min_obstacle_size = 0.5f,
    .ascend_height = OA_DEFAULT_ASCEND_HEIGHT,
    .retreat_distance = OA_DEFAULT_RETREAT_DIST,
    .bypass_angle = OA_DEFAULT_BYPASS_ANGLE
};

static oa_event_t active_event = {0};
static oa_detection_t current_detections[OA_MAX_DETECTIONS] = {0};
static uint8_t detection_count = 0;

static oa_heatmap_entry_t heatmap[64] = {0};
static uint16_t heatmap_count = 0;

static uint16_t total_detections = 0;
static uint16_t total_events = 0;

static uint8_t bypass_phase = 0;
static uint32_t phase_start_time = 0;
static float hover_stable_alt = 0.0f;
static float lateral_offset_lat = 0.0f;
static float lateral_offset_lng = 0.0f;

#define BYPASS_PHASE_HOVER       0
#define BYPASS_PHASE_ASCEND      1
#define BYPASS_PHASE_LATERAL     2
#define BYPASS_PHASE_FORWARD     3
#define BYPASS_PHASE_DESCEND     4
#define BYPASS_PHASE_RESUME      5

#define HOVER_STABILIZE_MS       500
#define ASCEND_TOLERANCE_M       0.3f
#define LATERAL_DISTANCE_M       5.0f

static const float sensitivity_ranges[] = {
    OA_DETECTION_RANGE_FAR,
    OA_DETECTION_RANGE_MEDIUM,
    OA_DETECTION_RANGE_NEAR
};

static float get_reaction_distance(void)
{
    float range = config.detection_range;
    switch (config.sensitivity) {
        case OA_SENSITIVITY_FAR:    return range * 0.8f;
        case OA_SENSITIVITY_MEDIUM: return range * 0.6f;
        case OA_SENSITIVITY_NEAR:   return range * 0.4f;
        default:                    return range * 0.6f;
    }
}

static void update_detection_range(void)
{
    if (config.sensitivity <= OA_SENSITIVITY_NEAR) {
        config.detection_range = sensitivity_ranges[config.sensitivity];
    }
}

static void read_sensor_data(void)
{
    detection_count = 0;
    memset(current_detections, 0, sizeof(current_detections));

    for (uint8_t i = 0; i < OA_MAX_DETECTIONS; i++) {
        float dist = 0.0f;
        float angle = 0.0f;
        float size = 0.0f;
        float conf = 0.0f;
        oa_direction_t dir = OA_DIRECTION_FRONT;

        if (config.sensor_type == OA_SENSOR_MILLIMETER_WAVE) {
            dist = sensor_manager_get_mmwave_distance(i, &angle, &size, &conf);
        } else if (config.sensor_type == OA_SENSOR_STEREO_VISION) {
            dist = sensor_manager_get_stereo_distance(i, &angle, &size, &conf);
        } else {
            continue;
        }

        if (dist <= 0.0f || dist > config.detection_range || conf < 0.3f) {
            continue;
        }

        if (angle < -60.0f) dir = OA_DIRECTION_LEFT;
        else if (angle > 60.0f) dir = OA_DIRECTION_RIGHT;
        else if (angle > 150.0f || angle < -150.0f) dir = OA_DIRECTION_REAR;
        else dir = OA_DIRECTION_FRONT;

        current_detections[detection_count].distance = dist;
        current_detections[detection_count].relative_angle = angle;
        current_detections[detection_count].obstacle_size = size;
        current_detections[detection_count].confidence = conf;
        current_detections[detection_count].direction = dir;
        current_detections[detection_count].sensor_type = config.sensor_type;
        current_detections[detection_count].timestamp = xTaskGetTickCount() * portTICK_PERIOD_MS;
        detection_count++;
        total_detections++;
    }
}

static void update_heatmap(float lat, float lng, float alt, float distance)
{
    for (uint16_t i = 0; i < heatmap_count; i++) {
        float dlat = fabsf(heatmap[i].lat - lat);
        float dlng = fabsf(heatmap[i].lng - lng);
        if (dlat < 0.00001f && dlng < 0.00001f) {
            heatmap[i].trigger_count++;
            heatmap[i].last_trigger_time = xTaskGetTickCount() * portTICK_PERIOD_MS;
            if (distance < heatmap[i].min_distance) {
                heatmap[i].min_distance = distance;
            }
            return;
        }
    }

    if (heatmap_count < 64) {
        heatmap[heatmap_count].lat = lat;
        heatmap[heatmap_count].lng = lng;
        heatmap[heatmap_count].alt = alt;
        heatmap[heatmap_count].trigger_count = 1;
        heatmap[heatmap_count].last_trigger_time = xTaskGetTickCount() * portTICK_PERIOD_MS;
        heatmap[heatmap_count].min_distance = distance;
        heatmap_count++;
    }
}

static void execute_hover(void)
{
    flight_controller_set_mode(FLIGHT_MODE_POS_HOLD);
}

static void execute_ascend_bypass(float current_alt)
{
    float target_alt = current_alt + config.ascend_height;
    flight_controller_set_target_altitude(target_alt);

    if (active_event.bypass_path_count < 16) {
        PositionState pos;
        sensor_manager_get_position(&pos);
        uint8_t idx = active_event.bypass_path_count;
        active_event.bypass_path[idx].lat = pos.position.lat / 1e7f;
        active_event.bypass_path[idx].lng = pos.position.lon / 1e7f;
        active_event.bypass_path[idx].alt = target_alt;
        active_event.bypass_path[idx].timestamp = xTaskGetTickCount() * portTICK_PERIOD_MS;
        active_event.bypass_path[idx].type = 1;
        active_event.bypass_path_count++;
    }
}

static void execute_retreat_bypass(void)
{
    float heading_rad = flight_controller_get_heading() * DEG_TO_RAD;
    float retreat_north = -config.retreat_distance * cosf(heading_rad);
    float retreat_east = -config.retreat_distance * sinf(heading_rad);

    PositionState pos;
    sensor_manager_get_position(&pos);
    float current_lat = pos.position.lat / 1e7f;
    float current_lng = pos.position.lon / 1e7f;

    float dlat = retreat_north / 111320.0f;
    float dlng = retreat_east / (111320.0f * cosf(current_lat * DEG_TO_RAD));

    float target_lat = current_lat + dlat;
    float target_lng = current_lng + dlng;

    flight_controller_goto_position(target_lat, target_lng, pos.altitude);

    if (active_event.bypass_path_count < 16) {
        uint8_t idx = active_event.bypass_path_count;
        active_event.bypass_path[idx].lat = target_lat;
        active_event.bypass_path[idx].lng = target_lng;
        active_event.bypass_path[idx].alt = pos.altitude;
        active_event.bypass_path[idx].timestamp = xTaskGetTickCount() * portTICK_PERIOD_MS;
        active_event.bypass_path[idx].type = 2;
        active_event.bypass_path_count++;
    }
}

static void handle_obstacle_detected(oa_detection_t *det)
{
    if (!config.enabled) return;
    if (active_event.status == OA_STATUS_AVOIDING || active_event.status == OA_STATUS_BYPASSING) return;

    float reaction_dist = get_reaction_distance();
    if (det->distance > reaction_dist) return;

    active_event.status = OA_STATUS_TRIGGERED;
    active_event.detection = *det;

    PositionState pos;
    sensor_manager_get_position(&pos);
    active_event.start_lat = pos.position.lat / 1e7f;
    active_event.start_lng = pos.position.lon / 1e7f;
    active_event.start_alt = pos.altitude;
    active_event.start_timestamp = xTaskGetTickCount() * portTICK_PERIOD_MS;
    active_event.bypass_path_count = 0;
    active_event.complete_timestamp = 0;

    if (active_event.bypass_path_count < 16) {
        active_event.bypass_path[0].lat = active_event.start_lat;
        active_event.bypass_path[0].lng = active_event.start_lng;
        active_event.bypass_path[0].alt = active_event.start_alt;
        active_event.bypass_path[0].timestamp = active_event.start_timestamp;
        active_event.bypass_path[0].type = 0;
        active_event.bypass_path_count = 1;
    }

    total_events++;

    update_heatmap(active_event.start_lat, active_event.start_lng, active_event.start_alt, det->distance);

    blackbox_log_event(BLACKBOX_EVENT_OBSTACLE_DETECTED,
                       (int32_t)det->direction,
                       (int32_t)(det->distance * 100),
                       (float)config.strategy,
                       0.0f,
                       "OA: obstacle detected");

    mavlink_send_obstacle_avoidance_event(&active_event);
}

static void execute_avoidance(void)
{
    if (active_event.status != OA_STATUS_TRIGGERED &&
        active_event.status != OA_STATUS_AVOIDING &&
        active_event.status != OA_STATUS_BYPASSING) {
        return;
    }

    switch (config.strategy) {
        case OA_STRATEGY_HOVER:
            execute_hover();
            active_event.status = OA_STATUS_AVOIDING;
            bypass_phase = BYPASS_PHASE_HOVER;
            break;

        case OA_STRATEGY_ASCEND_BYPASS:
            active_event.status = OA_STATUS_BYPASSING;
            if (bypass_phase == BYPASS_PHASE_HOVER && phase_start_time == 0) {
                phase_start_time = xTaskGetTickCount() * portTICK_PERIOD_MS;
                execute_hover();
            }
            break;

        case OA_STRATEGY_RETREAT_BYPASS:
            active_event.status = OA_STATUS_BYPASSING;
            if (bypass_phase == BYPASS_PHASE_HOVER && phase_start_time == 0) {
                phase_start_time = xTaskGetTickCount() * portTICK_PERIOD_MS;
                execute_hover();
            }
            break;

        default:
            execute_hover();
            active_event.status = OA_STATUS_AVOIDING;
            break;
    }
}

static void run_ascend_bypass_fsm(void)
{
    PositionState pos;
    sensor_manager_get_position(&pos);
    float current_alt = pos.altitude;
    uint32_t now = xTaskGetTickCount() * portTICK_PERIOD_MS;
    float current_lat = pos.position.lat / 1e7f;
    float current_lng = pos.position.lon / 1e7f;

    switch (bypass_phase) {
        case BYPASS_PHASE_HOVER: {
            execute_hover();
            if (now - phase_start_time >= HOVER_STABILIZE_MS) {
                hover_stable_alt = current_alt;
                bypass_phase = BYPASS_PHASE_ASCEND;
                phase_start_time = now;

                float target_alt = hover_stable_alt + config.ascend_height;
                flight_controller_set_target_altitude(target_alt);

                if (active_event.bypass_path_count < 16) {
                    uint8_t idx = active_event.bypass_path_count;
                    active_event.bypass_path[idx].lat = current_lat;
                    active_event.bypass_path[idx].lng = current_lng;
                    active_event.bypass_path[idx].alt = target_alt;
                    active_event.bypass_path[idx].timestamp = now;
                    active_event.bypass_path[idx].type = 1;
                    active_event.bypass_path_count++;
                }
            }
            break;
        }

        case BYPASS_PHASE_ASCEND: {
            float target_alt = hover_stable_alt + config.ascend_height;
            if (fabsf(current_alt - target_alt) < ASCEND_TOLERANCE_M) {
                bypass_phase = BYPASS_PHASE_LATERAL;
                phase_start_time = now;

                float bypass_angle_rad = config.bypass_angle * DEG_TO_RAD;
                float heading_rad = flight_controller_get_heading();

                float lateral_angle = heading_rad + bypass_angle_rad;
                float lateral_north = LATERAL_DISTANCE_M * cosf(lateral_angle);
                float lateral_east = LATERAL_DISTANCE_M * sinf(lateral_angle);

                lateral_offset_lat = current_lat + lateral_north / 111320.0f;
                lateral_offset_lng = current_lng + lateral_east / (111320.0f * cosf(current_lat * DEG_TO_RAD));

                flight_controller_goto_position(lateral_offset_lat, lateral_offset_lng, target_alt);

                if (active_event.bypass_path_count < 16) {
                    uint8_t idx = active_event.bypass_path_count;
                    active_event.bypass_path[idx].lat = lateral_offset_lat;
                    active_event.bypass_path[idx].lng = lateral_offset_lng;
                    active_event.bypass_path[idx].alt = target_alt;
                    active_event.bypass_path[idx].timestamp = now;
                    active_event.bypass_path[idx].type = 1;
                    active_event.bypass_path_count++;
                }
            } else {
                flight_controller_set_target_altitude(target_alt);
            }
            break;
        }

        case BYPASS_PHASE_LATERAL: {
            float dlat = (current_lat - lateral_offset_lat) * 111320.0f;
            float dlng = (current_lng - lateral_offset_lng) * 111320.0f * cosf(current_lat * DEG_TO_RAD);
            float dist_to_target = sqrtf(dlat * dlat + dlng * dlng);

            if (dist_to_target < 1.0f) {
                bypass_phase = BYPASS_PHASE_FORWARD;
                phase_start_time = now;

                float heading_rad = flight_controller_get_heading();
                float forward_north = config.detection_range * cosf(heading_rad);
                float forward_east = config.detection_range * sinf(heading_rad);

                float forward_lat = current_lat + forward_north / 111320.0f;
                float forward_lng = current_lng + forward_east / (111320.0f * cosf(current_lat * DEG_TO_RAD));

                flight_controller_goto_position(forward_lat, forward_lng, current_alt);

                if (active_event.bypass_path_count < 16) {
                    uint8_t idx = active_event.bypass_path_count;
                    active_event.bypass_path[idx].lat = forward_lat;
                    active_event.bypass_path[idx].lng = forward_lng;
                    active_event.bypass_path[idx].alt = current_alt;
                    active_event.bypass_path[idx].timestamp = now;
                    active_event.bypass_path[idx].type = 1;
                    active_event.bypass_path_count++;
                }
            }
            break;
        }

        case BYPASS_PHASE_FORWARD: {
            bool obstacle_cleared = true;
            for (uint8_t i = 0; i < detection_count; i++) {
                if (current_detections[i].distance < config.detection_range * 0.5f) {
                    obstacle_cleared = false;
                    break;
                }
            }

            if (obstacle_cleared) {
                bypass_phase = BYPASS_PHASE_DESCEND;
                phase_start_time = now;

                flight_controller_set_target_altitude(hover_stable_alt);

                if (active_event.bypass_path_count < 16) {
                    uint8_t idx = active_event.bypass_path_count;
                    active_event.bypass_path[idx].lat = current_lat;
                    active_event.bypass_path[idx].lng = current_lng;
                    active_event.bypass_path[idx].alt = hover_stable_alt;
                    active_event.bypass_path[idx].timestamp = now;
                    active_event.bypass_path[idx].type = 2;
                    active_event.bypass_path_count++;
                }
            }
            break;
        }

        case BYPASS_PHASE_DESCEND: {
            if (fabsf(current_alt - hover_stable_alt) < ASCEND_TOLERANCE_M) {
                bypass_phase = BYPASS_PHASE_RESUME;
                phase_start_time = now;
            }
            break;
        }

        case BYPASS_PHASE_RESUME: {
            FlightMode prev_mode = flight_controller_get_mode();
            flight_controller_set_mode(FLIGHT_MODE_AUTO);
            break;
        }

        default:
            break;
    }
}

static void run_retreat_bypass_fsm(void)
{
    PositionState pos;
    sensor_manager_get_position(&pos);
    uint32_t now = xTaskGetTickCount() * portTICK_PERIOD_MS;
    float current_lat = pos.position.lat / 1e7f;
    float current_lng = pos.position.lon / 1e7f;

    switch (bypass_phase) {
        case BYPASS_PHASE_HOVER: {
            execute_hover();
            if (now - phase_start_time >= HOVER_STABILIZE_MS) {
                bypass_phase = BYPASS_PHASE_LATERAL;
                phase_start_time = now;
                execute_retreat_bypass();
            }
            break;
        }

        default: {
            bool obstacle_cleared = true;
            for (uint8_t i = 0; i < detection_count; i++) {
                if (current_detections[i].distance < config.detection_range * 0.5f) {
                    obstacle_cleared = false;
                    break;
                }
            }
            if (obstacle_cleared) {
                flight_controller_set_mode(FLIGHT_MODE_AUTO);
            }
            break;
        }
    }
}

static bool check_avoidance_complete(void)
{
    if (active_event.status != OA_STATUS_AVOIDING && active_event.status != OA_STATUS_BYPASSING) {
        return false;
    }

    bool obstacle_cleared = true;
    for (uint8_t i = 0; i < detection_count; i++) {
        if (current_detections[i].distance < config.detection_range * 0.5f) {
            obstacle_cleared = false;
            break;
        }
    }

    if (obstacle_cleared) {
        active_event.status = OA_STATUS_COMPLETED;
        active_event.complete_timestamp = xTaskGetTickCount() * portTICK_PERIOD_MS;

        blackbox_log_event(BLACKBOX_EVENT_OBSTACLE_CLEARED,
                           0,
                           (int32_t)(active_event.complete_timestamp - active_event.start_timestamp),
                           0.0f,
                           0.0f,
                           "OA: avoidance completed");

        mavlink_send_obstacle_avoidance_complete(&active_event);
        return true;
    }

    uint32_t elapsed = (xTaskGetTickCount() * portTICK_PERIOD_MS) - active_event.start_timestamp;
    if (elapsed > 30000) {
        active_event.status = OA_STATUS_FAILED;
        active_event.complete_timestamp = xTaskGetTickCount() * portTICK_PERIOD_MS;

        blackbox_log_event(BLACKBOX_EVENT_OBSTACLE_CLEARED,
                           1,
                           (int32_t)elapsed,
                           0.0f,
                           0.0f,
                           "OA: avoidance timeout");

        mavlink_send_obstacle_avoidance_failed(&active_event, "timeout");
        return true;
    }

    return false;
}

void task_obstacle_avoidance_init(void)
{
    memset(&active_event, 0, sizeof(active_event));
    active_event.status = OA_STATUS_IDLE;
    detection_count = 0;
    heatmap_count = 0;
    total_detections = 0;
    total_events = 0;

    xTaskCreate(task_obstacle_avoidance_main,
                "ObstAvoid",
                OA_TASK_STACK_SIZE,
                NULL,
                OA_TASK_PRIORITY,
                &task_handle);
}

void task_obstacle_avoidance_main(void *argument)
{
    UNUSED(argument);

    TickType_t last_wake_time = xTaskGetTickCount();
    const TickType_t period = pdMS_TO_TICKS(1000 / OA_TASK_FREQ);

    while (1) {
        vTaskDelayUntil(&last_wake_time, period);

        if (!config.enabled) {
            continue;
        }

        read_sensor_data();

        for (uint8_t i = 0; i < detection_count; i++) {
            handle_obstacle_detected(&current_detections[i]);
        }

        if (active_event.status == OA_STATUS_TRIGGERED) {
            bypass_phase = BYPASS_PHASE_HOVER;
            phase_start_time = 0;
            execute_avoidance();
        }

        if (active_event.status == OA_STATUS_AVOIDING || active_event.status == OA_STATUS_BYPASSING) {
            if (!check_avoidance_complete()) {
                if (active_event.status == OA_STATUS_BYPASSING) {
                    switch (config.strategy) {
                        case OA_STRATEGY_ASCEND_BYPASS:
                            run_ascend_bypass_fsm();
                            break;
                        case OA_STRATEGY_RETREAT_BYPASS:
                            run_retreat_bypass_fsm();
                            break;
                        default:
                            break;
                    }
                }
            }
        }
    }
}

oa_config_t* oa_get_config(void)
{
    return &config;
}

void oa_set_enabled(bool enabled)
{
    config.enabled = enabled;
    if (!enabled && active_event.status != OA_STATUS_IDLE) {
        active_event.status = OA_STATUS_FAILED;
        active_event.complete_timestamp = xTaskGetTickCount() * portTICK_PERIOD_MS;
    }
}

void oa_set_sensitivity(oa_sensitivity_t sensitivity)
{
    config.sensitivity = sensitivity;
    update_detection_range();
}

void oa_set_strategy(oa_strategy_t strategy)
{
    config.strategy = strategy;
}

void oa_set_detection_range(float range)
{
    config.detection_range = range;
}

void oa_set_ascend_height(float height)
{
    config.ascend_height = height;
}

void oa_set_retreat_distance(float distance)
{
    config.retreat_distance = distance;
}

void oa_set_bypass_angle(float angle)
{
    config.bypass_angle = angle;
}

oa_status_t oa_get_status(void)
{
    return active_event.status;
}

oa_event_t* oa_get_active_event(void)
{
    return &active_event;
}

uint16_t oa_get_total_detections(void)
{
    return total_detections;
}

uint16_t oa_get_total_events(void)
{
    return total_events;
}

void oa_get_heatmap_entries(oa_heatmap_entry_t *entries, uint16_t max_count, uint16_t *out_count)
{
    uint16_t count = (heatmap_count < max_count) ? heatmap_count : max_count;
    memcpy(entries, heatmap, count * sizeof(oa_heatmap_entry_t));
    *out_count = count;
}

void oa_clear_heatmap(void)
{
    memset(heatmap, 0, sizeof(heatmap));
    heatmap_count = 0;
}
