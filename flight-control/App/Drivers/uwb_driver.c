#include "uwb_driver.h"
#include "formation_manager.h"
#include "stm32f4xx_hal.h"
#include <string.h>

static UWBState uwb_state;
static Vector3f neighbor_positions[UWB_MAX_NEIGHBORS];

bool uwb_driver_init(void)
{
    memset(&uwb_state, 0, sizeof(UWBState));
    memset(neighbor_positions, 0, sizeof(neighbor_positions));

    uwb_state.local_id = 0;
    uwb_state.initialized = true;
    uwb_state.last_update = HAL_GetTick();

    return true;
}

void uwb_driver_update(void)
{
    uint32_t now = HAL_GetTick();

    for (uint8_t i = 0; i < UWB_MAX_NEIGHBORS; i++) {
        if (uwb_state.measurements[i].valid) {
            uint32_t age = now - uwb_state.measurements[i].timestamp;
            if (age > UWB_RANGE_VALIDITY_MS) {
                uwb_state.measurements[i].valid = false;
            }
        }
    }

    uint8_t count = 0;
    for (uint8_t i = 0; i < UWB_MAX_NEIGHBORS; i++) {
        if (uwb_state.measurements[i].valid) {
            count++;
        }
    }
    uwb_state.neighbor_count = count;
}

void uwb_driver_update_neighbor(uint8_t anchor_id, float range, float rel_x, float rel_y, float rel_z)
{
    if (anchor_id >= UWB_MAX_NEIGHBORS) return;

    uwb_state.measurements[anchor_id].anchor_id = anchor_id;
    uwb_state.measurements[anchor_id].range_m = range;
    uwb_state.measurements[anchor_id].quality = 1.0f;
    uwb_state.measurements[anchor_id].timestamp = HAL_GetTick();
    uwb_state.measurements[anchor_id].valid = true;

    neighbor_positions[anchor_id].x = rel_x;
    neighbor_positions[anchor_id].y = rel_y;
    neighbor_positions[anchor_id].z = rel_z;

    Vector3f vel = {0.0f, 0.0f, 0.0f};
    float yaw = 0.0f;
    formation_manager_update_neighbor(anchor_id, &neighbor_positions[anchor_id], &vel, yaw);

    uwb_state.last_update = HAL_GetTick();
}

void uwb_driver_set_local_id(uint8_t id)
{
    uwb_state.local_id = id;
}

uint8_t uwb_driver_get_local_id(void)
{
    return uwb_state.local_id;
}

bool uwb_driver_get_range(uint8_t anchor_id, float *range)
{
    if (anchor_id >= UWB_MAX_NEIGHBORS) return false;
    if (!uwb_state.measurements[anchor_id].valid) return false;

    *range = uwb_state.measurements[anchor_id].range_m;
    return true;
}

bool uwb_driver_get_position(uint8_t anchor_id, Vector3f *relative_pos)
{
    if (anchor_id >= UWB_MAX_NEIGHBORS) return false;
    if (!uwb_state.measurements[anchor_id].valid) return false;

    *relative_pos = neighbor_positions[anchor_id];
    return true;
}

uint8_t uwb_driver_get_neighbor_count(void)
{
    return uwb_state.neighbor_count;
}
