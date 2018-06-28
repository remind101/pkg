package envelope

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/kms"
)

type KMSCrypt struct {
	// The KMS Customer Master Key to use for encryption.
	KeyId string

	kms *kms.KMS
}

func NewKMSCrypt(c client.ConfigProvider) *KMSCrypt {
	k := kms.New(c)
	return &KMSCrypt{kms: k}
}

// Encrypt encrypts plaintext using the KMS CMK, and returns an Envelope.
func (c *KMSCrypt) Encrypt(plaintext []byte) (*Envelope, error) {
	resp, err := c.kms.Encrypt(&kms.EncryptInput{
		KeyId:     aws.String(c.KeyId),
		Plaintext: plaintext,
	})
	if err != nil {
		return nil, err
	}
	return &Envelope{Ciphertext: resp.CiphertextBlob}, nil
}

// GenerateDataKey generates a data encryption key by calling
// kms.GenerateDataKey. The data encryption key is encrypted with the KMS CMK
// and must be decrypted by calling kms.Decrypt.
func (c *KMSCrypt) GenerateDataKey() (*DataEncryptionKey, error) {
	resp, err := c.kms.GenerateDataKey(&kms.GenerateDataKeyInput{
		KeyId: aws.String(c.KeyId),
	})
	if err != nil {
		return nil, err
	}

	return &DataEncryptionKey{
		Plaintext: resp.Plaintext,
		Envelope: &Envelope{
			Ciphertext: resp.CiphertextBlob,
		},
	}, nil
}

// Decrypt decrypts an envelope that was encrypted with KMS.
func (c *KMSCrypt) Decrypt(envelope *Envelope) ([]byte, error) {
	resp, err := c.kms.Decrypt(&kms.DecryptInput{
		CiphertextBlob: envelope.Ciphertext,
	})
	if err != nil {
		return nil, err
	}

	return resp.Plaintext, nil
}
