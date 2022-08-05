package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// EncryptAuth encrypts a plaintext byte array with authentication data byte array
// and returns the IV and CipherText as byte array
func EncryptAuth(pd, ek, ad []byte) ([]byte, error) {
	cipherText, cryptoIv, err := encryptBytesWithAuthData(pd, ek, ad)
	if err != nil {
		return []byte{}, err
	}

	ed := make([]byte, 0)
	ed = append(ed, cryptoIv[0:]...)
	ed = append(ed, cipherText[0:]...)

	return ed, nil
}

// DecryptAuth decrypts a given base64 encoded ciphertext byte array and IV byte array
// and returns the plaintext as byte array
func DecryptAuth(c, dk, ad []byte) ([]byte, error) {
	iv := c[:12]
	ct := c[12:]

	pt, err := decryptBytesWithAuthData(ct, dk, iv, ad)
	if err != nil {
		return []byte{}, err
	}

	return pt, nil
}

// encryptBytesWithAuthData internally handles the encryption and auth data handling
func encryptBytesWithAuthData(d []byte, k []byte, a []byte) ([]byte, []byte, error) {
	// Check key length
	if len(k) != 32 {
		return []byte{}, []byte{}, fmt.Errorf("invalid key size, required key size is 256 bits")
	}

	// Create a new cipher object
	cphr, err := aes.NewCipher(k)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	// Read random data for nonce
	iv, err := RandomBytes(12)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	// Create a GCM block cipher
	gcm, err := cipher.NewGCM(cphr)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	// Encrypt the data
	ct := gcm.Seal(nil, iv, d, a)
	return ct, iv, nil
}

// decryptBytesWithAuthData internally handles the decryption and auth data handling
func decryptBytesWithAuthData(d []byte, k []byte, i []byte, a []byte) ([]byte, error) {
	// Create a new cipher object
	cphr, err := aes.NewCipher(k)
	if err != nil {
		return []byte{}, err
	}

	// Create a GCM block cipher
	gcm, err := cipher.NewGCM(cphr)
	if err != nil {
		return []byte{}, err
	}

	// Encrypt the data
	plainText, err := gcm.Open(nil, i, d, a)
	if err != nil {
		return []byte{}, err
	}
	return plainText, nil
}
