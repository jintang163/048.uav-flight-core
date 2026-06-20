#include "task_mission.h"
#include "mission_manager.h"
#include "flight_controller.h"
#include "sensor_manager.h"
#include "coordinate.h"

static TaskHandle_t task_handle = NULL;

void task_mission_init(void)
{
    mission_manager_init();

    xTaskCreate(task_mission_main,
                "Mission",
                TASK_MISSION_STACK_SIZE,
                NULL,
                TASK_MISSION_PRIORITY,
                &task_handle);
}

void task_mission_main(void *argument)
{
    UNUSED(argument);

    TickType_t last_wake_time = xTaskGetTickCount();
    const TickType_t period = pdMS_TO_TICKS(1000 / TASK_MISSION_FREQ);
    float dt = 1.0f / TASK_MISSION_FREQ;

    while (1) {
        GPSPosition gps_pos;
        GPSVelocity gps_vel;
        sensor_manager_get_gps(&gps_pos, &gps_vel);

        if (flight_controller_is_home_set() == false && gps_pos.lat != 0 && gps_pos.lon != 0) {
            flight_controller_set_home_position(&gps_pos);
            mission_manager_set_home(&gps_pos);
        }

        FlightMode mode = flight_controller_get_mode();
        if (mode == FLIGHT_MODE_AUTO) {
            mission_manager_update(dt);

            if (mission_manager_is_active()) {
                ControlCommand mission_cmd;
                int32_t target_lat, target_lon;
                float target_alt;

                mission_manager_get_target_position(&target_lat, &target_lon, &target_alt);

                float north, east, down;
                wgs84_to_ned(gps_pos.lat, gps_pos.lon, gps_pos.alt,
                             target_lat, target_lon, (int32_t)(target_alt * 1000),
                             &north, &east, &down);

                float dist = sqrtf(north * north + east * east);
                float bearing = atan2f(east, north);

                float max_vel = 5.0f;
                float vel_scale = (dist < 5.0f) ? (dist / 5.0f) : 1.0f;
                float target_vel_n = cosf(bearing) * max_vel * vel_scale;
                float target_vel_e = sinf(bearing) * max_vel * vel_scale;

                float vel_error_n = target_vel_n - gps_vel.vn;
                float vel_error_e = target_vel_e - gps_vel.ve;

                mission_cmd.roll = CONSTRAIN(vel_error_e * 0.5f, -MAX_TILT_ANGLE, MAX_TILT_ANGLE);
                mission_cmd.pitch = CONSTRAIN(-vel_error_n * 0.5f, -MAX_TILT_ANGLE, MAX_TILT_ANGLE);
                mission_cmd.yaw = bearing;
                mission_cmd.throttle = 0.5f;

                float alt_error = target_alt - (float)gps_pos.alt / 1000.0f;
                if (alt_error > 0) {
                    mission_cmd.throttle = CONSTRAIN(0.5f + alt_error * 0.1f, 0.3f, 0.8f);
                } else {
                    mission_cmd.throttle = CONSTRAIN(0.5f + alt_error * 0.1f, 0.2f, 0.7f);
                }

                flight_controller_set_mavlink_command(&mission_cmd);
            }
        } else if (mode == FLIGHT_MODE_RTL) {
            GPSPosition home_pos;
            flight_controller_get_home_position(&home_pos);

            float north, east, down;
            wgs84_to_ned(gps_pos.lat, gps_pos.lon, gps_pos.alt,
                         home_pos.lat, home_pos.lon, (int32_t)(RTL_ALTITUDE * 1000),
                         &north, &east, &down);

            float dist = sqrtf(north * north + east * east);

            ControlCommand rtl_cmd;
            if (dist > WAYPOINT_ACCEPTANCE_RADIUS) {
                float bearing = atan2f(east, north);
                float max_vel = 3.0f;
                float vel_scale = (dist < 5.0f) ? (dist / 5.0f) : 1.0f;
                float target_vel_n = cosf(bearing) * max_vel * vel_scale;
                float target_vel_e = sinf(bearing) * max_vel * vel_scale;

                float vel_error_n = target_vel_n - gps_vel.vn;
                float vel_error_e = target_vel_e - gps_vel.ve;

                rtl_cmd.roll = CONSTRAIN(vel_error_e * 0.5f, -MAX_TILT_ANGLE, MAX_TILT_ANGLE);
                rtl_cmd.pitch = CONSTRAIN(-vel_error_n * 0.5f, -MAX_TILT_ANGLE, MAX_TILT_ANGLE);
                rtl_cmd.yaw = bearing;

                float alt_error = RTL_ALTITUDE - (float)gps_pos.alt / 1000.0f;
                if (alt_error > 0) {
                    rtl_cmd.throttle = CONSTRAIN(0.5f + alt_error * 0.1f, 0.3f, 0.8f);
                } else {
                    rtl_cmd.throttle = CONSTRAIN(0.5f + alt_error * 0.1f, 0.2f, 0.7f);
                }
            } else {
                rtl_cmd.roll = 0;
                rtl_cmd.pitch = 0;
                rtl_cmd.yaw = 0;
                rtl_cmd.throttle = 0.3f;

                if ((float)gps_pos.alt / 1000.0f < 1.0f) {
                    flight_controller_disarm();
                    flight_controller_set_mode(FLIGHT_MODE_MANUAL);
                }
            }

            flight_controller_set_mavlink_command(&rtl_cmd);
        } else if (mode == FLIGHT_MODE_LAND) {
            ControlCommand land_cmd;
            land_cmd.roll = 0;
            land_cmd.pitch = 0;
            land_cmd.yaw = 0;
            land_cmd.throttle = CONSTRAIN(0.4f - 0.05f * fabsf(gps_vel.vd), 0.2f, 0.5f);

            if ((float)gps_pos.alt / 1000.0f < 0.5f) {
                flight_controller_disarm();
                flight_controller_set_mode(FLIGHT_MODE_MANUAL);
            }

            flight_controller_set_mavlink_command(&land_cmd);
        }

        vTaskDelayUntil(&last_wake_time, period);
    }
}
