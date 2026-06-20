#ifndef TLS_MANAGER_H
#define TLS_MANAGER_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

#define TLS_MAX_RECORD_SIZE    16384
#define TLS_HANDSHAKE_TIMEOUT  30000

typedef enum {
    TLS_STATE_INIT = 0,
    TLS_STATE_HANDSHAKE = 1,
    TLS_STATE_CONNECTED = 2,
    TLS_STATE_CLOSING = 3,
    TLS_STATE_CLOSED = 4,
    TLS_STATE_ERROR = 5
} tls_state_t;

typedef enum {
    TLS_VERSION_1_2 = 0x0303,
    TLS_VERSION_1_3 = 0x0304
} tls_version_t;

typedef enum {
    TLS_CIPHER_TLS_SM4_GCM_SM3 = 0,
    TLS_CIPHER_TLS_ECDHE_SM4_CBC_SM3 = 1,
    TLS_CIPHER_TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384 = 2
} tls_cipher_suite_t;

typedef struct {
    tls_state_t state;
    tls_version_t version;
    tls_cipher_suite_t cipher_suite;
    uint8_t session_key[32];
    uint8_t client_random[32];
    uint8_t server_random[32];
    bool is_server;
    bool use_sm4;
    uint64_t bytes_sent;
    uint64_t bytes_received;
    uint32_t handshake_start_time;
    uint32_t timeout_ms;
    void* user_data;
} tls_manager_t;

typedef int (*tls_send_cb_t)(const uint8_t* data, size_t len, void* user_data);
typedef int (*tls_recv_cb_t)(uint8_t* data, size_t len, size_t* recv_len, void* user_data);

int tls_manager_init(tls_manager_t* tls, tls_version_t version, bool is_server);
int tls_manager_configure(tls_manager_t* tls, tls_cipher_suite_t cipher, bool use_sm4);
int tls_manager_set_callbacks(tls_manager_t* tls, tls_send_cb_t send_cb,
                              tls_recv_cb_t recv_cb, void* user_data);
int tls_manager_handshake(tls_manager_t* tls);
int tls_manager_encrypt(tls_manager_t* tls, const uint8_t* plaintext, size_t plaintext_len,
                        uint8_t* ciphertext, size_t* ciphertext_len);
int tls_manager_decrypt(tls_manager_t* tls, const uint8_t* ciphertext, size_t ciphertext_len,
                        uint8_t* plaintext, size_t* plaintext_len);
int tls_manager_close(tls_manager_t* tls);
void tls_manager_cleanup(tls_manager_t* tls);
tls_state_t tls_manager_get_state(const tls_manager_t* tls);
uint64_t tls_manager_get_bytes_sent(const tls_manager_t* tls);
uint64_t tls_manager_get_bytes_received(const tls_manager_t* tls);

#endif
