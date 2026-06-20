#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#include "../mavlink/mavlink_v2.h"
#include "../mavlink/parser.h"
#include "../mavlink/encoder.h"
#include "../mavlink/messages/common.h"

#define TEST_ASSERT(cond) do { if (!(cond)) { printf("FAIL: %s line %d\n", __FILE__, __LINE__); return -1; } } while(0)

static int test_heartbeat_encode_decode(void)
{
    printf("Testing HEARTBEAT encode/decode... ");

    mavlink_message_t msg;
    mavlink_heartbeat_t hb_send, hb_recv;

    hb_send.type = 2;
    hb_send.autopilot = 12;
    hb_send.base_mode = 89;
    hb_send.custom_mode = 0x12345678;
    hb_send.system_status = 4;
    hb_send.mavlink_version = 3;

    mavlink_msg_heartbeat_encode(1, 1, &msg, &hb_send);

    TEST_ASSERT(msg.msgid == MAVLINK_MSG_ID_HEARTBEAT);
    TEST_ASSERT(msg.sysid == 1);
    TEST_ASSERT(msg.compid == 1);

    mavlink_msg_heartbeat_decode(&msg, &hb_recv);

    TEST_ASSERT(hb_recv.type == hb_send.type);
    TEST_ASSERT(hb_recv.autopilot == hb_send.autopilot);
    TEST_ASSERT(hb_recv.base_mode == hb_send.base_mode);
    TEST_ASSERT(hb_recv.custom_mode == hb_send.custom_mode);
    TEST_ASSERT(hb_recv.system_status == hb_send.system_status);
    TEST_ASSERT(hb_recv.mavlink_version == hb_send.mavlink_version);

    printf("OK\n");
    return 0;
}

static int test_attitude_encode_decode(void)
{
    printf("Testing ATTITUDE encode/decode... ");

    mavlink_message_t msg;
    mavlink_attitude_t att_send, att_recv;

    att_send.time_boot_ms = 1234567;
    att_send.roll = 0.1234f;
    att_send.pitch = -0.5678f;
    att_send.yaw = 1.2345f;
    att_send.rollspeed = 0.01f;
    att_send.pitchspeed = -0.02f;
    att_send.yawspeed = 0.03f;

    mavlink_msg_attitude_encode(1, 1, &msg, &att_send);

    TEST_ASSERT(msg.msgid == MAVLINK_MSG_ID_ATTITUDE);

    mavlink_msg_attitude_decode(&msg, &att_recv);

    TEST_ASSERT(att_recv.time_boot_ms == att_send.time_boot_ms);
    TEST_ASSERT(fabs(att_recv.roll - att_send.roll) < 0.0001f);
    TEST_ASSERT(fabs(att_recv.pitch - att_send.pitch) < 0.0001f);
    TEST_ASSERT(fabs(att_recv.yaw - att_send.yaw) < 0.0001f);

    printf("OK\n");
    return 0;
}

static int test_global_position_encode_decode(void)
{
    printf("Testing GLOBAL_POSITION_INT encode/decode... ");

    mavlink_message_t msg;
    mavlink_global_position_int_t pos_send, pos_recv;

    pos_send.time_boot_ms = 7654321;
    pos_send.lat = 312345678;
    pos_send.lon = 1214567890;
    pos_send.alt = 100000;
    pos_send.relative_alt = 50000;
    pos_send.vx = 1000;
    pos_send.vy = -500;
    pos_send.vz = -200;
    pos_send.hdg = 18000;

    mavlink_msg_global_position_int_encode(1, 1, &msg, &pos_send);

    TEST_ASSERT(msg.msgid == MAVLINK_MSG_ID_GLOBAL_POSITION_INT);

    mavlink_msg_global_position_int_decode(&msg, &pos_recv);

    TEST_ASSERT(pos_recv.time_boot_ms == pos_send.time_boot_ms);
    TEST_ASSERT(pos_recv.lat == pos_send.lat);
    TEST_ASSERT(pos_recv.lon == pos_send.lon);
    TEST_ASSERT(pos_recv.alt == pos_send.alt);

    printf("OK\n");
    return 0;
}

static int test_parser_byte_stream(void)
{
    printf("Testing MAVLink parser byte stream... ");

    mavlink_parser_t parser;
    mavlink_message_t msg;
    mavlink_heartbeat_t hb_send, hb_recv;

    mavlink_parser_init(&parser);

    hb_send.type = 2;
    hb_send.autopilot = 12;
    hb_send.base_mode = 89;
    hb_send.custom_mode = 0x12345678;
    hb_send.system_status = 4;
    hb_send.mavlink_version = 3;

    mavlink_msg_heartbeat_encode(1, 1, &msg, &hb_send);

    uint8_t bytes[MAVLINK_MAX_PACKET_LEN];
    uint16_t len = mavlink_msg_to_send_buffer(bytes, &msg);

    bool msg_received = false;
    for (uint16_t i = 0; i < len; i++) {
        if (mavlink_parse_byte(&parser, bytes[i], &msg)) {
            msg_received = true;
            break;
        }
    }

    TEST_ASSERT(msg_received);
    TEST_ASSERT(msg.msgid == MAVLINK_MSG_ID_HEARTBEAT);

    mavlink_msg_heartbeat_decode(&msg, &hb_recv);
    TEST_ASSERT(hb_recv.type == hb_send.type);
    TEST_ASSERT(hb_recv.autopilot == hb_send.autopilot);
    TEST_ASSERT(hb_recv.system_status == hb_send.system_status);

    printf("OK\n");
    return 0;
}

static int test_command_encode_decode(void)
{
    printf("Testing COMMAND_LONG encode/decode... ");

    mavlink_message_t msg;
    mavlink_command_long_t cmd_send, cmd_recv;

    cmd_send.target_system = 1;
    cmd_send.target_component = 1;
    cmd_send.command = MAV_CMD_COMPONENT_ARM_DISARM;
    cmd_send.confirmation = 0;
    cmd_send.param1 = 1.0f;
    cmd_send.param2 = 0.0f;
    cmd_send.param3 = 0.0f;
    cmd_send.param4 = 0.0f;
    cmd_send.param5 = 0.0f;
    cmd_send.param6 = 0.0f;
    cmd_send.param7 = 0.0f;

    mavlink_msg_command_long_encode(255, 0, &msg, &cmd_send);

    TEST_ASSERT(msg.msgid == MAVLINK_MSG_ID_COMMAND_LONG);

    mavlink_msg_command_long_decode(&msg, &cmd_recv);

    TEST_ASSERT(cmd_recv.target_system == cmd_send.target_system);
    TEST_ASSERT(cmd_recv.command == cmd_send.command);
    TEST_ASSERT(fabs(cmd_recv.param1 - cmd_send.param1) < 0.0001f);

    printf("OK\n");
    return 0;
}

static int test_mission_encode_decode(void)
{
    printf("Testing MISSION_ITEM encode/decode... ");

    mavlink_message_t msg;
    mavlink_mission_item_t wp_send, wp_recv;

    wp_send.target_system = 1;
    wp_send.target_component = 1;
    wp_send.seq = 5;
    wp_send.frame = 3;
    wp_send.command = 16;
    wp_send.current = 0;
    wp_send.autocontinue = 1;
    wp_send.param1 = 0.0f;
    wp_send.param2 = 0.0f;
    wp_send.param3 = 0.0f;
    wp_send.param4 = 0.0f;
    wp_send.x = 31.234567f;
    wp_send.y = 121.456789f;
    wp_send.z = 100.0f;

    mavlink_msg_mission_item_encode(255, 0, &msg, &wp_send);

    TEST_ASSERT(msg.msgid == MAVLINK_MSG_ID_MISSION_ITEM);

    mavlink_msg_mission_item_decode(&msg, &wp_recv);

    TEST_ASSERT(wp_recv.seq == wp_send.seq);
    TEST_ASSERT(wp_recv.command == wp_send.command);
    TEST_ASSERT(fabs(wp_recv.x - wp_send.x) < 0.0001f);
    TEST_ASSERT(fabs(wp_recv.y - wp_send.y) < 0.0001f);
    TEST_ASSERT(fabs(wp_recv.z - wp_send.z) < 0.0001f);

    printf("OK\n");
    return 0;
}

static int test_crc_calculation(void)
{
    printf("Testing CRC calculation... ");

    mavlink_checksum_t crc;
    mavlink_crc_init(&crc);

    const uint8_t test_data[] = {0x01, 0x02, 0x03, 0x04, 0x05};
    for (size_t i = 0; i < sizeof(test_data); i++) {
        mavlink_crc_accumulate(test_data[i], &crc);
    }

    uint16_t crc_result = mavlink_crc_finish(&crc);
    TEST_ASSERT(crc_result != 0);

    mavlink_checksum_t crc2;
    mavlink_crc_init(&crc2);
    for (size_t i = 0; i < sizeof(test_data); i++) {
        mavlink_crc_accumulate(test_data[i], &crc2);
    }
    TEST_ASSERT(mavlink_crc_finish(&crc2) == crc_result);

    printf("OK\n");
    return 0;
}

static int test_battery_status_encode_decode(void)
{
    printf("Testing BATTERY_STATUS encode/decode... ");

    mavlink_message_t msg;
    mavlink_battery_status_t bat_send, bat_recv;

    bat_send.id = 0;
    bat_send.battery_function = 0;
    bat_send.type = 3;
    bat_send.temperature = 2500;
    bat_send.voltages[0] = 12600;
    bat_send.voltages[1] = 12600;
    bat_send.current_battery = -150;
    bat_send.current_consumed = 5000;
    bat_send.energy_consumed = 10000;
    bat_send.battery_remaining = 85;
    bat_send.time_remaining = 3600;
    bat_send.charge_state = 3;

    mavlink_msg_battery_status_encode(1, 1, &msg, &bat_send);

    TEST_ASSERT(msg.msgid == MAVLINK_MSG_ID_BATTERY_STATUS);

    mavlink_msg_battery_status_decode(&msg, &bat_recv);

    TEST_ASSERT(bat_recv.id == bat_send.id);
    TEST_ASSERT(bat_recv.voltages[0] == bat_send.voltages[0]);
    TEST_ASSERT(bat_recv.current_battery == bat_send.current_battery);
    TEST_ASSERT(bat_recv.battery_remaining == bat_send.battery_remaining);

    printf("OK\n");
    return 0;
}

int main(void)
{
    printf("\n=== MAVLink Parser Unit Tests ===\n\n");

    int ret = 0;
    ret |= test_heartbeat_encode_decode();
    ret |= test_attitude_encode_decode();
    ret |= test_global_position_encode_decode();
    ret |= test_parser_byte_stream();
    ret |= test_command_encode_decode();
    ret |= test_mission_encode_decode();
    ret |= test_crc_calculation();
    ret |= test_battery_status_encode_decode();

    printf("\n=== MAVLink Parser Tests %s ===\n\n", ret == 0 ? "PASSED" : "FAILED");

    return ret;
}
