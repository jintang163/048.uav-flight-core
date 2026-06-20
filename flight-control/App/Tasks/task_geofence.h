#ifndef __TASK_GEOFENCE_H__
#define __TASK_GEOFENCE_H__

#include "main.h"

#define TASK_GEOFENCE_FREQ          10
#define TASK_GEOFENCE_STACK_SIZE    1024
#define TASK_GEOFENCE_PRIORITY      (tskIDLE_PRIORITY + 2)

void task_geofence_init(void);
void task_geofence_main(void *argument);
bool task_geofence_is_armed_blocked(void);
const char* task_geofence_get_block_reason(void);

#endif
