#include "geofence_manager.h"
#include <string.h>
#include <math.h>

#ifndef M_PI
#define M_PI 3.14159265358979323846f
#endif

#define DEG2RAD(x) ((x) * (M_PI / 180.0f))
#define EARTH_RADIUS_M 6371000.0f

static GeofenceManager gfm;
static float g_max_altitude = 120.0f;
static float g_max_distance = 500.0f;
static GPSPosition g_home_position;
static bool g_home_set = false;

void geofence_manager_init(void)
{
    memset(&gfm, 0, sizeof(GeofenceManager));
    gfm.custom_count = 0;
    gfm.national_count = 0;
    gfm.has_violation = false;
    gfm.action_triggered = false;
    gfm.current_severity = 0;
    gfm.violation_count = 0;
    gfm.compressed_size = 0;
}

bool geofence_add_custom(Geofence *fence)
{
    if (gfm.custom_count >= GEOFENCE_MAX_COUNT) {
        return false;
    }
    if (fence == NULL) {
        return false;
    }
    
    memcpy(&gfm.custom_fences[gfm.custom_count], fence, sizeof(Geofence));
    gfm.custom_count++;
    return true;
}

bool geofence_remove_custom(uint16_t fence_id)
{
    for (uint16_t i = 0; i < gfm.custom_count; i++) {
        if (gfm.custom_fences[i].id == fence_id) {
            if (i < gfm.custom_count - 1) {
                memmove(&gfm.custom_fences[i], &gfm.custom_fences[i + 1],
                        (gfm.custom_count - i - 1) * sizeof(Geofence));
            }
            gfm.custom_count--;
            memset(&gfm.custom_fences[gfm.custom_count], 0, sizeof(Geofence));
            return true;
        }
    }
    return false;
}

bool geofence_get_custom(uint16_t fence_id, Geofence *fence)
{
    if (fence == NULL) return false;
    
    for (uint16_t i = 0; i < gfm.custom_count; i++) {
        if (gfm.custom_fences[i].id == fence_id) {
            memcpy(fence, &gfm.custom_fences[i], sizeof(Geofence));
            return true;
        }
    }
    return false;
}

uint16_t geofence_custom_count(void)
{
    return gfm.custom_count;
}

bool geofence_set_national_data(uint8_t *data, uint16_t len)
{
    if (data == NULL || len == 0 || len > sizeof(gfm.national_data_compressed)) {
        return false;
    }
    
    memcpy(gfm.national_data_compressed, data, len);
    gfm.compressed_size = len;
    
    uint16_t count = 0;
    uint16_t offset = 0;
    
    while (offset + sizeof(Geofence) <= len && count < GEOFENCE_NATIONAL_MAX) {
        memcpy(&gfm.national_fences[count], &data[offset], sizeof(Geofence));
        offset += sizeof(Geofence);
        count++;
    }
    
    gfm.national_count = count;
    return true;
}

bool geofence_get_national(uint16_t index, Geofence *fence)
{
    if (fence == NULL || index >= gfm.national_count) {
        return false;
    }
    memcpy(fence, &gfm.national_fences[index], sizeof(Geofence));
    return true;
}

uint16_t geofence_national_count(void)
{
    return gfm.national_count;
}

uint16_t geofence_national_compressed_size(void)
{
    return gfm.compressed_size;
}

void geofence_set_max_altitude(float alt)
{
    if (alt > 0) {
        g_max_altitude = alt;
    }
}

void geofence_set_max_distance(float dist)
{
    if (dist > 0) {
        g_max_distance = dist;
    }
}

float geofence_get_max_altitude(void)
{
    return g_max_altitude;
}

float geofence_get_max_distance(void)
{
    return g_max_distance;
}

float geofence_haversine_distance(float lat1, float lon1, float lat2, float lon2)
{
    float dlat = DEG2RAD(lat2 - lat1);
    float dlon = DEG2RAD(lon2 - lon1);
    
    float a = sinf(dlat / 2.0f) * sinf(dlat / 2.0f) +
              cosf(DEG2RAD(lat1)) * cosf(DEG2RAD(lat2)) *
              sinf(dlon / 2.0f) * sinf(dlon / 2.0f);
    
    float c = 2.0f * atan2f(sqrtf(a), sqrtf(1.0f - a));
    return EARTH_RADIUS_M * c;
}

bool geofence_is_point_in_circle(float lat, float lon, float center_lat, float center_lon, float radius)
{
    float dist = geofence_haversine_distance(lat, lon, center_lat, center_lon);
    return dist <= radius;
}

bool geofence_is_point_in_polygon(float lat, float lon, float vertices[][2], uint8_t count)
{
    if (count < 3 || vertices == NULL) {
        return false;
    }
    
    bool inside = false;
    float x = lon;
    float y = lat;
    
    for (uint8_t i = 0, j = count - 1; i < count; j = i++) {
        float xi = vertices[i][1];
        float yi = vertices[i][0];
        float xj = vertices[j][1];
        float yj = vertices[j][0];
        
        bool intersect = ((yi > y) != (yj > y)) &&
                         (x < (xj - xi) * (y - yi) / (yj - yi) + xi);
        
        if (intersect) {
            inside = !inside;
        }
    }
    
    return inside;
}

static uint8_t get_violation_severity(Geofence *fence)
{
    if (fence->category == GEOFENCE_CATEGORY_AIRPORT ||
        fence->category == GEOFENCE_CATEGORY_MILITARY ||
        fence->category == GEOFENCE_CATEGORY_NUCLEAR ||
        fence->category == GEOFENCE_CATEGORY_NATIONAL) {
        return VIOLATION_SEVERITY_CRITICAL;
    }
    
    if (fence->fail_action == GEOFENCE_ACTION_RTL ||
        fence->fail_action == GEOFENCE_ACTION_LAND) {
        return VIOLATION_SEVERITY_FATAL;
    }
    
    return VIOLATION_SEVERITY_WARNING;
}

static void record_violation(Geofence *fence, uint8_t type, float distance,
                              float lat, float lon, float alt)
{
    gfm.last_violation.geofence_id = fence ? fence->id : 0;
    gfm.last_violation.violation_type = type;
    gfm.last_violation.severity = fence ? get_violation_severity(fence) : VIOLATION_SEVERITY_WARNING;
    gfm.last_violation.distance = distance;
    gfm.last_violation.latitude = lat;
    gfm.last_violation.longitude = lon;
    gfm.last_violation.altitude = alt;
    gfm.last_violation.timestamp = 0;
    gfm.last_violation.action_taken = fence ? fence->fail_action : GEOFENCE_ACTION_WARN;
    gfm.has_violation = true;
    gfm.current_severity = gfm.last_violation.severity;
    gfm.violation_count++;
    gfm.action_triggered = false;
}

static bool check_single_geofence(Geofence *fence, float lat, float lon, float alt)
{
    if (fence == NULL || !fence->enabled) {
        return false;
    }
    
    float max_alt = fence->max_altitude > 0 ? fence->max_altitude : g_max_altitude;
    if (alt > max_alt && max_alt > 0) {
        record_violation(fence, VIOLATION_TYPE_ALTITUDE_EXCEEDED,
                        alt - max_alt, lat, lon, alt);
        return true;
    }
    
    if (fence->min_altitude > 0 && alt > 0 && alt < fence->min_altitude) {
        record_violation(fence, VIOLATION_TYPE_ALTITUDE_TOO_LOW,
                        fence->min_altitude - alt, lat, lon, alt);
        return true;
    }
    
    switch (fence->shape) {
        case GEOFENCE_SHAPE_CIRCLE: {
            bool inside = geofence_is_point_in_circle(
                lat, lon, fence->center_lat, fence->center_lon, fence->radius);
            
            if (fence->type == GEOFENCE_TYPE_EXCLUSION && inside) {
                float dist = geofence_haversine_distance(
                    lat, lon, fence->center_lat, fence->center_lon);
                record_violation(fence, VIOLATION_TYPE_INSIDE_EXCLUSION,
                                fence->radius - dist, lat, lon, alt);
                return true;
            }
            
            if (fence->type == GEOFENCE_TYPE_INCLUSION && !inside) {
                float dist = geofence_haversine_distance(
                    lat, lon, fence->center_lat, fence->center_lon);
                record_violation(fence, VIOLATION_TYPE_OUTSIDE_INCLUSION,
                                dist - fence->radius, lat, lon, alt);
                return true;
            }
            break;
        }
        
        case GEOFENCE_SHAPE_POLYGON: {
            bool inside = geofence_is_point_in_polygon(
                lat, lon, fence->vertices, fence->vertex_count);
            
            if (fence->type == GEOFENCE_TYPE_EXCLUSION && inside) {
                float min_dist = 1e9f;
                for (uint8_t i = 0; i < fence->vertex_count; i++) {
                    float d = geofence_haversine_distance(
                        lat, lon, fence->vertices[i][0], fence->vertices[i][1]);
                    if (d < min_dist) min_dist = d;
                }
                record_violation(fence, VIOLATION_TYPE_INSIDE_EXCLUSION,
                                min_dist, lat, lon, alt);
                return true;
            }
            
            if (fence->type == GEOFENCE_TYPE_INCLUSION && !inside) {
                float min_dist = 1e9f;
                for (uint8_t i = 0; i < fence->vertex_count; i++) {
                    float d = geofence_haversine_distance(
                        lat, lon, fence->vertices[i][0], fence->vertices[i][1]);
                    if (d < min_dist) min_dist = d;
                }
                record_violation(fence, VIOLATION_TYPE_OUTSIDE_INCLUSION,
                                min_dist, lat, lon, alt);
                return true;
            }
            break;
        }
        
        default:
            break;
    }
    
    return false;
}

bool geofence_check_position(float lat, float lon, float alt)
{
    gfm.has_violation = false;
    
    if (g_home_set && g_max_distance > 0) {
        float home_lat = g_home_position.lat / 1e7f;
        float home_lon = g_home_position.lon / 1e7f;
        float dist = geofence_haversine_distance(lat, lon, home_lat, home_lon);
        
        if (dist > g_max_distance) {
            Geofence virtual_fence;
            memset(&virtual_fence, 0, sizeof(Geofence));
            virtual_fence.id = 0xFFFF;
            virtual_fence.type = GEOFENCE_TYPE_INCLUSION;
            virtual_fence.shape = GEOFENCE_SHAPE_CIRCLE;
            virtual_fence.category = GEOFENCE_CATEGORY_CUSTOM;
            virtual_fence.fail_action = GEOFENCE_ACTION_HOVER;
            virtual_fence.enabled = true;
            virtual_fence.center_lat = home_lat;
            virtual_fence.center_lon = home_lon;
            virtual_fence.radius = g_max_distance;
            
            record_violation(&virtual_fence, VIOLATION_TYPE_DISTANCE_EXCEEDED,
                            dist - g_max_distance, lat, lon, alt);
            return true;
        }
    }
    
    for (uint16_t i = 0; i < gfm.custom_count; i++) {
        if (check_single_geofence(&gfm.custom_fences[i], lat, lon, alt)) {
            return true;
        }
    }
    
    for (uint16_t i = 0; i < gfm.national_count; i++) {
        if (check_single_geofence(&gfm.national_fences[i], lat, lon, alt)) {
            return true;
        }
    }
    
    gfm.has_violation = false;
    gfm.current_severity = 0;
    return false;
}

bool geofence_check_takeoff(float lat, float lon, float alt)
{
    for (uint16_t i = 0; i < gfm.national_count; i++) {
        Geofence *fence = &gfm.national_fences[i];
        if (!fence->enabled) continue;
        
        if (fence->category == GEOFENCE_CATEGORY_AIRPORT ||
            fence->category == GEOFENCE_CATEGORY_MILITARY ||
            fence->category == GEOFENCE_CATEGORY_NUCLEAR ||
            fence->category == GEOFENCE_CATEGORY_NATIONAL) {
            
            if (fence->type == GEOFENCE_TYPE_EXCLUSION) {
                if (fence->shape == GEOFENCE_SHAPE_CIRCLE) {
                    if (geofence_is_point_in_circle(lat, lon, fence->center_lat,
                                                    fence->center_lon, fence->radius)) {
                        record_violation(fence, VIOLATION_TYPE_INSIDE_EXCLUSION, 0,
                                        lat, lon, alt);
                        return false;
                    }
                } else if (fence->shape == GEOFENCE_SHAPE_POLYGON) {
                    if (geofence_is_point_in_polygon(lat, lon, fence->vertices,
                                                     fence->vertex_count)) {
                        record_violation(fence, VIOLATION_TYPE_INSIDE_EXCLUSION, 0,
                                        lat, lon, alt);
                        return false;
                    }
                }
            }
        }
    }
    
    for (uint16_t i = 0; i < gfm.custom_count; i++) {
        Geofence *fence = &gfm.custom_fences[i];
        if (!fence->enabled) continue;
        
        if (fence->type == GEOFENCE_TYPE_EXCLUSION) {
            if (fence->shape == GEOFENCE_SHAPE_CIRCLE) {
                if (geofence_is_point_in_circle(lat, lon, fence->center_lat,
                                                fence->center_lon, fence->radius)) {
                    return false;
                }
            } else if (fence->shape == GEOFENCE_SHAPE_POLYGON) {
                if (geofence_is_point_in_polygon(lat, lon, fence->vertices,
                                                 fence->vertex_count)) {
                    return false;
                }
            }
        }
    }
    
    return true;
}

bool geofence_has_active_violation(void)
{
    return gfm.has_violation;
}

void geofence_get_last_violation(GeofenceViolation *violation)
{
    if (violation != NULL) {
        memcpy(violation, &gfm.last_violation, sizeof(GeofenceViolation));
    }
}

uint32_t geofence_get_violation_count(void)
{
    return gfm.violation_count;
}

uint8_t geofence_get_current_severity(void)
{
    return gfm.current_severity;
}

void geofence_clear_violation(void)
{
    gfm.has_violation = false;
    gfm.action_triggered = false;
    gfm.current_severity = 0;
    memset(&gfm.last_violation, 0, sizeof(GeofenceViolation));
}

void geofence_enable_all(void)
{
    for (uint16_t i = 0; i < gfm.custom_count; i++) {
        gfm.custom_fences[i].enabled = true;
    }
    for (uint16_t i = 0; i < gfm.national_count; i++) {
        gfm.national_fences[i].enabled = true;
    }
}

void geofence_disable_all(void)
{
    for (uint16_t i = 0; i < gfm.custom_count; i++) {
        gfm.custom_fences[i].enabled = false;
    }
    for (uint16_t i = 0; i < gfm.national_count; i++) {
        gfm.national_fences[i].enabled = false;
    }
}
