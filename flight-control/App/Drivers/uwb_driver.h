#ifndef __UWB_DRIVER_H__
#define __UWB_DRIVER_H__

#include "types.h"
#include "flight_config.h"

#define UWB_MAX_NEIGHBORS MAX_FORMATION_UAVS
#define UWB_RANGE_VALIDITY_MS 500

typedef struct {
    uint8_t anchor_id;
    float range_m;
    float quality;
    uint32_t timestamp;
    bool valid;
} UWBRangeMeasurement;

typedef struct {
    uint8_t local_id;
    uint8_t neighbor_count;
    UWBRangeMeasurement measurements[UWB_MAX_NEIGHBORS];
    bool initialized;
    uint32_t last_update;
} UWBState;

bool uwb_driver_init(void);
void uwb_driver_update(void);
void uwb_driver_update_neighbor(uint8_t anchor_id, float range, float rel_x, float rel_y, float rel_z);
void uwb_driver_set_local_id(uint8_t id);
uint8_t uwb_driver_get_local_id(void);
bool uwb_driver_get_range(uint8_t anchor_id, float *range);
bool uwb_driver_get_position(uint8_t anchor_id, Vector3f *relative_pos);
uint8_t uwb_driver_get_neighbor_count(void);

#endif
