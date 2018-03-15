package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

const (
	aesKeySize = 32 // force 256-bit AES
)

var AES_KEY []byte

func RandBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	got, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	if n != got {
		return nil, errors.New("unable to generate enough random data")
	}
	return b, nil
}

// pad uses the PKCS#7 padding scheme to align the a payload to a specific block size
func pad(plaintext []byte, bsize int) ([]byte, error) {
	if bsize >= 256 {
		return nil, errors.New("bsize must be < 256")
	}
	pad := bsize - (len(plaintext) % bsize)
	if pad == 0 {
		pad = bsize
	}
	for i := 0; i < pad; i++ {
		plaintext = append(plaintext, byte(pad))
	}
	return plaintext, nil
}

// unpad strips the padding previously added using the PKCS#7 padding scheme
func unpad(paddedtext []byte) ([]byte, error) {
	length := len(paddedtext)
	paddedtext, lbyte := paddedtext[:length-1], paddedtext[length-1]
	pad := int(lbyte)
	if pad >= 256 || pad > length {
		return nil, errors.New("padding malformed")
	}
	return paddedtext[:length-(pad)], nil
}

// AESEncrypt encrypts a payload using AES-CBC and PKCS#7 padding.
// The returned ciphertext has three notable properties:
// 1. ciphertext is aligned to the standard AES block size
// 2. ciphertext is padded using PKCS#7
// 3. IV is prepended to the ciphertext
// This function is DEPRECATED. Use Encrypt() instead.
func CBCEncrypt(plaintext []byte) ([]byte, error) {
	plaintext, err := pad(plaintext, aes.BlockSize)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(AES_KEY)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

// AESDecrypt decrypts a payload encrypted using AES-CBC and PKCS#7 padding.
// The decryption algorithm makes three assumptions:
// 1. ciphertext is aligned to the standard AES block size
// 2. ciphertext is padded using PKCS#7
// 3. the IV is prepended to ciphertext
// This function is DEPRECATED. Use Decrypt() instead.
func CBCDecrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	block, err := aes.NewCipher(AES_KEY)
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	if len(plaintext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	return unpad(plaintext)
}

// Encrypt encrypts data using 256-bit AES-GCM.
// This both hides the content of the data and provides a check that it hasn't been altered.
// Output takes the form nonce|ciphertext|tag where '|' indicates concatenation.
func GCMEncrypt(plaintext []byte) (ciphertext []byte, err error) {
	if len(AES_KEY) != aesKeySize {
		return nil, aes.KeySizeError(len(AES_KEY))
	}

	aes, err := aes.NewCipher(AES_KEY)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, err
	}

	nonce, err := RandBytes(gcm.NonceSize())
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts data using 256-bit AES-GCM.
// This both hides the content of the data and provides a check that it hasn't been altered.
// Expects input form nonce|ciphertext|tag where '|' indicates concatenation.
func GCMDecrypt(ciphertext []byte) (plaintext []byte, err error) {
	if len(AES_KEY) != aesKeySize {
		return nil, aes.KeySizeError(len(AES_KEY))
	}

	aes, err := aes.NewCipher(AES_KEY)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():], nil)
}
