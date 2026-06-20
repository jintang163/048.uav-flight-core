#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#include "../transport/transport.h"
#include "../transport/uart_transport.h"
#include "../transport/tcp_transport.h"
#include "../transport/udp_transport.h"
#include "../transport/transport_mux.h"
#include "../transport/link_monitor.h"
#include "../transport/fragment.h"

#define TEST_ASSERT(cond) do { if (!(cond)) { printf("FAIL: %s line %d\n", __FILE__, __LINE__); return -1; } } while(0)

static int test_transport_base(void)
{
    printf("Testing transport base... ");

    transport_t transport;
    int ret = transport_init(&transport, TRANSPORT_TYPE_TCP);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(transport_get_type(&transport) == TRANSPORT_TYPE_TCP);
    TEST_ASSERT(transport_get_state(&transport) == TRANSPORT_STATE_DISCONNECTED);
    TEST_ASSERT(transport_get_bytes_sent(&transport) == 0);
    TEST_ASSERT(transport_get_error_count(&transport) == 0);

    transport_reset_stats(&transport);
    transport_cleanup(&transport);

    printf("OK\n");
    return 0;
}

static int test_uart_transport(void)
{
    printf("Testing UART transport... ");

    uart_transport_t uart;
    uart_config_t config;

    uart_transport_default_config(&config);
    TEST_ASSERT(strcmp(config.port_name, "/dev/ttyUSB0") == 0);
    TEST_ASSERT(config.baud_rate == UART_DEFAULT_BAUD_RATE);
    TEST_ASSERT(config.data_bits == 8);

    int ret = uart_transport_init(&uart, &config);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(transport_get_type(&uart.base) == TRANSPORT_TYPE_UART);

    uart_config_t config2;
    ret = uart_transport_get_config(&uart, &config2);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(config2.baud_rate == config.baud_rate);

    ret = transport_connect(&uart.base);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(transport_get_state(&uart.base) == TRANSPORT_STATE_CONNECTED);

    const uint8_t test_data[] = "UART test data";
    int sent = transport_send(&uart.base, test_data, sizeof(test_data), 1000);
    TEST_ASSERT(sent == sizeof(test_data));

    TEST_ASSERT(transport_get_bytes_sent(&uart.base) == sizeof(test_data));

    ret = transport_disconnect(&uart.base);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(transport_get_state(&uart.base) == TRANSPORT_STATE_DISCONNECTED);

    printf("OK\n");
    return 0;
}

static int test_tcp_transport(void)
{
    printf("Testing TCP transport... ");

    tcp_transport_t tcp;
    tcp_config_t config;

    tcp_transport_default_config(&config);
    TEST_ASSERT(strcmp(config.host, "127.0.0.1") == 0);
    TEST_ASSERT(config.port == TCP_DEFAULT_PORT);

    int ret = tcp_transport_init(&tcp, &config);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(transport_get_type(&tcp.base) == TRANSPORT_TYPE_TCP);

    ret = tcp_transport_set_server_mode(&tcp, true);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(tcp.is_server == true);

    ret = transport_connect(&tcp.base);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(transport_get_state(&tcp.base) == TRANSPORT_STATE_CONNECTED);

    const uint8_t test_data[] = "TCP test message";
    int sent = transport_send(&tcp.base, test_data, sizeof(test_data), 1000);
    TEST_ASSERT(sent == sizeof(test_data));

    ret = transport_disconnect(&tcp.base);
    TEST_ASSERT(ret == 0);

    printf("OK\n");
    return 0;
}

static int test_udp_transport(void)
{
    printf("Testing UDP transport... ");

    udp_transport_t udp;
    udp_config_t config;

    udp_transport_default_config(&config);
    TEST_ASSERT(config.local_port == UDP_DEFAULT_PORT);

    int ret = udp_transport_init(&udp, &config);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(transport_get_type(&udp.base) == TRANSPORT_TYPE_UDP);

    ret = udp_transport_add_broadcast(&udp, "255.255.255.255", 14550);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(config.broadcast == false);

    ret = transport_connect(&udp.base);
    TEST_ASSERT(ret == 0);

    const uint8_t test_data[] = "UDP broadcast test";
    int sent = udp_transport_send_broadcast(&udp, test_data, sizeof(test_data));
    TEST_ASSERT(sent > 0);

    ret = transport_disconnect(&udp.base);
    TEST_ASSERT(ret == 0);

    printf("OK\n");
    return 0;
}

static int test_transport_mux(void)
{
    printf("Testing transport multiplexer... ");

    transport_mux_t mux;
    int ret = transport_mux_init(&mux, MUX_MODE_FAILOVER);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(transport_get_type(&mux.base) == TRANSPORT_TYPE_MUX);

    uart_transport_t uart1, uart2;
    uart_config_t uart_config;
    uart_transport_default_config(&uart_config);
    uart_transport_init(&uart1, &uart_config);
    strcpy(uart_config.port_name, "/dev/ttyUSB1");
    uart_transport_init(&uart2, &uart_config);

    ret = transport_mux_add_link(&mux, &uart1.base, 10);
    TEST_ASSERT(ret == 0);
    ret = transport_mux_add_link(&mux, &uart2.base, 5);
    TEST_ASSERT(ret == 0);

    ret = transport_mux_set_primary(&mux, 0);
    TEST_ASSERT(ret == 0);

    ret = transport_connect(&mux.base);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(transport_get_state(&mux.base) == TRANSPORT_STATE_CONNECTED);

    const uint8_t test_data[] = "MUX test message";
    int sent = transport_send(&mux.base, test_data, sizeof(test_data), 1000);
    TEST_ASSERT(sent == sizeof(test_data));

    int32_t active = transport_mux_get_active_link(&mux);
    TEST_ASSERT(active >= 0);

    mux_link_status_t status;
    ret = transport_mux_get_link_status(&mux, 0, &status);
    TEST_ASSERT(ret == 0);

    ret = transport_mux_manual_switch(&mux, 1);
    TEST_ASSERT(ret == 0);
    active = transport_mux_get_active_link(&mux);
    TEST_ASSERT(active == 1);

    ret = transport_disconnect(&mux.base);
    TEST_ASSERT(ret == 0);

    printf("OK\n");
    return 0;
}

static int test_link_monitor(void)
{
    printf("Testing link monitor... ");

    link_monitor_t monitor;
    int ret = link_monitor_init(&monitor);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(link_monitor_is_monitoring(&monitor) == false);
    TEST_ASSERT(link_monitor_get_quality_score(&monitor) == LINK_QUALITY_MAX);

    ret = link_monitor_start(&monitor);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(link_monitor_is_monitoring(&monitor) == true);

    for (uint32_t i = 0; i < 10; i++) {
        ret = link_monitor_add_ping_sent(&monitor, i);
        TEST_ASSERT(ret == 0);
    }

    for (uint32_t i = 0; i < 9; i++) {
        ret = link_monitor_add_ping_received(&monitor, i, 100 + i * 10);
        TEST_ASSERT(ret == 0);
    }

    link_monitor_add_packet_lost(&monitor, 1);
    link_monitor_set_signal_strength(&monitor, -75);
    link_monitor_set_signal_quality(&monitor, 80);
    link_monitor_add_packet_sent(&monitor, 1024);
    link_monitor_add_packet_received(&monitor, 512);

    link_stats_t stats;
    ret = link_monitor_get_stats(&monitor, &stats);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(stats.packets_sent == 10);
    TEST_ASSERT(stats.packets_received == 9);
    TEST_ASSERT(stats.packets_lost >= 1);
    TEST_ASSERT(stats.signal_strength == -75);
    TEST_ASSERT(stats.signal_quality == 80);

    int32_t quality = link_monitor_get_quality_score(&monitor);
    TEST_ASSERT(quality > 0 && quality <= LINK_QUALITY_MAX);

    ret = link_monitor_stop(&monitor);
    TEST_ASSERT(ret == 0);

    ret = link_monitor_reset(&monitor);
    TEST_ASSERT(ret == 0);

    printf("OK\n");
    return 0;
}

static int test_fragmentation(void)
{
    printf("Testing fragmentation/reassembly... ");

    fragment_manager_t frag_mgr;
    int ret = fragment_manager_init(&frag_mgr, 128, 5000);
    TEST_ASSERT(ret == 0);

    const size_t data_len = 512;
    uint8_t test_data[512];
    for (size_t i = 0; i < data_len; i++) {
        test_data[i] = (uint8_t)(i & 0xFF);
    }

    uint8_t fragments[1024];
    size_t fragment_count;
    ret = fragment_manager_split(&frag_mgr, test_data, data_len,
                                  fragments, &fragment_count, 10);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(fragment_count > 1);
    TEST_ASSERT(frag_mgr.packets_fragmented == 1);
    TEST_ASSERT(frag_mgr.fragments_sent == fragment_count);

    fragment_reassembly_t reassembly;
    ret = fragment_reassembly_init(&reassembly, 128, 5000);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(fragment_reassembly_get_status(&reassembly) == FRAGMENT_STATUS_INIT);

    for (size_t i = 0; i < fragment_count; i++) {
        fragment_packet_t* pkt = (fragment_packet_t*)(fragments + i * 128);
        ret = fragment_reassembly_add(&reassembly, pkt);
        TEST_ASSERT(ret == 0);
    }

    TEST_ASSERT(fragment_reassembly_is_complete(&reassembly) == 1);
    TEST_ASSERT(fragment_reassembly_get_status(&reassembly) == FRAGMENT_STATUS_COMPLETE);

    uint8_t reassembled[512];
    size_t reassembled_len = sizeof(reassembled);
    ret = fragment_reassembly_get_data(&reassembly, reassembled, &reassembled_len);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(reassembled_len == data_len);
    TEST_ASSERT(memcmp(reassembled, test_data, data_len) == 0);

    ret = fragment_reassembly_reset(&reassembly);
    TEST_ASSERT(ret == 0);

    printf("OK\n");
    return 0;
}

int main(void)
{
    printf("\n=== Transport Layer Unit Tests ===\n\n");

    int ret = 0;
    ret |= test_transport_base();
    ret |= test_uart_transport();
    ret |= test_tcp_transport();
    ret |= test_udp_transport();
    ret |= test_transport_mux();
    ret |= test_link_monitor();
    ret |= test_fragmentation();

    printf("\n=== Transport Layer Tests %s ===\n\n", ret == 0 ? "PASSED" : "FAILED");

    return ret;
}
