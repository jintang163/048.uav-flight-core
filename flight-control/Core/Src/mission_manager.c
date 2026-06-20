#include "mission_manager.h"
#include "coordinate.h"
#include "sensor_manager.h"
#include "flight_controller.h"

static MissionManagerData mission_data;

void mission_manager_init(void)
{
    memset(&mission_data, 0, sizeof(MissionManagerData));
    mission_data.state = MISSION_STATE_IDLE;
    mission_data.plan.count = 0;
    mission_data.plan.current_index = 0;
    mission_data.plan.active = false;
    mission_data.plan.finished = false;
    mission_data.takeoff_complete = false;
}

void mission_manager_update(float dt)
{
    if (!mission_data.plan.active) {
        return;
    }

    GPSPosition current_pos;
    GPSVelocity current_vel;
    sensor_manager_get_gps(&current_pos, &current_vel);

    switch (mission_data.state) {
        case MISSION_STATE_IDLE:
            break;

        case MISSION_STATE_READY:
            if (mission_data.plan.count > 0) {
                mission_data.state = MISSION_STATE_EXECUTING;
                mission_data.plan.current_index = 0;
                mission_manager_goto_waypoint(0);
            }
            break;

        case MISSION_STATE_EXECUTING:
            if (mission_data.plan.current_index < mission_data.plan.count) {
                MissionItem *current_item = &mission_data.plan.items[mission_data.plan.current_index];

                mission_data.target_lat = current_item->lat;
                mission_data.target_lon = current_item->lon;
                mission_data.target_altitude = (float)current_item->alt / 1000.0f;
                mission_data.target_heading = current_item->heading;

                float north, east, down;
                wgs84_to_ned(current_pos.lat, current_pos.lon, current_pos.alt,
                             current_item->lat, current_item->lon, current_item->alt,
                             &north, &east, &down);

                mission_data.distance_to_target = sqrtf(north * north + east * east);
                mission_data.bearing_to_target = atan2f(east, north);

                if (mission_data.distance_to_target < WAYPOINT_ACCEPTANCE_RADIUS) {
                    if (current_item->hold_time > 0) {
                        if (mission_data.hold_start_time == 0) {
                            mission_data.hold_start_time = HAL_GetTick();
                        } else if ((HAL_GetTick() - mission_data.hold_start_time) / 1000 >= current_item->hold_time) {
                            mission_data.hold_start_time = 0;
                            mission_data.plan.current_index++;
                        }
                    } else {
                        mission_data.plan.current_index++;
                    }
                }
            } else {
                mission_data.state = MISSION_STATE_COMPLETED;
                mission_data.plan.finished = true;
                mission_data.plan.active = false;
            }
            break;

        case MISSION_STATE_PAUSED:
            break;

        case MISSION_STATE_COMPLETED:
            break;

        default:
            break;
    }
}

void mission_manager_set_plan(MissionPlan *plan)
{
    mission_data.plan = *plan;
    mission_data.state = MISSION_STATE_READY;
    mission_data.plan.finished = false;
    mission_data.plan.active = false;
}

void mission_manager_get_plan(MissionPlan *plan)
{
    *plan = mission_data.plan;
}

void mission_manager_start(void)
{
    if (mission_data.state == MISSION_STATE_READY || mission_data.state == MISSION_STATE_PAUSED) {
        mission_data.state = MISSION_STATE_EXECUTING;
        mission_data.plan.active = true;
    }
}

void mission_manager_pause(void)
{
    if (mission_data.state == MISSION_STATE_EXECUTING) {
        mission_data.state = MISSION_STATE_PAUSED;
    }
}

void mission_manager_resume(void)
{
    if (mission_data.state == MISSION_STATE_PAUSED) {
        mission_data.state = MISSION_STATE_EXECUTING;
    }
}

void mission_manager_stop(void)
{
    mission_data.state = MISSION_STATE_IDLE;
    mission_data.plan.active = false;
    mission_data.hold_start_time = 0;
}

void mission_manager_reset(void)
{
    mission_data.state = MISSION_STATE_IDLE;
    mission_data.plan.current_index = 0;
    mission_data.plan.active = false;
    mission_data.plan.finished = false;
    mission_data.hold_start_time = 0;
}

void mission_manager_clear_mission(void)
{
    mission_manager_reset();
    mission_data.plan.count = 0;
}

void mission_manager_add_waypoint(MissionItem *item)
{
    if (mission_data.plan.count < 50) {
        mission_data.plan.items[mission_data.plan.count++] = *item;
    }
}

void mission_manager_goto_waypoint(uint16_t index)
{
    if (index < mission_data.plan.count) {
        mission_data.plan.current_index = index;
        mission_data.hold_start_time = 0;

        MissionItem *item = &mission_data.plan.items[index];
        mission_data.target_lat = item->lat;
        mission_data.target_lon = item->lon;
        mission_data.target_altitude = (float)item->alt / 1000.0f;
        mission_data.target_heading = item->heading;
    }
}

void mission_manager_set_home(GPSPosition *home)
{
    MissionItem rtl_item;
    rtl_item.type = MISSION_RETURN;
    rtl_item.lat = home->lat;
    rtl_item.lon = home->lon;
    rtl_item.alt = (int32_t)(RTL_ALTITUDE * 1000);
    rtl_item.heading = 0.0f;
    rtl_item.hold_time = 0;
    rtl_item.radius = LOITER_RADIUS;

    for (int i = 0; i < mission_data.plan.count; i++) {
        if (mission_data.plan.items[i].type == MISSION_RETURN) {
            mission_data.plan.items[i] = rtl_item;
            return;
        }
    }

    if (mission_data.plan.count < 50) {
        mission_data.plan.items[mission_data.plan.count++] = rtl_item;
    }
}

void mission_manager_start_takeoff(float height)
{
    mission_data.target_altitude = height;
    mission_data.takeoff_complete = false;

    GPSPosition current_pos;
    GPSVelocity current_vel;
    sensor_manager_get_gps(&current_pos, &current_vel);

    mission_data.target_lat = current_pos.lat;
    mission_data.target_lon = current_pos.lon;
}

void mission_manager_start_land(void)
{
    flight_controller_set_mode(FLIGHT_MODE_LAND);
}

void mission_manager_start_rtl(void)
{
    flight_controller_set_mode(FLIGHT_MODE_RTL);
}

bool mission_manager_is_active(void)
{
    return mission_data.plan.active;
}

uint16_t mission_manager_get_current_index(void)
{
    return mission_data.plan.current_index;
}

MissionState mission_manager_get_state(void)
{
    return mission_data.state;
}

float mission_manager_get_distance_to_target(void)
{
    return mission_data.distance_to_target;
}

float mission_manager_get_bearing_to_target(void)
{
    return mission_data.bearing_to_target;
}

void mission_manager_get_target_position(int32_t *lat, int32_t *lon, float *alt)
{
    *lat = mission_data.target_lat;
    *lon = mission_data.target_lon;
    *alt = mission_data.target_altitude;
}
