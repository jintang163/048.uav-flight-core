#ifndef __FLIGHT_CONTROLLER_H__
#define __FLIGHT_CONTROLLER_H__

#include "types.h"
#include "flight_config.h"
#include "pid_controller.h"

typedef enum {
    FCS_DISARMED = 0,
    FCS_ARMED = 1,
    FCS_TAKEOFF = 2,
    FCS_LANDING = 3,
    FCS_EMERGENCY = 4
} FlightControlState;

typedef struct {
    FlightControlState state;
    FlightMode mode;
    FlightMode requested_mode;
    ControlSource control_source;
    bool armed;
    bool failsafe_active;
    AttitudeState attitude;
    PositionState position;
    ControlCommand rc_command;
    ControlCommand mavlink_command;
    ControlCommand final_command;
    ControlCommand attitude_target;
    ControlCommand rate_target;
    GPSPosition home_position;
    bool home_set;
    uint32_t arm_time;
    uint32_t flight_time;
} FlightControllerData;

void flight_controller_init(void);
void flight_controller_update(float dt);
void flight_controller_set_mode(FlightMode mode);
FlightMode flight_controller_get_mode(void);
void flight_controller_arm(void);
void flight_controller_disarm(void);
bool flight_controller_is_armed(void);
void flight_controller_set_rc_command(ControlCommand *cmd);
void flight_controller_set_mavlink_command(ControlCommand *cmd);
void flight_controller_get_attitude_target(ControlCommand *target);
void flight_controller_get_rate_target(ControlCommand *target);
void flight_controller_get_final_command(ControlCommand *cmd);
void flight_controller_set_home_position(GPSPosition *pos);
void flight_controller_get_home_position(GPSPosition *pos);
bool flight_controller_is_home_set(void);
void flight_controller_trigger_failsafe(void);
void flight_controller_clear_failsafe(void);
bool flight_controller_is_failsafe_active(void);
FlightControlState flight_controller_get_state(void);
void flight_controller_get_position(PositionState *pos);
void flight_controller_set_target_altitude(float altitude);
void flight_controller_goto_position(float lat, float lng, float alt);
float flight_controller_get_heading(void);

#endif
