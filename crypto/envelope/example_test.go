package envelope_test

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/remind101/pkg/crypto/envelope"
)

// This example demonstrates how one might use envelope encryption to encrypt large
// infrequently accessed secrets in a database. For this, we use KMS to generate
// a new DEK for every value that we encrypt.
func Example_LargeInfrequentlyUsedSymmetricSecrets_WithKMS() {
	dataEncryptionKeys := envelope.NewKMSCrypt(session.New())
	dataEncryptionKeys.KeyId = "some KMS key ARN"

	// Use NaCl secretbox to encrypt values with KMS generate data key.
	enc := envelope.NewSecretBoxCrypt(dataEncryptionKeys)

	envelope, _ := enc.Encrypt([]byte("some super secret api key"))
	raw, _ := json.Marshal(envelope)

	// Now store the raw envelope in the database.

	// When you want to encrypt it at a later date:

	json.Unmarshal(raw, &envelope)

	plaintext, _ := enc.Decrypt(envelope)
	fmt.Println(plaintext)
}

// This example demonstrates how one might use envelope encryption to encrypt
// small frequently accessed secrets in a database. For this, we need to limit
// the number of KMS API calls we make. To achieve this, we only use KMS to
// generate a limited number of Key Encryption Keys (KEK's) and then rely on
// NaCl for Data Encryption Key generation and encryption.
//
// This setups up a key encryption chain so that to decrypt a value, one must:
//
//	1. Decrypt the Key Encryption Key (KEK) using kms.Decrypt.
//	2. Decrypt the Data Encryption Key (DEK) with NaCl using the decrypted
//	   KEK.
//	3. Finally, decrypt the value using the decrypted DEK.
//
// This example trivializes KEK management. In a real world scenario, you would
// want to rate limit the number of KEK's that are generated, provide a means
// for lifecycle management, cache calls to decrypt KEK's, and other
// considerations. You can provide your own envelope.Crypt implmentation to
// achieve these goals.
func Example_SmallFrequentlyUsedSymmetricSecrets_WithKMS() {
	// KMS will be used to generate data keys, which will only be used to
	// encrypt Data Encryption Keys (DEK's).
	keyEncryptionKeys := envelope.NewKMSCrypt(session.New())
	keyEncryptionKeys.KeyId = "some KMS key ARN"

	// generate random 256 data encryption keys, which will be protected by
	// the KMS key encryption keys above.
	dataEncryptionKeys := envelope.NewSecretBoxCrypt(keyEncryptionKeys)

	// Use NaCl secretbox to encrypt values with the randomly generated data
	// keys.
	enc := envelope.NewSecretBoxCrypt(dataEncryptionKeys)

	envelope, _ := enc.Encrypt([]byte("some super secret api key"))
	fmt.Println(envelope)
}
