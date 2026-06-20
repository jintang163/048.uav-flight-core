#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "../include/communication_link.h"
#include "../transport/transport.h"
#include "../transport/tcp_transport.h"
#include "../transport/udp_transport.h"
#include "../transport/transport_mux.h"
#include "../transport/link_monitor.h"
#include "../transport/fragment.h"
#include "../mavlink/mavlink_v2.h"
#include "../mavlink/encoder.h"
#include "../mavlink/messages/common.h"

static void on_state_change(transport_t* transport, transport_state_t new_state, void* user_data)
{
    (void)user_data;
    printf("链路状态变化: %s -> %d\n",
           transport_get_type(transport) == TRANSPORT_TYPE_TCP ? "TCP" : "UDP", new_state);
}

int main(int argc, char* argv[])
{
    (void)argc;
    (void)argv;

    printf("=== 通信链路示例 - 多链路复用\n\n");

    tcp_transport_t tcp_4g;
    tcp_config_t tcp_config;
    tcp_transport_default_config(&tcp_config);
    strcpy(tcp_config.host, "192.168.1.100");
    tcp_config.port = 5760;
    tcp_config.connect_timeout_ms = 5000;
    tcp_transport_init(&tcp_4g, &tcp_config);
    transport_set_state_callback(&tcp_4g.base, on_state_change, NULL);

    udp_transport_t udp_radio;
    udp_config_t udp_config;
    udp_transport_default_config(&udp_config);
    udp_config.local_port = 14550;
    strcpy(udp_config.remote_host, "192.168.1.100");
    udp_config.remote_port = 14550;
    udp_transport_init(&udp_radio, &udp_config);
    transport_set_state_callback(&udp_radio.base, on_state_change, NULL);

    printf("创建多链路复用器 (故障转移模式)...\n");
    transport_mux_t mux;
    transport_mux_init(&mux, MUX_MODE_FAILOVER);
    transport_mux_add_link(&mux, &tcp_4g.base, 10);
    transport_mux_add_link(&mux, &udp_radio.base, 5);
    transport_mux_set_failover_threshold(&mux, 3);
    transport_mux_set_failback_delay(&mux, 30000);

    printf("链路0: 4G/TCP (优先级10)\n");
    printf("链路1: 数传电台/UDP (优先级5)\n\n");

    link_monitor_t monitor;
    link_monitor_init(&monitor);
    link_monitor_set_interval(&monitor, 1000);
    link_monitor_set_timeout(&monitor, 2000);
    link_monitor_start(&monitor);

    fragment_manager_t frag_mgr;
    fragment_manager_init(&frag_mgr, 255, 5000);

    if (transport_connect(&mux.base) == 0) {
        printf("多链路连接成功！\n\n");

        int32_t active_link = transport_mux_get_active_link(&mux);
        printf("当前活动链路: %d\n\n", active_link);

        mavlink_message_t msg;
        uint8_t buffer[MAVLINK_MAX_PACKET_LEN];

        for (int seq = 0; seq < 100; seq++) {
            mavlink_attitude_t att;
            att.time_boot_ms = seq * 100;
            att.roll = 0.1f * seq;
            att.pitch = -0.05f * seq;
            att.yaw = 0.2f * seq;
            att.rollspeed = 0.01f;
            att.pitchspeed = -0.02f;
            att.yawspeed = 0.03f;
            mavlink_msg_attitude_encode(1, 1, &msg, &att);

            uint16_t msg_len = mavlink_msg_to_send_buffer(buffer, &msg);

            link_monitor_add_ping_sent(&monitor, seq);

            int sent = transport_send(&mux.base, buffer, msg_len, 1000);
            if (sent > 0) {
                link_monitor_add_packet_sent(&monitor, sent);

                if (seq % 10 == 9) {
                    link_monitor_add_ping_received(&monitor, seq, 150 + (rand() % 100));
                }
            } else {
                    link_monitor_add_packet_lost(&monitor, 1);
            }

            if (seq % 10 == 0) {
                link_stats_t stats;
                link_monitor_get_stats(&monitor, &stats);

                printf("=== 第 %d 次发送统计\n", seq + 1);
                printf("  活动链路: %d\n", transport_mux_get_active_link(&mux));
                printf("  RTT: 最小=%dms, 最大=%dms, 平均=%dms\n",
                       stats.rtt_min_ms, stats.rtt_max_ms, stats.rtt_avg_ms);
                printf("  丢包率: %.2f%%\n", stats.packet_loss_rate);
                printf("  链路质量评分: %d/100\n", stats.link_quality_score);
                printf("  信号强度: %d dBm\n", stats.signal_strength);
                printf("  发送速率: %u B/s, 接收速率: %u B/s\n",
                       stats.bytes_sent_per_sec, stats.bytes_received_per_sec);

                mux_link_status_t status;
                for (int i = 0; i < 2; i++) {
                    transport_mux_get_link_status(&mux, i, &status);
                    int32_t quality;
                    transport_mux_get_link_quality(&mux, i, &quality);
                    printf("  链路%d: 状态=%d, 质量=%d\n", i, status, quality);
                }
                printf("\n");

                int32_t quality = link_monitor_get_quality_score(&monitor);
                if (quality < 50 && active_link == 0) {
                    printf("警告: 链路质量差，自动切换到备用链路...\n");
                    transport_mux_manual_switch(&mux, 1);
                    active_link = 1;
                } else if (quality > 80 && active_link == 1) {
                    printf("信息: 主链路恢复，切换回主链路...\n");
                    transport_mux_manual_switch(&mux, 0);
                    active_link = 0;
                }
            }
        }

        printf("\n=== 大消息分片测试\n");
        uint8_t large_data[4096];
        for (size_t i = 0; i < sizeof(large_data); i++) {
            large_data[i] = (uint8_t)(i & 0xFF);
        }

        uint8_t fragments[8192];
        size_t fragment_count;
        if (fragment_manager_split(&frag_mgr, large_data, sizeof(large_data),
                                    fragments, &fragment_count, 64) == 0) {
            printf("数据大小: %zu 字节\n", sizeof(large_data));
            printf("分片数量: %zu 个\n", fragment_count);
            printf("MTU: %d 字节\n\n", frag_mgr.mtu);

            fragment_reassembly_t reassembly;
            fragment_reassembly_init(&reassembly, 255, 5000);

            for (size_t i = 0; i < fragment_count; i++) {
                fragment_packet_t* pkt = (fragment_packet_t*)(fragments + i * 255);
                fragment_reassembly_add(&reassembly, pkt);
            }

            if (fragment_reassembly_is_complete(&reassembly)) {
                printf("分片重组完成！\n");

                uint8_t reassembled[4096];
                size_t reassembled_len = sizeof(reassembled);
                if (fragment_reassembly_get_data(&reassembly, reassembled, &reassembled_len) == 0) {
                    printf("重组后大小: %zu 字节\n", reassembled_len);
                    printf("数据完整性: %s\n",
                           memcmp(reassembled, large_data, sizeof(large_data)) == 0 ? "验证通过" : "验证失败");
                }
            }
        }

        printf("\n=== 最终统计\n");
        printf("总发送字节: %llu\n", (unsigned long long)transport_get_bytes_sent(&mux.base));
        printf("总接收字节: %llu\n", (unsigned long long)transport_get_bytes_received(&mux.base));
        printf("总发送包数: %u\n", transport_get_packet_count_sent(&mux.base));
        printf("总接收包数: %u\n", transport_get_packet_count_received(&mux.base));
        printf("错误次数: %u\n", transport_get_error_count(&mux.base));

        printf("分片发送: %llu 个包, %llu 个分片\n",
               (unsigned long long)frag_mgr.packets_fragmented,
               (unsigned long long)frag_mgr.fragments_sent);

        transport_disconnect(&mux.base);
    } else {
        printf("连接失败！\n");
    }

    link_monitor_stop(&monitor);

    printf("\n=== 多链路复用示例完成 ===\n");

    return 0;
}
