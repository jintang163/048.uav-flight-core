package crypto

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/tjfoc/gmsm/sm4"
)

type SM4Cipher struct {
	key []byte
}

type PaddingMode string

const (
	PaddingPKCS7 PaddingMode = "pkcs7"
	PaddingZero  PaddingMode = "zero"
)

const (
	SM4KeySize = 16
	SM4BlockSize = 16
)

func NewSM4Cipher(keyHex string) (*SM4Cipher, error) {
	var key []byte
	var err error

	if keyHex == "" {
		keyHex = os.Getenv("SM4_KEY")
	}

	if keyHex == "" {
		return nil, errors.New("SM4 key not provided")
	}

	key, err = hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid SM4 key hex: %w", err)
	}

	if len(key) != SM4KeySize {
		return nil, fmt.Errorf("SM4 key must be %d bytes, got %d", SM4KeySize, len(key))
	}

	return &SM4Cipher{key: key}, nil
}

func (c *SM4Cipher) EncryptECB(plaintext []byte, padding PaddingMode) ([]byte, error) {
	block, err := sm4.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	plaintext = pad(plaintext, block.BlockSize(), padding)
	ciphertext := make([]byte, len(plaintext))

	for i := 0; i < len(plaintext); i += block.BlockSize() {
		block.Encrypt(ciphertext[i:i+block.BlockSize()], plaintext[i:i+block.BlockSize()])
	}

	return ciphertext, nil
}

func (c *SM4Cipher) DecryptECB(ciphertext []byte, padding PaddingMode) ([]byte, error) {
	block, err := sm4.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext)%block.BlockSize() != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	plaintext := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += block.BlockSize() {
		block.Decrypt(plaintext[i:i+block.BlockSize()], ciphertext[i:i+block.BlockSize()])
	}

	plaintext = unpad(plaintext, padding)
	return plaintext, nil
}

func (c *SM4Cipher) EncryptCBC(plaintext []byte, padding PaddingMode) ([]byte, []byte, error) {
	block, err := sm4.NewCipher(c.key)
	if err != nil {
		return nil, nil, err
	}

	plaintext = pad(plaintext, block.BlockSize(), padding)
	iv := make([]byte, block.BlockSize())
	if _, err := rand.Read(iv); err != nil {
		return nil, nil, err
	}

	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	return ciphertext, iv, nil
}

func (c *SM4Cipher) DecryptCBC(ciphertext, iv []byte, padding PaddingMode) ([]byte, error) {
	block, err := sm4.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext)%block.BlockSize() != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	if len(iv) != block.BlockSize() {
		return nil, errors.New("IV length must equal block size")
	}

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	plaintext = unpad(plaintext, padding)
	return plaintext, nil
}

func (c *SM4Cipher) EncryptToString(plaintext string) (string, error) {
	data, iv, err := c.EncryptCBC([]byte(plaintext), PaddingPKCS7)
	if err != nil {
		return "", err
	}
	result := make([]byte, len(iv)+len(data))
	copy(result, iv)
	copy(result[len(iv):], data)
	return base64.StdEncoding.EncodeToString(result), nil
}

func (c *SM4Cipher) DecryptFromString(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	if len(data) < SM4BlockSize {
		return "", errors.New("invalid ciphertext length")
	}
	iv := data[:SM4BlockSize]
	cipherData := data[SM4BlockSize:]
	plaintext, err := c.DecryptCBC(cipherData, iv, PaddingPKCS7)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func pad(data []byte, blockSize int, mode PaddingMode) []byte {
	padding := blockSize - len(data)%blockSize
	switch mode {
	case PaddingPKCS7:
		padBytes := bytes.Repeat([]byte{byte(padding)}, padding)
		return append(data, padBytes...)
	case PaddingZero:
		padBytes := bytes.Repeat([]byte{0}, padding)
		return append(data, padBytes...)
	default:
		return data
	}
}

func unpad(data []byte, mode PaddingMode) []byte {
	length := len(data)
	switch mode {
	case PaddingPKCS7:
		unpadding := int(data[length-1])
		if unpadding < 1 || unpadding > length {
			return data
		}
		return data[:length-unpadding]
	case PaddingZero:
		return bytes.TrimRight(data, "\x00")
	default:
		return data
	}
}
