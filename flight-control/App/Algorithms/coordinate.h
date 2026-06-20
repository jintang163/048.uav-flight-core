#ifndef __COORDINATE_H__
#define __COORDINATE_H__

#include "types.h"

void wgs84_to_ned(int32_t lat, int32_t lon, int32_t alt,
                  int32_t ref_lat, int32_t ref_lon, int32_t ref_alt,
                  float *north, float *east, float *down);
void ned_to_wgs84(float north, float east, float down,
                  int32_t ref_lat, int32_t ref_lon, int32_t ref_alt,
                  int32_t *lat, int32_t *lon, int32_t *alt);
void body_to_ned(const Vector3f *body, float roll, float pitch, float yaw, Vector3f *ned);
void ned_to_body(const Vector3f *ned, float roll, float pitch, float yaw, Vector3f *body);
float get_distance(int32_t lat1, int32_t lon1, int32_t lat2, int32_t lon2);
float get_bearing(int32_t lat1, int32_t lon1, int32_t lat2, int32_t lon2);
float wrap_pi(float angle);
float wrap_2pi(float angle);
float wrap_180(float angle);
float wrap_360(float angle);
float constrain_angle(float angle, float min, float max);

#endif
