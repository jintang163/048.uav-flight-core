#include "tracking_manager.h"
#include "flight_config.h"
#include "pid_controller.h"
#include <string.h>
#include <math.h>

static TrackingData tracking_data;
static PIDController tracking_vel_x_pid;
static PIDController tracking_vel_y_pid;
static PIDController tracking_yaw_pid;

#define TRACKING_VEL_P 2.0f
#define TRACKING_VEL_I 0.05f
#define TRACKING_VEL_D 0.3f
#define TRACKING_VEL_I_MAX 2.0f

#define TRACKING_YAW_P 1.5f
#define TRACKING_YAW_I 0.02f
#define TRACKING_YAW_D 0.2f
#define TRACKING_YAW_I_MAX 0.5f

#define SEARCH_RADIUS_INCREMENT 5.0f

void tracking_manager_init(void)
{
    memset(&tracking_data, 0, sizeof(TrackingData));
    tracking_data.state = TRACKING_STATE_IDLE;
    tracking_data.search_radius = TRACKING_DEFAULT_SEARCH_RADIUS;
    tracking_data.max_search_radius = TRACKING_MAX_SEARCH_RADIUS;

    pid_init(&tracking_vel_x_pid, TRACKING_VEL_P, TRACKING_VEL_I, TRACKING_VEL_D,
             TRACKING_VEL_I_MAX, TRACKING_MAX_VELOCITY);
    pid_init(&tracking_vel_y_pid, TRACKING_VEL_P, TRACKING_VEL_I, TRACKING_VEL_D,
             TRACKING_VEL_I_MAX, TRACKING_MAX_VELOCITY);
    pid_init(&tracking_yaw_pid, TRACKING_YAW_P, TRACKING_YAW_I, TRACKING_YAW_D,
             TRACKING_YAW_I_MAX, MAX_YAW_RATE);
}

void tracking_manager_update(float dt)
{
    if (tracking_data.state == TRACKING_STATE_IDLE) {
        return;
    }

    if (tracking_data.current_target.valid) {
        uint32_t age = HAL_GetTick() - tracking_data.current_target.last_update;
        if (age > 500) {
            tracking_data.current_target.valid = false;
        }
    }

    if (tracking_data.current_target.valid) {
        float offset_x = tracking_data.current_target.center_offset_x;
        float offset_y = tracking_data.current_target.center_offset_y;

        if (fabsf(offset_x) > TRACKING_CENTER_TOLERANCE ||
            fabsf(offset_y) > TRACKING_CENTER_TOLERANCE) {
            tracking_data.velocity_e = pid_compute(&tracking_vel_x_pid, 0.0f, offset_x, dt);
            tracking_data.velocity_n = pid_compute(&tracking_vel_y_pid, 0.0f, offset_y, dt);
            tracking_data.yaw_rate = pid_compute(&tracking_yaw_pid, 0.0f, offset_x, dt);
        } else {
            tracking_data.velocity_n = 0.0f;
            tracking_data.velocity_e = 0.0f;
            tracking_data.yaw_rate = 0.0f;
            pid_reset(&tracking_vel_x_pid);
            pid_reset(&tracking_vel_y_pid);
            pid_reset(&tracking_yaw_pid);
        }
    } else {
        tracking_data.velocity_n = 0.0f;
        tracking_data.velocity_e = 0.0f;
        tracking_data.yaw_rate = 0.0f;
        pid_reset(&tracking_vel_x_pid);
        pid_reset(&tracking_vel_y_pid);
        pid_reset(&tracking_yaw_pid);
    }
}

TrackingState tracking_manager_get_state(void)
{
    return tracking_data.state;
}

void tracking_manager_set_state(TrackingState state)
{
    tracking_data.state = state;
}

void tracking_manager_set_target(float bbox_x, float bbox_y,
                                  float bbox_width, float bbox_height,
                                  float confidence)
{
    tracking_data.current_target.bbox_x = bbox_x;
    tracking_data.current_target.bbox_y = bbox_y;
    tracking_data.current_target.bbox_width = bbox_width;
    tracking_data.current_target.bbox_height = bbox_height;
    tracking_data.current_target.confidence = confidence;

    float center_x = bbox_x + bbox_width / 2.0f;
    float center_y = bbox_y + bbox_height / 2.0f;
    tracking_data.current_target.center_offset_x = (center_x / 1280.0f - 0.5f) * 2.0f;
    tracking_data.current_target.center_offset_y = (center_y / 720.0f - 0.5f) * 2.0f;

    tracking_data.current_target.last_update = HAL_GetTick();
    tracking_data.current_target.valid = true;

    tracking_data.state = TRACKING_STATE_LOCKING;
    tracking_data.frames_visible = 0;
    tracking_data.frames_lost = 0;
    tracking_data.search_radius = TRACKING_DEFAULT_SEARCH_RADIUS;
    tracking_data.searching = false;
    tracking_data.start_time = HAL_GetTick();
}

void tracking_manager_update_detection(float bbox_x, float bbox_y,
                                        float bbox_width, float bbox_height,
                                        float confidence)
{
    tracking_data.current_target.bbox_x = bbox_x;
    tracking_data.current_target.bbox_y = bbox_y;
    tracking_data.current_target.bbox_width = bbox_width;
    tracking_data.current_target.bbox_height = bbox_height;
    tracking_data.current_target.confidence = confidence;

    float center_x = bbox_x + bbox_width / 2.0f;
    float center_y = bbox_y + bbox_height / 2.0f;
    tracking_data.current_target.center_offset_x = (center_x / 1280.0f - 0.5f) * 2.0f;
    tracking_data.current_target.center_offset_y = (center_y / 720.0f - 0.5f) * 2.0f;

    tracking_data.current_target.last_update = HAL_GetTick();
    tracking_data.current_target.valid = true;

    tracking_data.frames_visible++;
    tracking_data.frames_lost = 0;

    switch (tracking_data.state) {
        case TRACKING_STATE_LOCKING:
            if (tracking_data.frames_visible >= TRACKING_FRAMES_TO_LOCK) {
                tracking_data.state = TRACKING_STATE_TRACKING;
                tracking_data.searching = false;
                tracking_data.search_radius = TRACKING_DEFAULT_SEARCH_RADIUS;
            }
            break;

        case TRACKING_STATE_SEARCHING:
            tracking_data.state = TRACKING_STATE_TRACKING;
            tracking_data.searching = false;
            tracking_data.search_radius = TRACKING_DEFAULT_SEARCH_RADIUS;
            break;

        default:
            break;
    }
}

void tracking_manager_lost_target(void)
{
    tracking_data.frames_lost++;
    tracking_data.frames_visible = 0;

    switch (tracking_data.state) {
        case TRACKING_STATE_TRACKING:
        case TRACKING_STATE_LOCKING:
            if (tracking_data.frames_lost >= TRACKING_FRAMES_TO_SEARCH) {
                tracking_data.state = TRACKING_STATE_SEARCHING;
                tracking_data.searching = true;
                if (tracking_data.search_radius < tracking_data.max_search_radius) {
                    tracking_data.search_radius = fminf(
                        tracking_data.search_radius + SEARCH_RADIUS_INCREMENT,
                        tracking_data.max_search_radius);
                }
            }
            break;

        case TRACKING_STATE_SEARCHING:
            if (tracking_data.frames_lost >= TRACKING_FRAMES_TO_SEARCH &&
                tracking_data.frames_lost % TRACKING_FRAMES_TO_SEARCH == 0) {
                if (tracking_data.search_radius < tracking_data.max_search_radius) {
                    tracking_data.search_radius = fminf(
                        tracking_data.search_radius + SEARCH_RADIUS_INCREMENT,
                        tracking_data.max_search_radius);
                }
            }
            if (tracking_data.frames_lost >= TRACKING_FRAMES_TO_LOST) {
                tracking_data.state = TRACKING_STATE_LOST;
                tracking_data.searching = false;
            }
            break;

        default:
            break;
    }
}

void tracking_manager_reset(void)
{
    memset(&tracking_data, 0, sizeof(TrackingData));
    tracking_data.state = TRACKING_STATE_IDLE;
    tracking_data.search_radius = TRACKING_DEFAULT_SEARCH_RADIUS;
    tracking_data.max_search_radius = TRACKING_MAX_SEARCH_RADIUS;

    pid_reset(&tracking_vel_x_pid);
    pid_reset(&tracking_vel_y_pid);
    pid_reset(&tracking_yaw_pid);
}

void tracking_manager_get_velocity_commands(float *velocity_n,
                                             float *velocity_e,
                                             float *velocity_d,
                                             float *yaw_rate)
{
    *velocity_n = tracking_data.velocity_n;
    *velocity_e = tracking_data.velocity_e;
    *velocity_d = 0.0f;
    *yaw_rate = tracking_data.yaw_rate;
}

bool tracking_manager_is_searching(void)
{
    return tracking_data.searching;
}

float tracking_manager_get_search_radius(void)
{
    return tracking_data.search_radius;
}

void tracking_manager_update_target_position(float lat, float lon)
{
    tracking_data.target_latitude = lat;
    tracking_data.target_longitude = lon;
}
