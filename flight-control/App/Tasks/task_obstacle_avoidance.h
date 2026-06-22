#ifndef __TASK_OBSTACLE_AVOIDANCE_H__
#define __TASK_OBSTACLE_AVOIDANCE_H__

#include "main.h"

#define OA_TASK_FREQ                20
#define OA_TASK_STACK_SIZE          1536
#define OA_TASK_PRIORITY            (tskIDLE_PRIORITY + 3)

#define OA_MAX_DETECTIONS           8
#define OA_DETECTION_RANGE_FAR      15.0f
#define OA_DETECTION_RANGE_MEDIUM   10.0f
#define OA_DETECTION_RANGE_NEAR     5.0f
#define OA_DEFAULT_ASCEND_HEIGHT    5.0f
#define OA_DEFAULT_RETREAT_DIST     10.0f
#define OA_DEFAULT_BYPASS_ANGLE     45.0f

typedef enum {
    OA_SENSOR_MILLIMETER_WAVE = 0,
    OA_SENSOR_STEREO_VISION   = 1,
    OA_SENSOR_LIDAR           = 2,
    OA_SENSOR_ULTRASONIC      = 3
} oa_sensor_type_t;

typedef enum {
    OA_SENSITIVITY_FAR    = 0,
    OA_SENSITIVITY_MEDIUM = 1,
    OA_SENSITIVITY_NEAR   = 2
} oa_sensitivity_t;

typedef enum {
    OA_STRATEGY_HOVER         = 0,
    OA_STRATEGY_ASCEND_BYPASS = 1,
    OA_STRATEGY_RETREAT_BYPASS = 2
} oa_strategy_t;

typedef enum {
    OA_DIRECTION_FRONT  = 0,
    OA_DIRECTION_LEFT   = 1,
    OA_DIRECTION_RIGHT  = 2,
    OA_DIRECTION_TOP    = 3,
    OA_DIRECTION_BOTTOM = 4,
    OA_DIRECTION_REAR   = 5
} oa_direction_t;

typedef enum {
    OA_STATUS_IDLE      = 0,
    OA_STATUS_DETECTING = 1,
    OA_STATUS_TRIGGERED = 2,
    OA_STATUS_AVOIDING  = 3,
    OA_STATUS_BYPASSING = 4,
    OA_STATUS_COMPLETED = 5,
    OA_STATUS_FAILED    = 6
} oa_status_t;

typedef struct {
    float distance;
    float relative_angle;
    float obstacle_size;
    float confidence;
    oa_direction_t direction;
    oa_sensor_type_t sensor_type;
    uint32_t timestamp;
} oa_detection_t;

typedef struct {
    float lat;
    float lng;
    float alt;
    uint32_t timestamp;
    uint8_t type;
} oa_bypass_waypoint_t;

typedef struct {
    bool enabled;
    oa_sensitivity_t sensitivity;
    oa_strategy_t strategy;
    oa_sensor_type_t sensor_type;
    float detection_range;
    float min_obstacle_size;
    float ascend_height;
    float retreat_distance;
    float bypass_angle;
} oa_config_t;

typedef struct {
    oa_status_t status;
    oa_detection_t detection;
    oa_bypass_waypoint_t bypass_path[16];
    uint8_t bypass_path_count;
    uint8_t current_bypass_index;
    float start_lat;
    float start_lng;
    float start_alt;
    uint32_t start_timestamp;
    uint32_t complete_timestamp;
} oa_event_t;

typedef struct {
    float lat;
    float lng;
    float alt;
    uint32_t last_trigger_time;
    uint16_t trigger_count;
    float min_distance;
} oa_heatmap_entry_t;

void task_obstacle_avoidance_init(void);
void task_obstacle_avoidance_main(void *argument);

oa_config_t* oa_get_config(void);
void oa_set_enabled(bool enabled);
void oa_set_sensitivity(oa_sensitivity_t sensitivity);
void oa_set_strategy(oa_strategy_t strategy);
void oa_set_detection_range(float range);
void oa_set_ascend_height(float height);
void oa_set_retreat_distance(float distance);
void oa_set_bypass_angle(float angle);

oa_status_t oa_get_status(void);
oa_event_t* oa_get_active_event(void);
uint16_t oa_get_total_detections(void);
uint16_t oa_get_total_events(void);

void oa_get_heatmap_entries(oa_heatmap_entry_t *entries, uint16_t max_count, uint16_t *out_count);
void oa_clear_heatmap(void);

#endif
