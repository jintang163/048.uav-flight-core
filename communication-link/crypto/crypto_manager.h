#ifndef CRYPTO_MANAGER_H
#define CRYPTO_MANAGER_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>
#include "sm4.h"

typedef enum {
    CRYPTO_ALGO_NONE = 0,
    CRYPTO_ALGO_SM4_ECB = 1,
    CRYPTO_ALGO_SM4_CBC = 2,
    CRYPTO_ALGO_SM4_CFB = 3,
    CRYPTO_ALGO_SM4_OFB = 4,
    CRYPTO_ALGO_AES_256_CBC = 5
} crypto_algo_t;

typedef enum {
    CRYPTO_STATE_UNINIT = 0,
    CRYPTO_STATE_INIT = 1,
    CRYPTO_STATE_KEY_EXCHANGE = 2,
    CRYPTO_STATE_READY = 3,
    CRYPTO_STATE_ERROR = 4
} crypto_state_t;

typedef struct {
    crypto_state_t state;
    crypto_algo_t algo;
    uint8_t session_key[SM4_KEY_SIZE];
    uint8_t iv[SM4_IV_SIZE];
    bool session_key_set;
    bool iv_set;
    sm4_ctx_t sm4_ctx;
    uint64_t encrypted_bytes;
    uint64_t decrypted_bytes;
    uint32_t error_count;
} crypto_manager_t;

int crypto_manager_init(crypto_manager_t* manager, crypto_algo_t algo);
int crypto_manager_set_key(crypto_manager_t* manager, const uint8_t* key, size_t key_len);
int crypto_manager_set_iv(crypto_manager_t* manager, const uint8_t* iv, size_t iv_len);
int crypto_manager_generate_iv(crypto_manager_t* manager);
int crypto_manager_encrypt(crypto_manager_t* manager, const uint8_t* plaintext, size_t plaintext_len,
                           uint8_t* ciphertext, size_t* ciphertext_len);
int crypto_manager_decrypt(crypto_manager_t* manager, const uint8_t* ciphertext, size_t ciphertext_len,
                           uint8_t* plaintext, size_t* plaintext_len);
int crypto_manager_sign(const crypto_manager_t* manager, const uint8_t* data, size_t data_len,
                        uint8_t* signature, size_t* signature_len);
int crypto_manager_verify(const crypto_manager_t* manager, const uint8_t* data, size_t data_len,
                          const uint8_t* signature, size_t signature_len);
void crypto_manager_cleanup(crypto_manager_t* manager);
crypto_state_t crypto_manager_get_state(const crypto_manager_t* manager);
uint64_t crypto_manager_get_encrypted_bytes(const crypto_manager_t* manager);
uint64_t crypto_manager_get_decrypted_bytes(const crypto_manager_t* manager);
uint32_t crypto_manager_get_error_count(const crypto_manager_t* manager);

#endif
