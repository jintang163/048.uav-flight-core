#include "tls_manager.h"
#include "sm4.h"
#include "auth.h"
#include <string.h>
#include <stdlib.h>

int tls_manager_init(tls_manager_t* tls, tls_version_t version, bool is_server)
{
    if (!tls) return -1;

    memset(tls, 0, sizeof(tls_manager_t));
    tls->state = TLS_STATE_INIT;
    tls->version = version;
    tls->is_server = is_server;
    tls->use_sm4 = true;
    tls->timeout_ms = TLS_HANDSHAKE_TIMEOUT;
    tls->bytes_sent = 0;
    tls->bytes_received = 0;

    return 0;
}

int tls_manager_configure(tls_manager_t* tls, tls_cipher_suite_t cipher, bool use_sm4)
{
    if (!tls) return -1;
    tls->cipher_suite = cipher;
    tls->use_sm4 = use_sm4;
    return 0;
}

int tls_manager_set_callbacks(tls_manager_t* tls, tls_send_cb_t send_cb,
                              tls_recv_cb_t recv_cb, void* user_data)
{
    if (!tls) return -1;
    (void)send_cb;
    (void)recv_cb;
    tls->user_data = user_data;
    return 0;
}

int tls_manager_handshake(tls_manager_t* tls)
{
    if (!tls) return -1;
    if (tls->state != TLS_STATE_INIT) return -1;

    tls->state = TLS_STATE_HANDSHAKE;

    for (int i = 0; i < 32; i++) {
        tls->client_random[i] = (uint8_t)(rand() & 0xFF);
        tls->server_random[i] = (uint8_t)(rand() & 0xFF);
    }

    for (int i = 0; i < 32; i++) {
        tls->session_key[i] = tls->client_random[i] ^ tls->server_random[i];
    }

    tls->state = TLS_STATE_CONNECTED;
    return 0;
}

int tls_manager_encrypt(tls_manager_t* tls, const uint8_t* plaintext, size_t plaintext_len,
                        uint8_t* ciphertext, size_t* ciphertext_len)
{
    if (!tls || !plaintext || !ciphertext || !ciphertext_len) return -1;
    if (tls->state != TLS_STATE_CONNECTED) return -1;

    size_t pad_len = SM4_BLOCK_SIZE - (plaintext_len % SM4_BLOCK_SIZE);
    size_t padded_len = plaintext_len + pad_len;

    if (*ciphertext_len < padded_len + HMAC_SM3_SIZE) {
        return -1;
    }

    uint8_t* buf = (uint8_t*)malloc(padded_len);
    if (!buf) return -1;

    memcpy(buf, plaintext, plaintext_len);
    sm4_pkcs7_pad(buf, plaintext_len, SM4_BLOCK_SIZE);

    uint8_t iv[SM4_IV_SIZE];
    for (int i = 0; i < SM4_IV_SIZE; i++) {
        iv[i] = (uint8_t)(rand() & 0xFF);
    }

    size_t encrypted_len = 0;
    sm4_cbc_encrypt(tls->session_key, iv, buf, padded_len,
                    ciphertext + SM4_IV_SIZE, &encrypted_len);

    memcpy(ciphertext, iv, SM4_IV_SIZE);

    uint8_t hmac[HMAC_SM3_SIZE];
    hmac_sm3(tls->session_key, 32, ciphertext, SM4_IV_SIZE + encrypted_len, hmac);
    memcpy(ciphertext + SM4_IV_SIZE + encrypted_len, hmac, HMAC_SM3_SIZE);

    *ciphertext_len = SM4_IV_SIZE + encrypted_len + HMAC_SM3_SIZE;
    tls->bytes_sent += *ciphertext_len;

    free(buf);
    return 0;
}

int tls_manager_decrypt(tls_manager_t* tls, const uint8_t* ciphertext, size_t ciphertext_len,
                        uint8_t* plaintext, size_t* plaintext_len)
{
    if (!tls || !ciphertext || !plaintext || !plaintext_len) return -1;
    if (tls->state != TLS_STATE_CONNECTED) return -1;
    if (ciphertext_len < SM4_IV_SIZE + SM4_BLOCK_SIZE + HMAC_SM3_SIZE) return -1;

    size_t hmac_offset = ciphertext_len - HMAC_SM3_SIZE;
    uint8_t expected_hmac[HMAC_SM3_SIZE];
    hmac_sm3(tls->session_key, 32, ciphertext, hmac_offset, expected_hmac);

    uint8_t diff = 0;
    for (int i = 0; i < HMAC_SM3_SIZE; i++) {
        diff |= expected_hmac[i] ^ ciphertext[hmac_offset + i];
    }
    if (diff != 0) {
        return -1;
    }

    uint8_t iv[SM4_IV_SIZE];
    memcpy(iv, ciphertext, SM4_IV_SIZE);

    size_t encrypted_len = hmac_offset - SM4_IV_SIZE;
    uint8_t* buf = (uint8_t*)malloc(encrypted_len);
    if (!buf) return -1;

    size_t decrypted_len = 0;
    int ret = sm4_cbc_decrypt(tls->session_key, iv, ciphertext + SM4_IV_SIZE,
                              encrypted_len, buf, &decrypted_len);

    if (ret == 0) {
        size_t unpadded_len = 0;
        ret = sm4_pkcs7_unpad(buf, decrypted_len, &unpadded_len);
        if (ret == 0) {
            memcpy(plaintext, buf, unpadded_len);
            *plaintext_len = unpadded_len;
            tls->bytes_received += ciphertext_len;
        }
    }

    free(buf);
    return ret;
}

int tls_manager_close(tls_manager_t* tls)
{
    if (!tls) return -1;
    if (tls->state != TLS_STATE_CONNECTED) return -1;
    tls->state = TLS_STATE_CLOSED;
    return 0;
}

void tls_manager_cleanup(tls_manager_t* tls)
{
    if (!tls) return;
    memset(tls->session_key, 0, sizeof(tls->session_key));
    memset(tls->client_random, 0, sizeof(tls->client_random));
    memset(tls->server_random, 0, sizeof(tls->server_random));
    tls->state = TLS_STATE_CLOSED;
}

tls_state_t tls_manager_get_state(const tls_manager_t* tls)
{
    return tls ? tls->state : TLS_STATE_ERROR;
}

uint64_t tls_manager_get_bytes_sent(const tls_manager_t* tls)
{
    return tls ? tls->bytes_sent : 0;
}

uint64_t tls_manager_get_bytes_received(const tls_manager_t* tls)
{
    return tls ? tls->bytes_received : 0;
}
