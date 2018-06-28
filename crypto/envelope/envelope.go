// Package envelope provides a library for utilizing envelope encryption to
// encrypt small messages.
//
// Envelope encryption is the process of setting up an encryption chain, where
// keys are encrypted with other keys higher up the chain. In many cases, the
// top level key is one stored in an HSM, like a KMS CMK.
//
// The goals that envelope encryption attempt to solve are:
//
//	1. Easy encryption with sane key management.
//	2. Mitigate risks associated with leaking encryption keys.
//	3. Support fast key revokation and rotation.
//
// This package is designed in such a way that it operates as a black box, and
// consumers never see raw key material, making leaks of encryption keys
// unlikely.
package envelope

// Envelope represents some encrypted data, included the encryption key chain
// needed to decrypt the value.
type Envelope struct {
	// Ciphertext is encrypted version of the plaintext message.
	Ciphertext []byte `json:"ciphertext"`
	// Encryption key represents the key that was used encrypt Ciphertext.
	// This can be another Envelope instance or a pointer to a key stored in
	// another location. In some cases it's ok for this to be nil, when
	// Ciphertext can be decrypted without knowing the key that was used to
	// encrypt (e.g. with KMS).
	EncryptionKey interface{} `json:"encryption_key"`
}

// DataEncryptionKey represents a generated data encryption key, including the
// plaintext.
type DataEncryptionKey struct {
	*Envelope
	Plaintext []byte
}

// Crypt represents something that can encrypt a plaintext message with a Data
// Encryption Key (DEK), and returns an envelope with the encrypted DEK used to
// encrypt the data, along with the ciphertext of the encrypted data itself.
//
// In most cases, this interface is used recursively to setup a key encryption
// chain, whereby encrypting an envelope recursively decrypts the keys needed to
// perform the decryption.
type Crypt interface {
	Encrypt([]byte) (*Envelope, error)
	Decrypt(*Envelope) ([]byte, error)
	GenerateDataKey() (*DataEncryptionKey, error)
}
