#include "formation_manager.h"
#include <string.h>
#include <math.h>
#include "stm32f4xx_hal.h"

static FormationData formation_data;
static FormationLightCommand light_cmd;
static float collision_deceleration = 1.0f;

#define UWB_NEIGHBOR_TIMEOUT 1000

static void calculate_formation_line(uint8_t total, float spacing, Vector3f *positions)
{
    float start_offset = -((float)(total - 1) * spacing) / 2.0f;
    for (uint8_t i = 0; i < total; i++) {
        positions[i].x = 0.0f;
        positions[i].y = start_offset + (float)i * spacing;
        positions[i].z = 0.0f;
    }
}

static void calculate_formation_triangle(uint8_t total, float spacing, Vector3f *positions)
{
    if (total == 0) return;
    
    positions[0].x = 0.0f;
    positions[0].y = 0.0f;
    positions[0].z = 0.0f;
    
    if (total == 1) return;
    
    uint8_t row = 1;
    uint8_t count_in_row = 2;
    uint8_t placed = 1;
    
    while (placed < total) {
        float row_width = (float)(count_in_row - 1) * spacing;
        float start_y = -row_width / 2.0f;
        float row_x = -((float)row * spacing * 0.866f);
        
        for (uint8_t i = 0; i < count_in_row && placed < total; i++) {
            positions[placed].x = row_x;
            positions[placed].y = start_y + (float)i * spacing;
            positions[placed].z = 0.0f;
            placed++;
        }
        
        row++;
        count_in_row++;
    }
}

static void calculate_formation_circle(uint8_t total, float spacing, Vector3f *positions)
{
    if (total == 0) return;
    
    float radius;
    if (total == 1) {
        radius = 0.0f;
    } else {
        radius = spacing / (2.0f * sinf(M_PI / (float)total));
    }
    
    for (uint8_t i = 0; i < total; i++) {
        float angle = 2.0f * M_PI * (float)i / (float)total - M_PI / 2.0f;
        positions[i].x = radius * cosf(angle);
        positions[i].y = radius * sinf(angle);
        positions[i].z = 0.0f;
    }
}

void formation_manager_get_formation_position(uint8_t uav_index, FormationType type, float spacing, Vector3f *pos)
{
    Vector3f positions[MAX_FORMATION_UAVS];
    
    switch (type) {
        case FORMATION_LINE:
            calculate_formation_line(formation_data.total_uavs, spacing, positions);
            break;
        case FORMATION_TRIANGLE:
            calculate_formation_triangle(formation_data.total_uavs, spacing, positions);
            break;
        case FORMATION_CIRCLE:
            calculate_formation_circle(formation_data.total_uavs, spacing, positions);
            break;
        default:
            positions[0].x = 0;
            positions[0].y = 0;
            positions[0].z = 0;
            break;
    }
    
    if (uav_index < formation_data.total_uavs) {
        *pos = positions[uav_index];
    } else {
        pos->x = 0;
        pos->y = 0;
        pos->z = 0;
    }
}

void formation_manager_init(void)
{
    memset(&formation_data, 0, sizeof(FormationData));
    
    formation_data.type = FORMATION_LINE;
    formation_data.state = FORMATION_STATE_IDLE;
    formation_data.uav_id = 0;
    formation_data.total_uavs = 1;
    formation_data.spacing = 5.0f;
    formation_data.leader_id = 0;
    formation_data.is_leader = true;
    formation_data.neighbor_count = 0;
    formation_data.collision_warning = false;
    formation_data.min_distance = 0;
    formation_data.closest_uav_id = 0;
    formation_data.synced = false;
    formation_data.sync_timestamp = 0;
    
    memset(&light_cmd, 0, sizeof(FormationLightCommand));
    light_cmd.r = 0;
    light_cmd.g = 255;
    light_cmd.b = 0;
    light_cmd.effect = LIGHT_EFFECT_STATIC;
    
    collision_deceleration = 1.0f;
}

void formation_manager_update(float dt)
{
    uint32_t now = HAL_GetTick();
    
    uint8_t online_count = 0;
    float min_dist = 1e6f;
    uint8_t closest_id = 0;
    bool collision = false;
    
    for (uint8_t i = 0; i < MAX_FORMATION_UAVS; i++) {
        if (formation_data.neighbors[i].online) {
            uint32_t age = now - formation_data.neighbors[i].last_update;
            if (age > UWB_NEIGHBOR_TIMEOUT) {
                formation_data.neighbors[i].online = false;
            } else {
                online_count++;
                
                float dx = formation_data.neighbors[i].relative_pos.x;
                float dy = formation_data.neighbors[i].relative_pos.y;
                float dz = formation_data.neighbors[i].relative_pos.z;
                float dist = sqrtf(dx * dx + dy * dy + dz * dz);
                
                if (dist < min_dist) {
                    min_dist = dist;
                    closest_id = i;
                }
                
                if (dist < COLLISION_WARNING_DISTANCE) {
                    collision = true;
                }
            }
        }
    }
    
    formation_data.neighbor_count = online_count;
    formation_data.min_distance = min_dist;
    formation_data.closest_uav_id = closest_id;
    formation_data.collision_warning = collision;
    
    if (collision) {
        float dist_ratio = min_dist / COLLISION_WARNING_DISTANCE;
        if (dist_ratio < 0.2f) dist_ratio = 0.2f;
        collision_deceleration = dist_ratio * COLLISION_DECELERATION_FACTOR + (1.0f - COLLISION_DECELERATION_FACTOR);
        collision_deceleration = CONSTRAIN(collision_deceleration, 0.2f, 1.0f);
    } else {
        collision_deceleration = 1.0f;
    }
    
    if (formation_data.state == FORMATION_STATE_EXECUTING) {
        Vector3f target_pos;
        formation_manager_get_formation_position(
            formation_data.uav_id,
            formation_data.type,
            formation_data.spacing,
            &target_pos
        );
        formation_data.formation_offset = target_pos;
    }
}

void formation_manager_set_type(FormationType type)
{
    formation_data.type = type;
}

FormationType formation_manager_get_type(void)
{
    return formation_data.type;
}

void formation_manager_set_uav_id(uint8_t id)
{
    formation_data.uav_id = id;
    formation_data.is_leader = (id == formation_data.leader_id);
}

uint8_t formation_manager_get_uav_id(void)
{
    return formation_data.uav_id;
}

void formation_manager_set_total_uavs(uint8_t count)
{
    if (count > MAX_FORMATION_UAVS) {
        count = MAX_FORMATION_UAVS;
    }
    formation_data.total_uavs = count;
}

uint8_t formation_manager_get_total_uavs(void)
{
    return formation_data.total_uavs;
}

void formation_manager_set_spacing(float spacing)
{
    formation_data.spacing = spacing;
}

float formation_manager_get_spacing(void)
{
    return formation_data.spacing;
}

void formation_manager_set_leader(uint8_t leader_id)
{
    formation_data.leader_id = leader_id;
    formation_data.is_leader = (formation_data.uav_id == leader_id);
}

uint8_t formation_manager_get_leader(void)
{
    return formation_data.leader_id;
}

bool formation_manager_is_leader(void)
{
    return formation_data.is_leader;
}

void formation_manager_start(void)
{
    if (formation_data.state == FORMATION_STATE_READY || 
        formation_data.state == FORMATION_STATE_IDLE) {
        formation_data.state = FORMATION_STATE_EXECUTING;
    }
}

void formation_manager_stop(void)
{
    formation_data.state = FORMATION_STATE_IDLE;
    formation_data.synced = false;
}

void formation_manager_pause(void)
{
    if (formation_data.state == FORMATION_STATE_EXECUTING) {
        formation_data.state = FORMATION_STATE_PAUSED;
    }
}

void formation_manager_resume(void)
{
    if (formation_data.state == FORMATION_STATE_PAUSED) {
        formation_data.state = FORMATION_STATE_EXECUTING;
    }
}

FormationState formation_manager_get_state(void)
{
    return formation_data.state;
}

void formation_manager_update_neighbor(uint8_t uav_id, Vector3f *rel_pos, Vector3f *velocity, float yaw)
{
    if (uav_id >= MAX_FORMATION_UAVS) return;
    
    formation_data.neighbors[uav_id].uav_id = uav_id;
    formation_data.neighbors[uav_id].relative_pos = *rel_pos;
    formation_data.neighbors[uav_id].velocity = *velocity;
    formation_data.neighbors[uav_id].yaw = yaw;
    formation_data.neighbors[uav_id].last_update = HAL_GetTick();
    formation_data.neighbors[uav_id].online = true;
}

uint8_t formation_manager_get_neighbor_count(void)
{
    return formation_data.neighbor_count;
}

bool formation_manager_get_neighbor(uint8_t index, UWBNeighborInfo *info)
{
    if (index >= MAX_FORMATION_UAVS || !formation_data.neighbors[index].online) {
        return false;
    }
    *info = formation_data.neighbors[index];
    return true;
}

void formation_manager_get_target_offset(Vector3f *offset)
{
    *offset = formation_data.formation_offset;
}

bool formation_manager_is_collision_warning(void)
{
    return formation_data.collision_warning;
}

float formation_manager_get_min_distance(void)
{
    return formation_data.min_distance;
}

uint8_t formation_manager_get_closest_uav(void)
{
    return formation_data.closest_uav_id;
}

void formation_manager_set_light_command(FormationLightCommand *cmd)
{
    light_cmd = *cmd;
}

void formation_manager_get_light_command(FormationLightCommand *cmd)
{
    *cmd = light_cmd;
}

void formation_manager_sync(uint32_t timestamp)
{
    formation_data.sync_timestamp = timestamp;
    formation_data.synced = true;
}

bool formation_manager_is_synced(void)
{
    return formation_data.synced;
}

float formation_manager_get_collision_deceleration(void)
{
    return collision_deceleration;
}
