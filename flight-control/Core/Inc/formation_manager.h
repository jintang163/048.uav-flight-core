#ifndef __FORMATION_MANAGER_H__
#define __FORMATION_MANAGER_H__

#include "types.h"
#include "flight_config.h"

void formation_manager_init(void);
void formation_manager_update(float dt);

void formation_manager_set_type(FormationType type);
FormationType formation_manager_get_type(void);

void formation_manager_set_uav_id(uint8_t id);
uint8_t formation_manager_get_uav_id(void);

void formation_manager_set_total_uavs(uint8_t count);
uint8_t formation_manager_get_total_uavs(void);

void formation_manager_set_spacing(float spacing);
float formation_manager_get_spacing(void);

void formation_manager_set_leader(uint8_t leader_id);
uint8_t formation_manager_get_leader(void);
bool formation_manager_is_leader(void);

void formation_manager_start(void);
void formation_manager_stop(void);
void formation_manager_pause(void);
void formation_manager_resume(void);

FormationState formation_manager_get_state(void);

void formation_manager_update_neighbor(uint8_t uav_id, Vector3f *rel_pos, Vector3f *velocity, float yaw);
uint8_t formation_manager_get_neighbor_count(void);
bool formation_manager_get_neighbor(uint8_t index, UWBNeighborInfo *info);

void formation_manager_get_target_offset(Vector3f *offset);

bool formation_manager_is_collision_warning(void);
float formation_manager_get_min_distance(void);
uint8_t formation_manager_get_closest_uav(void);

void formation_manager_get_formation_position(uint8_t uav_index, FormationType type, float spacing, Vector3f *pos);

void formation_manager_set_light_command(FormationLightCommand *cmd);
void formation_manager_get_light_command(FormationLightCommand *cmd);

void formation_manager_sync(uint32_t timestamp);
bool formation_manager_is_synced(void);

float formation_manager_get_collision_deceleration(void);

#endif
