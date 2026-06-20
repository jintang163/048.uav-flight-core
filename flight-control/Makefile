######################################
# target
######################################
TARGET = flight_control_f4

######################################
# building variables
######################################
DEBUG = 1
OPT = -Og

#######################################
# paths
#######################################
BUILD_DIR = build

CORE_DIR = Core
APP_DIR = App
DRIVERS_DIR = Drivers

C_INCLUDES =  \
-I$(CORE_DIR)/Inc \
-I$(APP_DIR)/Algorithms \
-I$(APP_DIR)/Tasks \
-I$(APP_DIR)/Drivers \
-I$(DRIVERS_DIR)/STM32F4xx_HAL_Driver/Inc \
-I$(DRIVERS_DIR)/STM32F4xx_HAL_Driver/Inc/Legacy \
-I$(DRIVERS_DIR)/CMSIS/Device/ST/STM32F4xx/Include \
-I$(DRIVERS_DIR)/CMSIS/Include \
-I$(DRIVERS_DIR)/FreeRTOS/Source/include \
-I$(DRIVERS_DIR)/FreeRTOS/Source/portable/GCC/ARM_CM4F \
-I$(DRIVERS_DIR)/MAVLink/common \
-I$(DRIVERS_DIR)/MAVLink

######################################
# Sources
######################################

C_SOURCES =  \
$(CORE_DIR)/Src/main.c \
$(CORE_DIR)/Src/flight_controller.c \
$(CORE_DIR)/Src/sensor_manager.c \
$(CORE_DIR)/Src/mission_manager.c \
$(CORE_DIR)/Src/mavlink_handler.c \
$(CORE_DIR)/Src/motor_control.c \
$(CORE_DIR)/Src/attitude_estimator.c \
$(CORE_DIR)/Src/formation_manager.c \
$(CORE_DIR)/Src/stm32f4xx_it.c \
$(CORE_DIR)/Src/stm32f4xx_hal_msp.c \
$(CORE_DIR)/Src/system_stm32f4xx.c \
\
$(APP_DIR)/Algorithms/ekf.c \
$(APP_DIR)/Algorithms/pid_controller.c \
$(APP_DIR)/Algorithms/quaternion.c \
$(APP_DIR)/Algorithms/coordinate.c \
\
$(APP_DIR)/Tasks/task_attitude_estimation.c \
$(APP_DIR)/Tasks/task_flight_control.c \
$(APP_DIR)/Tasks/task_sensor_read.c \
$(APP_DIR)/Tasks/task_mavlink_comm.c \
$(APP_DIR)/Tasks/task_mission.c \
$(APP_DIR)/Tasks/task_motor_control.c \
$(APP_DIR)/Tasks/task_health_monitor.c \
$(APP_DIR)/Tasks/task_formation.c \
\
$(APP_DIR)/Drivers/mpu6050.c \
$(APP_DIR)/Drivers/gps_ublox.c \
$(APP_DIR)/Drivers/magnetometer.c \
$(APP_DIR)/Drivers/barometer.c \
$(APP_DIR)/Drivers/sbus_rc.c \
$(APP_DIR)/Drivers/pwm_output.c \
$(APP_DIR)/Drivers/rgb_led.c \
$(APP_DIR)/Drivers/uwb_driver.c \
\
$(DRIVERS_DIR)/STM32F4xx_HAL_Driver/Src/stm32f4xx_hal.c \
$(DRIVERS_DIR)/STM32F4xx_HAL_Driver/Src/stm32f4xx_hal_i2c.c \
$(DRIVERS_DIR)/STM32F4xx_HAL_Driver/Src/stm32f4xx_hal_tim.c \
$(DRIVERS_DIR)/STM32F4xx_HAL_Driver/Src/stm32f4xx_hal_uart.c \
$(DRIVERS_DIR)/STM32F4xx_HAL_Driver/Src/stm32f4xx_hal_dma.c \
$(DRIVERS_DIR)/STM32F4xx_HAL_Driver/Src/stm32f4xx_hal_rcc.c \
$(DRIVERS_DIR)/STM32F4xx_HAL_Driver/Src/stm32f4xx_hal_gpio.c \
$(DRIVERS_DIR)/STM32F4xx_HAL_Driver/Src/stm32f4xx_hal_cortex.c \
$(DRIVERS_DIR)/STM32F4xx_HAL_Driver/Src/stm32f4xx_hal_pwr.c \
$(DRIVERS_DIR)/STM32F4xx_HAL_Driver/Src/stm32f4xx_hal_flash.c \
$(DRIVERS_DIR)/STM32F4xx_HAL_Driver/Src/stm32f4xx_hal_exti.c \
\
$(DRIVERS_DIR)/FreeRTOS/Source/croutine.c \
$(DRIVERS_DIR)/FreeRTOS/Source/event_groups.c \
$(DRIVERS_DIR)/FreeRTOS/Source/list.c \
$(DRIVERS_DIR)/FreeRTOS/Source/queue.c \
$(DRIVERS_DIR)/FreeRTOS/Source/stream_buffer.c \
$(DRIVERS_DIR)/FreeRTOS/Source/tasks.c \
$(DRIVERS_DIR)/FreeRTOS/Source/timers.c \
$(DRIVERS_DIR)/FreeRTOS/Source/portable/MemMang/heap_4.c \
$(DRIVERS_DIR)/FreeRTOS/Source/portable/GCC/ARM_CM4F/port.c

ASM_SOURCES =  \
$(DRIVERS_DIR)/CMSIS/Device/ST/STM32F4xx/Source/Templates/gcc/startup_stm32f407xx.s

#######################################
# binaries
#######################################
PREFIX = arm-none-eabi-
CC = $(PREFIX)gcc
AS = $(PREFIX)gcc -x assembler-with-cpp
CP = $(PREFIX)objcopy
SZ = $(PREFIX)size

HEX = $(CP) -O ihex
BIN = $(CP) -O binary -S

#######################################
# CFLAGS
#######################################
CPU = -mcpu=cortex-m4
FPU = -mfpu=fpv4-sp-d16
FLOAT-ABI = -mfloat-abi=hard
MCU = $(CPU) -mthumb $(FPU) $(FLOAT-ABI)

C_DEFS =  \
-DUSE_HAL_DRIVER \
-DSTM32F407xx \
-DARM_MATH_CM4

CFLAGS += $(MCU) $(C_DEFS) $(C_INCLUDES) $(OPT) -Wall -fdata-sections -ffunction-sections

ifeq ($(DEBUG), 1)
CFLAGS += -g -gdwarf-2
endif

CFLAGS += -MMD -MP -MF"$(@:%.o=%.d)"

#######################################
# LDFLAGS
#######################################
LDSCRIPT = STM32F407VGTx_FLASH.ld

LIBS = -lc -lm -lnosys
LIBDIR =

LDFLAGS = $(MCU) -specs=nano.specs -T$(LDSCRIPT) $(LIBDIR) $(LIBS) -Wl,-Map=$(BUILD_DIR)/$(TARGET).map,--cref -Wl,--gc-sections

#######################################
# build the application
#######################################

OBJECTS = $(addprefix $(BUILD_DIR)/,$(notdir $(C_SOURCES:.c=.o)))
vpath %.c $(sort $(dir $(C_SOURCES)))

OBJECTS += $(addprefix $(BUILD_DIR)/,$(notdir $(ASM_SOURCES:.s=.o)))
vpath %.s $(sort $(dir $(ASM_SOURCES)))

all: $(BUILD_DIR)/$(TARGET).elf $(BUILD_DIR)/$(TARGET).hex $(BUILD_DIR)/$(TARGET).bin

$(BUILD_DIR)/%.o: %.c Makefile | $(BUILD_DIR)
	@echo [CC] $<
	@$(CC) -c $(CFLAGS) -Wa,-a,-ad,-alms=$(BUILD_DIR)/$(notdir $(<:.c=.lst)) $< -o $@

$(BUILD_DIR)/%.o: %.s Makefile | $(BUILD_DIR)
	@echo [AS] $<
	@$(AS) -c $(CFLAGS) $< -o $@

$(BUILD_DIR)/$(TARGET).elf: $(OBJECTS) Makefile
	@echo [LD] $(TARGET).elf
	@$(CC) $(OBJECTS) $(LDFLAGS) -o $@
	@$(SZ) $@

$(BUILD_DIR)/%.hex: $(BUILD_DIR)/%.elf | $(BUILD_DIR)
	@echo [HEX] $@
	@$(HEX) $< $@

$(BUILD_DIR)/%.bin: $(BUILD_DIR)/%.elf | $(BUILD_DIR)
	@echo [BIN] $@
	@$(BIN) $< $@

$(BUILD_DIR):
	mkdir $@

#######################################
# clean up
#######################################
clean:
	-rm -fR $(BUILD_DIR)

#######################################
# flash
#######################################
flash: $(BUILD_DIR)/$(TARGET).elf
	openocd -f interface/stlink.cfg -f target/stm32f4x.cfg -c "program $(BUILD_DIR)/$(TARGET).elf verify reset exit"

#######################################
# dependencies
#######################################
-include $(wildcard $(BUILD_DIR)/*.d)

.PHONY: all clean flash
