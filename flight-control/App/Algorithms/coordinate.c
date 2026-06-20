#include "coordinate.h"
#include "flight_config.h"

void wgs84_to_ned(int32_t lat, int32_t lon, int32_t alt,
                  int32_t ref_lat, int32_t ref_lon, int32_t ref_alt,
                  float *north, float *east, float *down)
{
    double lat_rad = (double)lat * 1e-7 * M_PI / 180.0;
    double lon_rad = (double)lon * 1e-7 * M_PI / 180.0;
    double ref_lat_rad = (double)ref_lat * 1e-7 * M_PI / 180.0;
    double ref_lon_rad = (double)ref_lon * 1e-7 * M_PI / 180.0;

    double dlat = lat_rad - ref_lat_rad;
    double dlon = lon_rad - ref_lon_rad;

    double sin_lat = sin(ref_lat_rad);
    double cos_lat = cos(ref_lat_rad);

    double Rn = WGS84_A / sqrt(1.0 - WGS84_E * WGS84_E * sin_lat * sin_lat);
    double Rm = Rn * (1.0 - WGS84_E * WGS84_E) / (1.0 - WGS84_E * WGS84_E * sin_lat * sin_lat);

    *north = (float)(dlat * Rm);
    *east = (float)(dlon * Rn * cos_lat);
    *down = (float)((ref_alt - alt) * 1e-3);
}

void ned_to_wgs84(float north, float east, float down,
                  int32_t ref_lat, int32_t ref_lon, int32_t ref_alt,
                  int32_t *lat, int32_t *lon, int32_t *alt)
{
    double ref_lat_rad = (double)ref_lat * 1e-7 * M_PI / 180.0;
    double sin_lat = sin(ref_lat_rad);
    double cos_lat = cos(ref_lat_rad);

    double Rn = WGS84_A / sqrt(1.0 - WGS84_E * WGS84_E * sin_lat * sin_lat);
    double Rm = Rn * (1.0 - WGS84_E * WGS84_E) / (1.0 - WGS84_E * WGS84_E * sin_lat * sin_lat);

    double dlat = (double)north / Rm;
    double dlon = (double)east / (Rn * cos_lat);

    double lat_rad = ref_lat_rad + dlat;
    double lon_rad = (double)ref_lon * 1e-7 * M_PI / 180.0 + dlon;

    *lat = (int32_t)(lat_rad * 180.0 / M_PI * 1e7);
    *lon = (int32_t)(lon_rad * 180.0 / M_PI * 1e7);
    *alt = (int32_t)((double)ref_alt - (double)down * 1000.0);
}

void body_to_ned(const Vector3f *body, float roll, float pitch, float yaw, Vector3f *ned)
{
    float cr = cosf(roll);
    float sr = sinf(roll);
    float cp = cosf(pitch);
    float sp = sinf(pitch);
    float cy = cosf(yaw);
    float sy = sinf(yaw);

    ned->x = (cy * cp) * body->x + (cy * sp * sr - sy * cr) * body->y + (cy * sp * cr + sy * sr) * body->z;
    ned->y = (sy * cp) * body->x + (sy * sp * sr + cy * cr) * body->y + (sy * sp * cr - cy * sr) * body->z;
    ned->z = (-sp) * body->x + (cp * sr) * body->y + (cp * cr) * body->z;
}

void ned_to_body(const Vector3f *ned, float roll, float pitch, float yaw, Vector3f *body)
{
    float cr = cosf(roll);
    float sr = sinf(roll);
    float cp = cosf(pitch);
    float sp = sinf(pitch);
    float cy = cosf(yaw);
    float sy = sinf(yaw);

    body->x = (cy * cp) * ned->x + (sy * cp) * ned->y + (-sp) * ned->z;
    body->y = (cy * sp * sr - sy * cr) * ned->x + (sy * sp * sr + cy * cr) * ned->y + (cp * sr) * ned->z;
    body->z = (cy * sp * cr + sy * sr) * ned->x + (sy * sp * cr - cy * sr) * ned->y + (cp * cr) * ned->z;
}

float get_distance(int32_t lat1, int32_t lon1, int32_t lat2, int32_t lon2)
{
    double lat1_rad = (double)lat1 * 1e-7 * M_PI / 180.0;
    double lon1_rad = (double)lon1 * 1e-7 * M_PI / 180.0;
    double lat2_rad = (double)lat2 * 1e-7 * M_PI / 180.0;
    double lon2_rad = (double)lon2 * 1e-7 * M_PI / 180.0;

    double dlat = lat2_rad - lat1_rad;
    double dlon = lon2_rad - lon1_rad;

    double a = sin(dlat / 2.0) * sin(dlat / 2.0) +
               cos(lat1_rad) * cos(lat2_rad) *
               sin(dlon / 2.0) * sin(dlon / 2.0);

    double c = 2.0 * atan2(sqrt(a), sqrt(1.0 - a));

    return (float)(EARTH_RADIUS * c);
}

float get_bearing(int32_t lat1, int32_t lon1, int32_t lat2, int32_t lon2)
{
    double lat1_rad = (double)lat1 * 1e-7 * M_PI / 180.0;
    double lon1_rad = (double)lon1 * 1e-7 * M_PI / 180.0;
    double lat2_rad = (double)lat2 * 1e-7 * M_PI / 180.0;
    double lon2_rad = (double)lon2 * 1e-7 * M_PI / 180.0;

    double dlon = lon2_rad - lon1_rad;

    double y = sin(dlon) * cos(lat2_rad);
    double x = cos(lat1_rad) * sin(lat2_rad) -
               sin(lat1_rad) * cos(lat2_rad) * cos(dlon);

    return (float)atan2(y, x);
}

float wrap_pi(float angle)
{
    while (angle > M_PI) {
        angle -= 2.0f * (float)M_PI;
    }
    while (angle < -(float)M_PI) {
        angle += 2.0f * (float)M_PI;
    }
    return angle;
}

float wrap_2pi(float angle)
{
    while (angle >= 2.0f * (float)M_PI) {
        angle -= 2.0f * (float)M_PI;
    }
    while (angle < 0.0f) {
        angle += 2.0f * (float)M_PI;
    }
    return angle;
}

float wrap_180(float angle)
{
    while (angle > 180.0f) {
        angle -= 360.0f;
    }
    while (angle < -180.0f) {
        angle += 360.0f;
    }
    return angle;
}

float wrap_360(float angle)
{
    while (angle >= 360.0f) {
        angle -= 360.0f;
    }
    while (angle < 0.0f) {
        angle += 360.0f;
    }
    return angle;
}

float constrain_angle(float angle, float min, float max)
{
    float diff = wrap_pi(angle - min);
    if (diff > wrap_pi(max - min)) {
        return max;
    }
    return min + diff;
}
