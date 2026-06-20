#ifndef __MISSION_MANAGER_H__
#define __MISSION_MANAGER_H__

#include "types.h"
#include "flight_config.h"

typedef enum {
    MISSION_STATE_IDLE = 0,
    MISSION_STATE_READY = 1,
    MISSION_STATE_EXECUTING = 2,
    MISSION_STATE_PAUSED = 3,
    MISSION_STATE_COMPLETED = 4
} MissionState;

typedef struct {
    MissionState state;
    MissionPlan plan;
    float target_altitude;
    float target_heading;
    int32_t target_lat;
    int32_t target_lon;
    float distance_to_target;
    float bearing_to_target;
    uint32_t hold_start_time;
    bool takeoff_complete;
} MissionManagerData;

void mission_manager_init(void);
void mission_manager_update(float dt);
void mission_manager_set_plan(MissionPlan *plan);
void mission_manager_get_plan(MissionPlan *plan);
void mission_manager_start(void);
void mission_manager_pause(void);
void mission_manager_resume(void);
void mission_manager_stop(void);
void mission_manager_reset(void);
void mission_manager_clear_mission(void);
void mission_manager_add_waypoint(MissionItem *item);
void mission_manager_goto_waypoint(uint16_t index);
void mission_manager_set_home(GPSPosition *home);
void mission_manager_start_takeoff(float height);
void mission_manager_start_land(void);
void mission_manager_start_rtl(void);
bool mission_manager_is_active(void);
uint16_t mission_manager_get_current_index(void);
MissionState mission_manager_get_state(void);
float mission_manager_get_distance_to_target(void);
float mission_manager_get_bearing_to_target(void);
void mission_manager_get_target_position(int32_t *lat, int32_t *lon, float *alt);

#endif
