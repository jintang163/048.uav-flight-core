#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include "../include/communication_link.h"
#include "../mavlink/mavlink_v2.h"
#include "../mavlink/parser.h"
#include "../mavlink/encoder.h"
#include "../mavlink/messages/common.h"
#include "../mavlink/heartbeat.h"
#include "../transport/transport.h"
#include "../transport/tcp_transport.h"

static void on_heartbeat(uint8_t sysid, uint8_t compid, const mavlink_heartbeat_t* hb, void* user_data)
{
    (void)sysid;
    (void)compid;
    (void)user_data;
    printf("收到心跳: 类型=%d, 自动驾驶=%d, 系统状态=%d\n",
           hb->type, hb->autopilot, hb->system_status);
}

static void on_attitude(uint8_t sysid, uint8_t compid, const mavlink_attitude_t* att, void* user_data)
{
    (void)sysid;
    (void)compid;
    (void)user_data;
    printf("姿态: 滚转=%.4f, 俯仰=%.4f, 偏航=%.4f\n",
           att->roll, att->pitch, att->yaw);
}

static void on_position(uint8_t sysid, uint8_t compid, const mavlink_global_position_int_t* pos, void* user_data)
{
    (void)sysid;
    (void)compid;
    (void)user_data;
    printf("位置: 纬度=%d, 经度=%d, 高度=%d, 航向=%d\n",
           pos->lat, pos->lon, pos->alt, pos->hdg);
}

int main(int argc, char* argv[])
{
    (void)argc;
    (void)argv;

    printf("=== 通信链路示例 - MAVLink通信\n");
    printf("版本: %d.%d.%d\n\n",
           COMMUNICATION_LINK_VERSION_MAJOR,
           COMMUNICATION_LINK_VERSION_MINOR,
           COMMUNICATION_LINK_VERSION_PATCH);

    tcp_transport_t tcp_transport;
    tcp_config_t tcp_config;
    tcp_transport_default_config(&tcp_config);
    strcpy(tcp_config.host, "127.0.0.1");
    tcp_config.port = 5760;
    tcp_transport_init(&tcp_transport, &tcp_config);

    mavlink_parser_t parser;
    mavlink_parser_init(&parser);

    heartbeat_manager_t hb_mgr;
    heartbeat_init(&hb_mgr, 1000, 3000, 10000);
    heartbeat_set_callback(&hb_mgr, on_heartbeat, NULL);

    parser.msg_callback[MAVLINK_MSG_ID_ATTITUDE] = (mavlink_msg_callback_t)on_attitude;
    parser.msg_callback[MAVLINK_MSG_ID_GLOBAL_POSITION_INT] = (mavlink_msg_callback_t)on_position;

    printf("正在连接到 %s:%d...\n", tcp_config.host, tcp_config.port);

    if (transport_connect(&tcp_transport.base) != 0) {
        printf("连接成功！\n\n");

        mavlink_message_t msg;
        uint8_t send_buffer[MAVLINK_MAX_PACKET_LEN];
        uint8_t recv_buffer[4096];

        uint32_t last_heartbeat_send = 0;
        uint32_t last_heartbeat_check = 0;

        while (1) {
            uint32_t now = 0;

            if (now - last_heartbeat_send >= 1000) {
                mavlink_heartbeat_t hb;
                hb.type = 2;
                hb.autopilot = 12;
                hb.base_mode = 89;
                hb.custom_mode = 0;
                hb.system_status = 4;
                hb.mavlink_version = 3;
                mavlink_msg_heartbeat_encode(1, 1, &msg, &hb);

                uint16_t len = mavlink_msg_to_send_buffer(send_buffer, &msg);
                transport_send(&tcp_transport.base, send_buffer, len, 1000);

                last_heartbeat_send = now;
                printf("发送心跳...\n");
            }

            size_t recv_len;
            int ret = transport_recv(&tcp_transport.base, recv_buffer, sizeof(recv_buffer),
                                  &recv_len, 100);
            if (ret == 0 && recv_len > 0) {
                for (size_t i = 0; i < recv_len; i++) {
                    if (mavlink_parse_byte(&parser, recv_buffer[i], &msg)) {
                        printf("收到消息 ID=%d, SYS=%d, COMP=%d, LEN=%d\n",
                               msg.msgid, msg.sysid, msg.compid, msg.len);

                        if (msg.msgid == MAVLINK_MSG_ID_HEARTBEAT) {
                            mavlink_heartbeat_t hb;
                            mavlink_msg_heartbeat_decode(&msg, &hb);
                            heartbeat_update(&hb_mgr, msg.sysid, msg.compid, &hb, now);
                        }

                        if (parser.msg_callback[msg.msgid]) {
                            parser.msg_callback[msg.msgid](msg.sysid, msg.compid,
                                                           mavlink_msg_payload(&msg), NULL);
                        }
                    }
                }
            }

            if (now - last_heartbeat_check >= 100) {
                heartbeat_check_timeout(&hb_mgr, now);

                link_status_t status = heartbeat_get_status(&hb_mgr, 1, 1);
                if (status == LINK_STATUS_WARNING) {
                    printf("警告: 心跳超时！\n");
                } else if (status == LINK_STATUS_LOST) {
                    printf("错误: 链接丢失！\n");
                }

                last_heartbeat_check = now;
            }

            usleep(10000);
        }

        transport_disconnect(&tcp_transport.base);
    } else {
        printf("连接失败！\n");
    }

    return 0;
}
