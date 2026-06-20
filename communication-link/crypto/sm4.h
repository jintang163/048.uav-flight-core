#ifndef SM4_H
#define SM4_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

#define SM4_BLOCK_SIZE    16
#define SM4_KEY_SIZE      16
#define SM4_IV_SIZE       16
#define SM4_ROUNDS        32

typedef enum {
    SM4_MODE_ECB = 0,
    SM4_MODE_CBC = 1,
    SM4_MODE_CFB = 2,
    SM4_MODE_OFB = 3
} sm4_mode_t;

typedef enum {
    SM4_ENCRYPT = 0,
    SM4_DECRYPT = 1
} sm4_direction_t;

typedef struct {
    uint32_t rk[SM4_ROUNDS];
    uint8_t iv[SM4_IV_SIZE];
    sm4_mode_t mode;
    sm4_direction_t direction;
    bool iv_set;
} sm4_ctx_t;

int sm4_init(sm4_ctx_t* ctx, const uint8_t* key, size_t key_len,
             sm4_mode_t mode, sm4_direction_t direction);
int sm4_set_iv(sm4_ctx_t* ctx, const uint8_t* iv, size_t iv_len);

int sm4_encrypt_update(sm4_ctx_t* ctx, const uint8_t* input, size_t input_len,
                       uint8_t* output, size_t* output_len);
int sm4_decrypt_update(sm4_ctx_t* ctx, const uint8_t* input, size_t input_len,
                       uint8_t* output, size_t* output_len);

int sm4_encrypt_final(sm4_ctx_t* ctx, uint8_t* output, size_t* output_len);
int sm4_decrypt_final(sm4_ctx_t* ctx, uint8_t* output, size_t* output_len);

int sm4_ecb_encrypt(const uint8_t* key, const uint8_t* input, size_t input_len,
                    uint8_t* output, size_t* output_len);
int sm4_ecb_decrypt(const uint8_t* key, const uint8_t* input, size_t input_len,
                    uint8_t* output, size_t* output_len);

int sm4_cbc_encrypt(const uint8_t* key, const uint8_t* iv,
                    const uint8_t* input, size_t input_len,
                    uint8_t* output, size_t* output_len);
int sm4_cbc_decrypt(const uint8_t* key, const uint8_t* iv,
                    const uint8_t* input, size_t input_len,
                    uint8_t* output, size_t* output_len);

int sm4_cfb_encrypt(const uint8_t* key, const uint8_t* iv,
                    const uint8_t* input, size_t input_len,
                    uint8_t* output, size_t* output_len);
int sm4_cfb_decrypt(const uint8_t* key, const uint8_t* iv,
                    const uint8_t* input, size_t input_len,
                    uint8_t* output, size_t* output_len);

int sm4_ofb_encrypt(const uint8_t* key, const uint8_t* iv,
                    const uint8_t* input, size_t input_len,
                    uint8_t* output, size_t* output_len);
int sm4_ofb_decrypt(const uint8_t* key, const uint8_t* iv,
                    const uint8_t* input, size_t input_len,
                    uint8_t* output, size_t* output_len);

int sm4_pkcs7_pad(uint8_t* data, size_t data_len, size_t block_size);
int sm4_pkcs7_unpad(const uint8_t* data, size_t data_len, size_t* output_len);

void sm4_self_test(void);

#endif
