#ifndef __TASK_MAVLINK_COMM_H__
#define __TASK_MAVLINK_COMM_H__

#include "types.h"
#include "flight_config.h"
#include "FreeRTOS.h"
#include "task.h"

void task_mavlink_comm_init(void);
void task_mavlink_comm_main(void *argument);

#endif
