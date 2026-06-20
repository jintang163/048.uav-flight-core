#ifndef __TYPES_H__
#define __TYPES_H__

#include <stdint.h>
#include <stdbool.h>
#include <math.h>

#ifndef M_PI
#define M_PI 3.14159265358979323846f
#endif

#define DEG2RAD(x) ((x) * (M_PI / 180.0f))
#define RAD2DEG(x) ((x) * (180.0f / M_PI))

#define CONSTRAIN(x, min, max) ((x) < (min) ? (min) : ((x) > (max) ? (max) : (x)))
#define DEADBAND(x, db) (fabsf(x) < (db) ? 0.0f : (x))

typedef struct {
    float x;
    float y;
    float z;
} Vector3f;

typedef struct {
    float w;
    float x;
    float y;
    float z;
} Quaternion;

typedef struct {
    float roll;
    float pitch;
    float yaw;
} EulerAngle;

typedef struct {
    int32_t lat;
    int32_t lon;
    int32_t alt;
} GPSPosition;

typedef struct {
    float vn;
    float ve;
    float vd;
} GPSVelocity;

typedef enum {
    FLIGHT_MODE_MANUAL = 0,
    FLIGHT_MODE_ALT_HOLD = 1,
    FLIGHT_MODE_POS_HOLD = 2,
    FLIGHT_MODE_AUTO = 3,
    FLIGHT_MODE_RTL = 4,
    FLIGHT_MODE_LAND = 5
} FlightMode;

typedef enum {
    CONTROL_SOURCE_RC = 0,
    CONTROL_SOURCE_MAVLINK = 1
} ControlSource;

typedef struct {
    float roll;
    float pitch;
    float yaw;
    float throttle;
} ControlCommand;

typedef struct {
    Quaternion quat;
    EulerAngle euler;
    Vector3f angular_velocity;
    Vector3f linear_accel;
    float yaw_rate;
} AttitudeState;

typedef struct {
    GPSPosition position;
    GPSVelocity velocity;
    float altitude;
    float ground_speed;
    float heading;
    uint8_t fix_type;
    uint8_t satellites;
    float hdop;
} PositionState;

typedef struct {
    float voltage;
    float current;
    float capacity_used;
    float battery_percent;
} BatteryState;

typedef struct {
    uint16_t channels[16];
    bool connected;
    uint32_t last_update;
} RCInput;

typedef struct {
    float motor[4];
    uint32_t pwm_output[4];
} MotorOutput;

typedef struct {
    bool imu_ok;
    bool gps_ok;
    bool mag_ok;
    bool baro_ok;
    bool rc_ok;
    bool battery_ok;
    uint32_t error_flags;
} HealthStatus;

typedef enum {
    MISSION_WAYPOINT = 0,
    MISSION_TAKEOFF = 1,
    MISSION_LAND = 2,
    MISSION_LOITER = 3,
    MISSION_RETURN = 4
} MissionItemType;

typedef struct {
    MissionItemType type;
    int32_t lat;
    int32_t lon;
    int32_t alt;
    float heading;
    float hold_time;
    float radius;
} MissionItem;

typedef struct {
    MissionItem items[50];
    uint16_t count;
    uint16_t current_index;
    bool active;
    bool finished;
} MissionPlan;

#define ERROR_FLAG_IMU_TIMEOUT     (1 << 0)
#define ERROR_FLAG_GPS_TIMEOUT     (1 << 1)
#define ERROR_FLAG_MAG_TIMEOUT     (1 << 2)
#define ERROR_FLAG_BARO_TIMEOUT    (1 << 3)
#define ERROR_FLAG_RC_SIGNAL_LOSS  (1 << 4)
#define ERROR_FLAG_LOW_BATTERY     (1 << 5)
#define ERROR_FLAG_MOTOR_FAILURE   (1 << 6)

#endif
