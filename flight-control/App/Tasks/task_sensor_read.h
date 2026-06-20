#ifndef __TASK_SENSOR_READ_H__
#define __TASK_SENSOR_READ_H__

#include "types.h"
#include "flight_config.h"
#include "FreeRTOS.h"
#include "task.h"

void task_sensor_read_init(void);
void task_sensor_read_main(void *argument);

#endif
