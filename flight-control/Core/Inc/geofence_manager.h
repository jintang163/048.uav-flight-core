#ifndef __GEOFENCE_MANAGER_H__
#define __GEOFENCE_MANAGER_H__

#include "types.h"
#include <stdint.h>
#include <stdbool.h>

#define GEOFENCE_MAX_COUNT        32
#define GEOFENCE_MAX_VERTICES     16
#define GEOFENCE_NATIONAL_MAX     200

#define GEOFENCE_ACTION_NONE      0
#define GEOFENCE_ACTION_WARN      1
#define GEOFENCE_ACTION_HOVER     2
#define GEOFENCE_ACTION_RTL       3
#define GEOFENCE_ACTION_LAND      4

#define GEOFENCE_TYPE_INCLUSION   0
#define GEOFENCE_TYPE_EXCLUSION   1

#define GEOFENCE_SHAPE_CIRCLE     0
#define GEOFENCE_SHAPE_POLYGON    1

#define GEOFENCE_CATEGORY_CUSTOM   0
#define GEOFENCE_CATEGORY_AIRPORT  1
#define GEOFENCE_CATEGORY_MILITARY 2
#define GEOFENCE_CATEGORY_NUCLEAR  3
#define GEOFENCE_CATEGORY_GOV      4
#define GEOFENCE_CATEGORY_NATIONAL 5

#define VIOLATION_TYPE_NONE                0
#define VIOLATION_TYPE_ALTITUDE_EXCEEDED   1
#define VIOLATION_TYPE_ALTITUDE_TOO_LOW    2
#define VIOLATION_TYPE_INSIDE_EXCLUSION    3
#define VIOLATION_TYPE_OUTSIDE_INCLUSION   4
#define VIOLATION_TYPE_DISTANCE_EXCEEDED   5

#define VIOLATION_SEVERITY_WARNING   0
#define VIOLATION_SEVERITY_CRITICAL  1
#define VIOLATION_SEVERITY_FATAL     2

typedef struct {
    uint16_t id;
    uint8_t  type;
    uint8_t  shape;
    uint8_t  category;
    uint8_t  fail_action;
    bool     enabled;
    float    max_altitude;
    float    min_altitude;
    float    max_distance;
    float    center_lat;
    float    center_lon;
    float    radius;
    uint8_t  vertex_count;
    float    vertices[GEOFENCE_MAX_VERTICES][2];
} Geofence;

typedef struct {
    uint16_t geofence_id;
    uint8_t  violation_type;
    uint8_t  severity;
    float    distance;
    float    latitude;
    float    longitude;
    float    altitude;
    uint32_t timestamp;
    uint8_t  action_taken;
} GeofenceViolation;

typedef struct {
    Geofence custom_fences[GEOFENCE_MAX_COUNT];
    uint16_t custom_count;
    Geofence national_fences[GEOFENCE_NATIONAL_MAX];
    uint16_t national_count;
    GeofenceViolation last_violation;
    bool     has_violation;
    bool     action_triggered;
    uint8_t  current_severity;
    uint32_t last_check_time;
    uint32_t violation_count;
    uint32_t compressed_size;
    uint8_t  national_data_compressed[4096];
} GeofenceManager;

void geofence_manager_init(void);

bool geofence_add_custom(Geofence *fence);
bool geofence_remove_custom(uint16_t fence_id);
bool geofence_get_custom(uint16_t fence_id, Geofence *fence);
uint16_t geofence_custom_count(void);

bool geofence_set_national_data(uint8_t *data, uint16_t len);
bool geofence_get_national(uint16_t index, Geofence *fence);
uint16_t geofence_national_count(void);
uint16_t geofence_national_compressed_size(void);

void geofence_set_max_altitude(float alt);
void geofence_set_max_distance(float dist);
float geofence_get_max_altitude(void);
float geofence_get_max_distance(void);

bool geofence_check_position(float lat, float lon, float alt);
bool geofence_check_takeoff(float lat, float lon, float alt);

bool geofence_has_active_violation(void);
void geofence_get_last_violation(GeofenceViolation *violation);
uint32_t geofence_get_violation_count(void);
uint8_t geofence_get_current_severity(void);

void geofence_clear_violation(void);
void geofence_enable_all(void);
void geofence_disable_all(void);

bool geofence_is_point_in_circle(float lat, float lon, float center_lat, float center_lon, float radius);
bool geofence_is_point_in_polygon(float lat, float lon, float vertices[][2], uint8_t count);
float geofence_haversine_distance(float lat1, float lon1, float lat2, float lon2);

#endif
