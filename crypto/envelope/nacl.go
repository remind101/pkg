package envelope

import (
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/nacl/secretbox"
)

// SecretBoxCrypt is a Crypt implementation that uses NaCl secretbox to encrypt
// values. It generates random data encryption keys and uses an embedded Crypt to
// encrypt/decrypt these keys.
//
// A key design decision here is that this object never has direct access to the
// encryption keys (KEK's) used to encrypt DEK's.
type SecretBoxCrypt struct {
	GenerateRandomKey func() [32]byte

	// A Crypt used to generate/decrypt Data Encryption Keys used by NaCL.
	encryptionKeys Crypt
}

func NewSecretBoxCrypt(keyCrypt Crypt) *SecretBoxCrypt {
	return &SecretBoxCrypt{
		encryptionKeys: keyCrypt,
	}
}

// Encrypt performs envelope encryption on the plaintext.
func (c *SecretBoxCrypt) Encrypt(plaintext []byte) (*Envelope, error) {
	key, err := c.encryptionKeys.GenerateDataKey()
	if err != nil {
		return nil, err
	}

	var secretKey [32]byte
	if n := copy(secretKey[:], key.Plaintext); n != 32 {
		panic("expected key size to be 256 bits")
	}

	ciphertext, err := seal(plaintext, secretKey)
	if err != nil {
		return nil, err
	}

	return &Envelope{
		EncryptionKey: key.Envelope,
		Ciphertext:    ciphertext,
	}, nil
}

// Decrypt decrypts the Envelope.
func (c *SecretBoxCrypt) Decrypt(m *Envelope) ([]byte, error) {
	// For messages encrypted with this object, the EncryptionKey in the
	// envelope with allways be another Envelope.
	encryptionKey, ok := m.EncryptionKey.(*Envelope)
	if !ok {
		return nil, fmt.Errorf("unable to decrypt envelope")
	}

	keyBytes, err := c.encryptionKeys.Decrypt(encryptionKey)
	if err != nil {
		return nil, err
	}

	var key [32]byte
	copy(key[:], keyBytes)

	plaintext, err := open(m.Ciphertext, key)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GenerateDataKey generates a random 256 bit Data Encryption Key and Encrypts
// it with the underlying Crypt.
func (c *SecretBoxCrypt) GenerateDataKey() (*DataEncryptionKey, error) {
	genKey := c.GenerateRandomKey
	if genKey == nil {
		genKey = GenerateRandomKey
	}

	key := genKey()
	e, err := c.Encrypt(key[:])
	if err != nil {
		return nil, err
	}
	return &DataEncryptionKey{
		Plaintext: key[:],
		Envelope:  e,
	}, nil
}

// Simple helper around calling secretbox.Seal, handling nonce generation
// automatically.
func seal(m []byte, key [32]byte) ([]byte, error) {
	// You must use a different nonce for each message you encrypt with the
	// same key. Since the nonce here is 192 bits long, a random value
	// provides a sufficiently small probability of repeats.
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		panic(err)
	}
	return secretbox.Seal(nonce[:], m, &nonce, &key), nil
}

// Simple helper around calling secretbox.Open with something sealed with Seal.
func open(box []byte, key [32]byte) ([]byte, error) {
	// When you decrypt, you must use the same nonce and key you used to
	// encrypt the message. One way to achieve this is to store the nonce
	// alongside the encrypted message. Above, we stored the nonce in the first
	// 24 bytes of the encrypted text.
	var nonce [24]byte
	copy(nonce[:], box[:24])
	decrypted, ok := secretbox.Open(nil, box[24:], &nonce, &key)
	if !ok {
		return nil, fmt.Errorf("unable to decrypt message")
	}
	return decrypted, nil
}

// GenerateRandomKey generates a secure 256 bit random key.
func GenerateRandomKey() [32]byte {
	var key [32]byte
	if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
		panic(err)
	}
	return key
}
