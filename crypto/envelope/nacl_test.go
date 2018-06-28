package envelope

import "testing"

func TestSecretBoxCrypt(t *testing.T) {
	crypt := &SecretBoxCrypt{
		encryptionKeys: newStaticCrypt(),
	}

	envelope, err := crypt.Encrypt([]byte("hello world"))
	if err != nil {
		t.Fatal(err)
	}

	plaintext, err := crypt.Decrypt(envelope)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := string(plaintext), "hello world"; got != want {
		t.Fatalf("expected to be able to decrypt the encrypted envelope")
	}
}

// staticCrypt implements the Crypt interface and generates secure 256 bit
// random keys, which are encrypted with this static key.
type staticCrypt [32]byte

func newStaticCrypt() staticCrypt {
	return GenerateRandomKey()
}

func (c staticCrypt) Encrypt(plaintext []byte) (*Envelope, error) {
	ciphertext, err := seal(plaintext, c)
	if err != nil {
		return nil, err
	}
	return &Envelope{
		Ciphertext: ciphertext,
	}, nil
}

func (c staticCrypt) Decrypt(envelope *Envelope) ([]byte, error) {
	plaintext, err := open(envelope.Ciphertext, c)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func (c staticCrypt) GenerateDataKey() (*DataEncryptionKey, error) {
	key := GenerateRandomKey()
	e, err := c.Encrypt(key[:])
	if err != nil {
		return nil, err
	}

	return &DataEncryptionKey{
		Envelope:  e,
		Plaintext: key[:],
	}, nil
}
