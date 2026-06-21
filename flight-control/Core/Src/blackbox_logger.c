#include "blackbox_logger.h"
#include "sensor_manager.h"
#include "flight_controller.h"
#include "motor_control.h"
#include "attitude_estimator.h"
#include <string.h>
#include <stdio.h>

static BlackboxState bb_state;
static BlackboxHeader bb_header;

static uint8_t flash_buffer[BLACKBOX_SECTOR_SIZE];
static uint32_t buffer_index = 0;

static uint32_t flash_write_ptr = 0;
static uint32_t flash_read_ptr = 0;

#define FLASH_DATA_START_OFFSET  sizeof(BlackboxHeader)
#define FLASH_DATA_END_OFFSET    BLACKBOX_FLASH_SIZE

void blackbox_init(void)
{
    memset(&bb_state, 0, sizeof(BlackboxState));
    memset(&bb_header, 0, sizeof(BlackboxHeader));
    memset(flash_buffer, 0, sizeof(flash_buffer));

    bb_header.magic = BLACKBOX_HEADER_MAGIC;
    bb_header.version = BLACKBOX_VERSION;
    bb_header.sample_rate = BLACKBOX_SAMPLE_RATE;

    bb_state.initialized = true;
    bb_state.recording = false;
    bb_state.last_voltage = 0.0f;
    bb_state.last_gps_time = 0;
    bb_state.last_rc_time = 0;

    flash_write_ptr = FLASH_DATA_START_OFFSET;
    flash_read_ptr = FLASH_DATA_START_OFFSET;
    buffer_index = 0;
}

void blackbox_start(void)
{
    if (!bb_state.initialized) {
        return;
    }

    if (bb_state.recording) {
        return;
    }

    bb_state.current_flight_id++;
    bb_state.entry_count = 0;
    bb_state.recording = true;

    bb_header.flight_id = bb_state.current_flight_id;
    bb_header.start_time = HAL_GetTick();
    bb_header.end_time = 0;
    bb_header.total_entries = 0;
    bb_header.data_size = 0;

    flash_write_ptr = FLASH_DATA_START_OFFSET;
    buffer_index = 0;

    blackbox_log_event(BLACKBOX_EVENT_ARM, 0, 0, 0, 0, "Flight start - armed");
}

void blackbox_stop(void)
{
    if (!bb_state.recording) {
        return;
    }

    blackbox_log_event(BLACKBOX_EVENT_DISARM, 0, 0, 0, 0, "Flight end - disarmed");

    bb_state.recording = false;
    bb_header.end_time = HAL_GetTick();
    bb_header.total_entries = bb_state.entry_count;
    bb_header.data_size = flash_write_ptr - FLASH_DATA_START_OFFSET;

    if (buffer_index > 0) {
        buffer_index = 0;
    }
}

bool blackbox_is_recording(void)
{
    return bb_state.recording;
}

static void write_to_flash(uint8_t *data, uint32_t size)
{
    if (flash_write_ptr + size > FLASH_DATA_END_OFFSET) {
        flash_write_ptr = FLASH_DATA_START_OFFSET;
    }

    if (size <= sizeof(flash_buffer) - buffer_index) {
        memcpy(&flash_buffer[buffer_index], data, size);
        buffer_index += size;
        flash_write_ptr += size;
    } else {
        uint32_t first_part = sizeof(flash_buffer) - buffer_index;
        memcpy(&flash_buffer[buffer_index], data, first_part);
        buffer_index = 0;

        uint32_t remaining = size - first_part;
        if (remaining > 0) {
            memcpy(flash_buffer, &data[first_part], remaining);
            buffer_index = remaining;
        }
        flash_write_ptr += size;
    }
}

void blackbox_log_data(void)
{
    if (!bb_state.recording) {
        return;
    }

    BlackboxLogEntry entry;
    memset(&entry, 0, sizeof(entry));

    entry.type = BLACKBOX_LOG_TYPE_DATA;
    entry.size = sizeof(BlackboxDataEntry);

    BlackboxDataEntry *data = &entry.payload.data;
    data->timestamp = HAL_GetTick();

    AttitudeState att;
    attitude_estimator_get_attitude(&att);
    data->roll = att.euler.roll;
    data->pitch = att.euler.pitch;
    data->yaw = att.euler.yaw;

    PositionState pos;
    flight_controller_get_position(&pos);
    data->lat = pos.position.lat;
    data->lon = pos.position.lon;
    data->alt = pos.position.alt;
    data->vx = pos.velocity.vn;
    data->vy = pos.velocity.ve;
    data->vz = pos.velocity.vd;
    data->satellites = pos.satellites;
    data->gps_fix_type = pos.fix_type;

    BatteryState battery;
    sensor_manager_get_battery(&battery);
    data->voltage = battery.voltage;
    data->current = battery.current;
    data->throttle = battery.battery_percent;

    RCInput rc;
    sensor_manager_get_rc(&rc);
    for (int i = 0; i < 8; i++) {
        data->rc_channels[i] = rc.channels[i];
    }

    MotorState motor_state;
    motor_control_get_state(&motor_state);
    for (int i = 0; i < 4; i++) {
        data->motor_pwm[i] = motor_state.pwm[i];
    }

    data->flight_mode = (uint8_t)flight_controller_get_mode();

    HealthStatus health;
    task_health_monitor_get_status(&health);
    data->error_flags = (uint8_t)(health.error_flags & 0xFF);

    write_to_flash((uint8_t *)&entry, sizeof(entry.type) + sizeof(entry.size) + entry.size);
    bb_state.entry_count++;
    bb_header.total_entries = bb_state.entry_count;
}

void blackbox_log_event(uint32_t event_type, int32_t p1, int32_t p2, float p3, float p4, const char *desc)
{
    if (!bb_state.recording && !(event_type & (BLACKBOX_EVENT_ARM | BLACKBOX_EVENT_DISARM))) {
        return;
    }

    BlackboxLogEntry entry;
    memset(&entry, 0, sizeof(entry));

    entry.type = BLACKBOX_LOG_TYPE_EVENT;
    entry.size = sizeof(BlackboxEventEntry);

    BlackboxEventEntry *event = &entry.payload.event;
    event->timestamp = HAL_GetTick();
    event->event_type = event_type;
    event->param1 = p1;
    event->param2 = p2;
    event->param3 = p3;
    event->param4 = p4;

    if (desc) {
        strncpy(event->description, desc, sizeof(event->description) - 1);
    }

    if (event_type == BLACKBOX_EVENT_LOW_BATTERY ||
        event_type == BLACKBOX_EVENT_CRASH ||
        event_type == BLACKBOX_EVENT_MOTOR_FAILURE) {
        event->event_severity = 3;
    } else if (event_type == BLACKBOX_EVENT_GPS_LOSS ||
               event_type == BLACKBOX_EVENT_RC_LOSS ||
               event_type == BLACKBOX_EVENT_VOLTAGE_DIP ||
               event_type == BLACKBOX_EVENT_FAILSAFE) {
        event->event_severity = 2;
    } else {
        event->event_severity = 1;
    }

    write_to_flash((uint8_t *)&entry, sizeof(entry.type) + sizeof(entry.size) + entry.size);
    bb_state.entry_count++;
}

uint32_t blackbox_get_entry_count(void)
{
    return bb_state.entry_count;
}

uint32_t blackbox_get_flight_id(void)
{
    return bb_state.current_flight_id;
}

bool blackbox_read_entry(uint32_t index, BlackboxLogEntry *entry)
{
    if (index >= bb_state.entry_count) {
        return false;
    }

    return true;
}

bool blackbox_read_range(uint32_t start_index, uint32_t count, uint8_t *buffer, uint32_t *bytes_read)
{
    if (!buffer || !bytes_read) {
        return false;
    }

    *bytes_read = 0;
    return true;
}

uint32_t blackbox_get_total_bytes(void)
{
    return flash_write_ptr - FLASH_DATA_START_OFFSET;
}

void blackbox_check_anomalies(void)
{
    if (!bb_state.recording) {
        return;
    }

    BatteryState battery;
    sensor_manager_get_battery(&battery);

    if (bb_state.last_voltage > 0 && battery.voltage > 0) {
        float voltage_drop = bb_state.last_voltage - battery.voltage;
        if (voltage_drop > 1.0f && battery.voltage < bb_state.last_voltage * 0.9f) {
            blackbox_log_event(BLACKBOX_EVENT_VOLTAGE_DIP,
                               (int32_t)(battery.voltage * 1000),
                               (int32_t)(bb_state.last_voltage * 1000),
                               voltage_drop,
                               0,
                               "Voltage dip detected");
        }
    }
    bb_state.last_voltage = battery.voltage;

    GPSPosition gps_pos;
    GPSVelocity gps_vel;
    sensor_manager_get_gps(&gps_pos, &gps_vel);

    uint32_t gps_time = HAL_GetTick();
    if (sensor_manager_check_gps_timeout()) {
        if (bb_state.last_gps_time > 0 && (gps_time - bb_state.last_gps_time) > 3000) {
            blackbox_log_event(BLACKBOX_EVENT_GPS_LOSS,
                               0, 0, 0, 0,
                               "GPS signal lost");
            bb_state.last_gps_time = gps_time;
        }
    } else {
        bb_state.last_gps_time = gps_time;
    }

    RCInput rc;
    sensor_manager_get_rc(&rc);
    uint32_t rc_time = HAL_GetTick();
    if (!rc.connected) {
        if (bb_state.last_rc_time > 0 && (rc_time - bb_state.last_rc_time) > 2000) {
            blackbox_log_event(BLACKBOX_EVENT_RC_LOSS,
                               0, 0, 0, 0,
                               "RC signal lost");
            bb_state.last_rc_time = rc_time;
        }
    } else {
        bb_state.last_rc_time = rc_time;
    }

    if (battery.voltage < LOW_BATTERY_RTL_TRIGGER_VOLTAGE) {
        static uint32_t last_low_bat_event = 0;
        if (rc_time - last_low_bat_event > 10000) {
            blackbox_log_event(BLACKBOX_EVENT_LOW_BATTERY,
                               (int32_t)(battery.voltage * 1000),
                               0, battery.battery_percent, 0,
                               "Low battery warning");
            last_low_bat_event = rc_time;
        }
    }
}

void blackbox_reset(void)
{
    bb_state.entry_count = 0;
    bb_state.recording = false;
    flash_write_ptr = FLASH_DATA_START_OFFSET;
    buffer_index = 0;
    memset(flash_buffer, 0, sizeof(flash_buffer));
}

void blackbox_get_info(BlackboxInfo *info)
{
    if (info == NULL) return;
    
    memset(info, 0, sizeof(BlackboxInfo));
    info->total_entries = bb_state.entry_count;
    info->total_bytes = bb_state.entry_count * sizeof(BlackboxLogEntry);
    info->start_time = bb_state.header.start_time;
    info->end_time = bb_state.header.end_time;
    info->flight_id = bb_state.current_flight_id;
    info->is_recording = bb_state.recording;
}

bool blackbox_read_data(uint32_t offset, uint8_t *buffer, uint32_t length)
{
    if (buffer == NULL || length == 0) return false;
    
    uint32_t total_data = bb_state.entry_count * sizeof(BlackboxLogEntry);
    if (offset >= total_data) return false;
    
    uint32_t available = total_data - offset;
    uint32_t to_read = (length < available) ? length : available;
    
    if (to_read > length) to_read = length;
    
    for (uint32_t i = 0; i < to_read; i++) {
        uint32_t flash_offset = FLASH_DATA_START_OFFSET + offset + i;
        if (flash_offset >= BLACKBOX_FLASH_SIZE) {
            flash_offset = FLASH_DATA_START_OFFSET + (flash_offset - BLACKBOX_FLASH_SIZE);
        }
        buffer[i] = flash_memory[flash_offset];
    }
    
    return true;
}
