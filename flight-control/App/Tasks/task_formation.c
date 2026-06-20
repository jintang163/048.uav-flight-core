#include "task_formation.h"
#include "formation_manager.h"
#include "flight_controller.h"
#include "sensor_manager.h"
#include "mission_manager.h"
#include "coordinate.h"
#include "pid_controller.h"

static TaskHandle_t task_handle = NULL;
static PIDController pos_xy_pid;
static PIDController pos_z_pid;

#define FORMATION_POS_KP 1.2f
#define FORMATION_POS_KI 0.05f
#define FORMATION_POS_KD 0.3f
#define FORMATION_POS_I_MAX 2.0f
#define FORMATION_POS_OUT_MAX 5.0f

#define FORMATION_ALT_KP 1.0f
#define FORMATION_ALT_KI 0.1f
#define FORMATION_ALT_KD 0.5f
#define FORMATION_ALT_I_MAX 100.0f
#define FORMATION_ALT_OUT_MAX 5.0f

#define FORMATION_SYNC_TIMEOUT_MS 100

void task_formation_init(void)
{
    formation_manager_init();

    pid_controller_init(&pos_xy_pid,
                        FORMATION_POS_KP,
                        FORMATION_POS_KI,
                        FORMATION_POS_KD,
                        FORMATION_POS_I_MAX,
                        FORMATION_POS_OUT_MAX);

    pid_controller_init(&pos_z_pid,
                        FORMATION_ALT_KP,
                        FORMATION_ALT_KI,
                        FORMATION_ALT_KD,
                        FORMATION_ALT_I_MAX,
                        FORMATION_ALT_OUT_MAX);

    xTaskCreate(task_formation_main,
                "Formation",
                TASK_MISSION_STACK_SIZE,
                NULL,
                TASK_MISSION_PRIORITY,
                &task_handle);
}

static void calculate_formation_control(float dt,
                                        float *target_vel_n,
                                        float *target_vel_e,
                                        float *target_vel_d,
                                        float *target_yaw)
{
    Vector3f target_offset;
    formation_manager_get_target_offset(&target_offset);

    float leader_vel_n = 0.0f;
    float leader_vel_e = 0.0f;
    float leader_yaw = 0.0f;

    uint8_t leader_id = formation_manager_get_leader();
    UWBNeighborInfo leader_info;
    if (formation_manager_get_neighbor(leader_id, &leader_info)) {
        leader_vel_n = leader_info.velocity.x;
        leader_vel_e = leader_info.velocity.y;
        leader_yaw = leader_info.yaw;
    }

    float pos_error_n = target_offset.x;
    float pos_error_e = target_offset.y;
    float pos_error_d = target_offset.z;

    if (fabsf(leader_yaw) > 0.01f) {
        float cos_yaw = cosf(leader_yaw);
        float sin_yaw = sinf(leader_yaw);
        float error_n = pos_error_n * cos_yaw + pos_error_e * sin_yaw;
        float error_e = -pos_error_n * sin_yaw + pos_error_e * cos_yaw;
        pos_error_n = error_n;
        pos_error_e = error_e;
    }

    float vel_correction_n = pid_controller_update(&pos_xy_pid, pos_error_n, dt);
    float vel_correction_e = pid_controller_update(&pos_xy_pid, pos_error_e, dt);
    float vel_correction_d = pid_controller_update(&pos_z_pid, pos_error_d, dt);

    float decel_factor = formation_manager_get_collision_deceleration();

    *target_vel_n = (leader_vel_n + vel_correction_n) * decel_factor;
    *target_vel_e = (leader_vel_e + vel_correction_e) * decel_factor;
    *target_vel_d = vel_correction_d * decel_factor;
    *target_yaw = leader_yaw;
}

void task_formation_main(void *argument)
{
    UNUSED(argument);

    TickType_t last_wake_time = xTaskGetTickCount();
    const TickType_t period = pdMS_TO_TICKS(1000 / TASK_MISSION_FREQ);
    float dt = 1.0f / TASK_MISSION_FREQ;

    while (1) {
        GPSPosition gps_pos;
        GPSVelocity gps_vel;
        sensor_manager_get_gps(&gps_pos, &gps_vel);

        FlightMode mode = flight_controller_get_mode();

        if (mode == FLIGHT_MODE_FORMATION) {
            formation_manager_update(dt);

            FormationState formation_state = formation_manager_get_state();

            if (formation_state == FORMATION_STATE_EXECUTING) {
                ControlCommand formation_cmd;
                float target_vel_n, target_vel_e, target_vel_d, target_yaw;

                calculate_formation_control(dt, &target_vel_n, &target_vel_e, &target_vel_d, &target_yaw);

                float vel_error_n = target_vel_n - gps_vel.vn;
                float vel_error_e = target_vel_e - gps_vel.ve;

                formation_cmd.roll = CONSTRAIN(vel_error_e * 0.5f, -MAX_TILT_ANGLE, MAX_TILT_ANGLE);
                formation_cmd.pitch = CONSTRAIN(-vel_error_n * 0.5f, -MAX_TILT_ANGLE, MAX_TILT_ANGLE);
                formation_cmd.yaw = target_yaw;

                float baro_alt;
                sensor_manager_get_baro(&baro_alt, NULL, NULL);

                float alt_error = -target_vel_d * 0.5f;
                formation_cmd.throttle = CONSTRAIN(0.5f + alt_error, 0.3f, 0.8f);

                flight_controller_set_mavlink_command(&formation_cmd);
            }
        } else {
            pid_controller_reset(&pos_xy_pid);
            pid_controller_reset(&pos_z_pid);
        }

        vTaskDelayUntil(&last_wake_time, period);
    }
}
