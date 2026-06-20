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
    FLIGHT_MODE_LAND = 5,
    FLIGHT_MODE_FORMATION = 6,
    FLIGHT_MODE_TRACKING = 7
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

typedef enum {
    FORMATION_LINE = 0,
    FORMATION_TRIANGLE = 1,
    FORMATION_CIRCLE = 2
} FormationType;

typedef enum {
    FORMATION_STATE_IDLE = 0,
    FORMATION_STATE_READY = 1,
    FORMATION_STATE_EXECUTING = 2,
    FORMATION_STATE_PAUSED = 3
} FormationState;

typedef struct {
    uint8_t uav_id;
    Vector3f relative_pos;
    Vector3f velocity;
    float yaw;
    uint32_t last_update;
    bool online;
} UWBNeighborInfo;

#define MAX_FORMATION_UAVS 16
#define FORMATION_POSITION_ERROR_MAX 0.3f
#define COLLISION_WARNING_DISTANCE 5.0f
#define COLLISION_DECELERATION_FACTOR 0.5f

typedef struct {
    FormationType type;
    FormationState state;
    uint8_t uav_id;
    uint8_t total_uavs;
    float spacing;
    uint8_t leader_id;
    bool is_leader;
    Vector3f formation_offset;
    UWBNeighborInfo neighbors[MAX_FORMATION_UAVS];
    uint8_t neighbor_count;
    bool collision_warning;
    float min_distance;
    uint8_t closest_uav_id;
    uint32_t sync_timestamp;
    bool synced;
} FormationData;

typedef struct {
    uint8_t r;
    uint8_t g;
    uint8_t b;
    uint8_t effect;
    uint32_t timestamp;
} FormationLightCommand;

#define LIGHT_EFFECT_STATIC 0
#define LIGHT_EFFECT_BLINK 1
#define LIGHT_EFFECT_RAINBOW 2
#define LIGHT_EFFECT_BREATHING 3

typedef enum {
    TRACKING_STATE_IDLE = 0,
    TRACKING_STATE_LOCKING = 1,
    TRACKING_STATE_TRACKING = 2,
    TRACKING_STATE_SEARCHING = 3,
    TRACKING_STATE_LOST = 4
} TrackingState;

typedef struct {
    float bbox_x;
    float bbox_y;
    float bbox_width;
    float bbox_height;
    float center_offset_x;
    float center_offset_y;
    float confidence;
    uint32_t last_update;
    bool valid;
} DetectionTarget;

typedef struct {
    TrackingState state;
    DetectionTarget current_target;
    float search_radius;
    float max_search_radius;
    uint16_t frames_visible;
    uint16_t frames_lost;
    float target_latitude;
    float target_longitude;
    float velocity_n;
    float velocity_e;
    float yaw_rate;
    bool searching;
    uint32_t start_time;
} TrackingData;

#define TRACKING_DEFAULT_SEARCH_RADIUS 10.0f
#define TRACKING_MAX_SEARCH_RADIUS 50.0f
#define TRACKING_FRAMES_TO_LOCK 10
#define TRACKING_FRAMES_TO_SEARCH 15
#define TRACKING_FRAMES_TO_LOST 60
#define TRACKING_CENTER_TOLERANCE 0.05f
#define TRACKING_MAX_VELOCITY 3.0f

#endif
