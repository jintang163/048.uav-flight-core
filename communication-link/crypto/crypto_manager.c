#include "crypto_manager.h"
#include "auth.h"
#include <string.h>
#include <stdlib.h>

int crypto_manager_init(crypto_manager_t* manager, crypto_algo_t algo)
{
    if (!manager) return -1;

    memset(manager, 0, sizeof(crypto_manager_t));
    manager->state = CRYPTO_STATE_INIT;
    manager->algo = algo;
    manager->session_key_set = false;
    manager->iv_set = false;
    manager->encrypted_bytes = 0;
    manager->decrypted_bytes = 0;
    manager->error_count = 0;

    return 0;
}

int crypto_manager_set_key(crypto_manager_t* manager, const uint8_t* key, size_t key_len)
{
    if (!manager || !key || key_len != SM4_KEY_SIZE) return -1;

    memcpy(manager->session_key, key, SM4_KEY_SIZE);
    manager->session_key_set = true;

    sm4_mode_t mode = SM4_MODE_ECB;
    switch (manager->algo) {
    case CRYPTO_ALGO_SM4_ECB: mode = SM4_MODE_ECB; break;
    case CRYPTO_ALGO_SM4_CBC: mode = SM4_MODE_CBC; break;
    case CRYPTO_ALGO_SM4_CFB: mode = SM4_MODE_CFB; break;
    case CRYPTO_ALGO_SM4_OFB: mode = SM4_MODE_OFB; break;
    default: return -1;
    }

    if (sm4_init(&manager->sm4_ctx, manager->session_key, SM4_KEY_SIZE,
                 mode, SM4_ENCRYPT) != 0) {
        manager->state = CRYPTO_STATE_ERROR;
        manager->error_count++;
        return -1;
    }

    manager->state = CRYPTO_STATE_READY;
    return 0;
}

int crypto_manager_set_iv(crypto_manager_t* manager, const uint8_t* iv, size_t iv_len)
{
    if (!manager || !iv || iv_len != SM4_IV_SIZE) return -1;
    memcpy(manager->iv, iv, SM4_IV_SIZE);
    manager->iv_set = true;
    sm4_set_iv(&manager->sm4_ctx, manager->iv, SM4_IV_SIZE);
    return 0;
}

int crypto_manager_generate_iv(crypto_manager_t* manager)
{
    if (!manager) return -1;

    for (int i = 0; i < SM4_IV_SIZE; i++) {
        manager->iv[i] = (uint8_t)(rand() & 0xFF);
    }
    manager->iv_set = true;
    sm4_set_iv(&manager->sm4_ctx, manager->iv, SM4_IV_SIZE);
    return 0;
}

static sm4_mode_t crypto_get_sm4_mode(crypto_algo_t algo)
{
    switch (algo) {
    case CRYPTO_ALGO_SM4_ECB: return SM4_MODE_ECB;
    case CRYPTO_ALGO_SM4_CBC: return SM4_MODE_CBC;
    case CRYPTO_ALGO_SM4_CFB: return SM4_MODE_CFB;
    case CRYPTO_ALGO_SM4_OFB: return SM4_MODE_OFB;
    default: return SM4_MODE_ECB;
    }
}

int crypto_manager_encrypt(crypto_manager_t* manager, const uint8_t* plaintext, size_t plaintext_len,
                           uint8_t* ciphertext, size_t* ciphertext_len)
{
    if (!manager || !plaintext || !ciphertext || !ciphertext_len) return -1;
    if (manager->state != CRYPTO_STATE_READY) return -1;
    if (!manager->session_key_set) return -1;

    size_t pad_len = 0;
    size_t buf_len = plaintext_len + SM4_BLOCK_SIZE;
    uint8_t* buf = (uint8_t*)malloc(buf_len);
    if (!buf) {
        manager->error_count++;
        manager->state = CRYPTO_STATE_ERROR;
        return -1;
    }

    memcpy(buf, plaintext, plaintext_len);
    pad_len = sm4_pkcs7_pad(buf, plaintext_len, SM4_BLOCK_SIZE);
    size_t padded_len = plaintext_len + pad_len;

    sm4_mode_t mode = crypto_get_sm4_mode(manager->algo);
    int ret = -1;

    switch (mode) {
    case SM4_MODE_ECB:
        ret = sm4_ecb_encrypt(manager->session_key, buf, padded_len, ciphertext, ciphertext_len);
        break;
    case SM4_MODE_CBC:
        if (!manager->iv_set) goto cleanup;
        ret = sm4_cbc_encrypt(manager->session_key, manager->iv, buf, padded_len,
                              ciphertext, ciphertext_len);
        break;
    case SM4_MODE_CFB:
        if (!manager->iv_set) goto cleanup;
        ret = sm4_cfb_encrypt(manager->session_key, manager->iv, buf, padded_len,
                              ciphertext, ciphertext_len);
        break;
    case SM4_MODE_OFB:
        if (!manager->iv_set) goto cleanup;
        ret = sm4_ofb_encrypt(manager->session_key, manager->iv, buf, padded_len,
                              ciphertext, ciphertext_len);
        break;
    default:
        break;
    }

cleanup:
    free(buf);

    if (ret == 0) {
        manager->encrypted_bytes += *ciphertext_len;
    } else {
        manager->error_count++;
    }

    return ret;
}

int crypto_manager_decrypt(crypto_manager_t* manager, const uint8_t* ciphertext, size_t ciphertext_len,
                           uint8_t* plaintext, size_t* plaintext_len)
{
    if (!manager || !ciphertext || !plaintext || !plaintext_len) return -1;
    if (manager->state != CRYPTO_STATE_READY) return -1;
    if (!manager->session_key_set) return -1;
    if (ciphertext_len % SM4_BLOCK_SIZE != 0) return -1;

    uint8_t* buf = (uint8_t*)malloc(ciphertext_len);
    if (!buf) {
        manager->error_count++;
        manager->state = CRYPTO_STATE_ERROR;
        return -1;
    }

    size_t decrypted_len = 0;
    sm4_mode_t mode = crypto_get_sm4_mode(manager->algo);
    int ret = -1;

    switch (mode) {
    case SM4_MODE_ECB:
        ret = sm4_ecb_decrypt(manager->session_key, ciphertext, ciphertext_len, buf, &decrypted_len);
        break;
    case SM4_MODE_CBC:
        if (!manager->iv_set) goto cleanup;
        ret = sm4_cbc_decrypt(manager->session_key, manager->iv, ciphertext, ciphertext_len,
                              buf, &decrypted_len);
        break;
    case SM4_MODE_CFB:
        if (!manager->iv_set) goto cleanup;
        ret = sm4_cfb_decrypt(manager->session_key, manager->iv, ciphertext, ciphertext_len,
                              buf, &decrypted_len);
        break;
    case SM4_MODE_OFB:
        if (!manager->iv_set) goto cleanup;
        ret = sm4_ofb_decrypt(manager->session_key, manager->iv, ciphertext, ciphertext_len,
                              buf, &decrypted_len);
        break;
    default:
        break;
    }

    if (ret == 0) {
        size_t unpadded_len = 0;
        ret = sm4_pkcs7_unpad(buf, decrypted_len, &unpadded_len);
        if (ret == 0) {
            memcpy(plaintext, buf, unpadded_len);
            *plaintext_len = unpadded_len;
            manager->decrypted_bytes += ciphertext_len;
        } else {
            manager->error_count++;
        }
    } else {
        manager->error_count++;
    }

cleanup:
    free(buf);
    return ret;
}

int crypto_manager_sign(const crypto_manager_t* manager, const uint8_t* data, size_t data_len,
                        uint8_t* signature, size_t* signature_len)
{
    if (!manager || !data || !signature || !signature_len) return -1;
    if (*signature_len < HMAC_SM3_SIZE) return -1;

    hmac_sm3(manager->session_key, SM4_KEY_SIZE, data, data_len, signature);
    *signature_len = HMAC_SM3_SIZE;
    return 0;
}

int crypto_manager_verify(const crypto_manager_t* manager, const uint8_t* data, size_t data_len,
                          const uint8_t* signature, size_t signature_len)
{
    if (!manager || !data || !signature) return -1;

    return hmac_sm3_verify(manager->session_key, SM4_KEY_SIZE,
                           data, data_len, signature, signature_len) ? 0 : -1;
}

void crypto_manager_cleanup(crypto_manager_t* manager)
{
    if (!manager) return;
    memset(manager->session_key, 0, SM4_KEY_SIZE);
    memset(manager->iv, 0, SM4_IV_SIZE);
    manager->state = CRYPTO_STATE_UNINIT;
    manager->session_key_set = false;
    manager->iv_set = false;
}

crypto_state_t crypto_manager_get_state(const crypto_manager_t* manager)
{
    return manager ? manager->state : CRYPTO_STATE_UNINIT;
}

uint64_t crypto_manager_get_encrypted_bytes(const crypto_manager_t* manager)
{
    return manager ? manager->encrypted_bytes : 0;
}

uint64_t crypto_manager_get_decrypted_bytes(const crypto_manager_t* manager)
{
    return manager ? manager->decrypted_bytes : 0;
}

uint32_t crypto_manager_get_error_count(const crypto_manager_t* manager)
{
    return manager ? manager->error_count : 0;
}
