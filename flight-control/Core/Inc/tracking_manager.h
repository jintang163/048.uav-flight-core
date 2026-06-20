#ifndef __TRACKING_MANAGER_H__
#define __TRACKING_MANAGER_H__

#include "types.h"

void tracking_manager_init(void);
void tracking_manager_update(float dt);

TrackingState tracking_manager_get_state(void);
void tracking_manager_set_state(TrackingState state);

void tracking_manager_set_target(float bbox_x, float bbox_y,
                                  float bbox_width, float bbox_height,
                                  float confidence);

void tracking_manager_update_detection(float bbox_x, float bbox_y,
                                        float bbox_width, float bbox_height,
                                        float confidence);

void tracking_manager_lost_target(void);
void tracking_manager_reset(void);

void tracking_manager_get_velocity_commands(float *velocity_n,
                                             float *velocity_e,
                                             float *velocity_d,
                                             float *yaw_rate);

bool tracking_manager_is_searching(void);
float tracking_manager_get_search_radius(void);

void tracking_manager_update_target_position(float lat, float lon);

#endif
