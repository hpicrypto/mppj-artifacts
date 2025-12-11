package mppj

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, KEYSIZE)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	plaintext := []byte("This is a secret message")

	ciphertext, err := SymmetricEncrypt(key, plaintext)
	if err != nil {
		t.Errorf("encrypt() error = %v", err)
	}
	if len(ciphertext) == 0 {
		t.Fatalf("encrypt() returned empty ciphertext")
	}

	decrypted, err := SymmetricDecrypt(key, ciphertext)
	if err != nil {
		t.Errorf("decrypt() error = %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypt() = %s, want %s", decrypted, plaintext)
	}
}

func TestEncryptDecrypt_EmptyPlaintext(t *testing.T) {
	key := make([]byte, KEYSIZE)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	plaintext := []byte("")

	ciphertext, err := SymmetricEncrypt(key, plaintext)
	if err != nil {
		t.Errorf("encrypt() error = %v", err)
	}

	decrypted, err := SymmetricDecrypt(key, ciphertext)
	if err != nil {
		t.Errorf("decrypt() error = %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypt() = %s, want %s", decrypted, plaintext)
	}
}

func TestSymmetricEncryptDecrypt(t *testing.T) {
	key := make([]byte, KEYSIZE)
	_, err := io.ReadFull(rand.Reader, key)
	require.NoError(t, err, "Failed to generate key")

	invalidKey := make([]byte, KEYSIZE)
	_, err = io.ReadFull(rand.Reader, invalidKey)
	require.NoError(t, err, "Failed to generate invalid key")

	plaintext := []byte("This is a secret message")

	ciphertext, err := SymmetricEncrypt(key, plaintext)
	require.NoError(t, err, "SymmetricEncrypt() error")
	require.NotEmpty(t, ciphertext, "SymmetricEncrypt() returned empty ciphertext")

	decryptedText, err := SymmetricDecrypt(key, ciphertext)
	require.NoError(t, err, "SymmetricDecrypt() error")
	require.Equal(t, plaintext, decryptedText, "SymmetricDecrypt() did not return the original plaintext")
}

func TestDecrypt(t *testing.T) {
	msgBytes := make([]byte, 16)
	_, err := rand.Read(msgBytes)
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
	}
	msg, err := NewMessageFromBytes(msgBytes)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}
	sk, pk := PKEKeyGen()

	ciphertext := PKEEncrypt(pk, msg)
	decryptedMsg := PKEDecrypt(sk, ciphertext)
	if decryptedMsg == nil {
		t.Fatalf("Decrypt() returned nil message")
	}

	msgbytes, err := decryptedMsg.GetMessageBytes()
	if err != nil {
		t.Fatalf("Failed to get message bytes: %v", err)
	}

	if !bytes.Equal(msgbytes, msgBytes) {
		t.Errorf("Integration test failed: Decrypt() = %v, want %v", msgbytes, msgBytes)
	}
}

func TestReRand(t *testing.T) {
	msgBytes := make([]byte, 16)
	_, err := rand.Read(msgBytes)
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
	}
	msg, err := NewMessageFromBytes(msgBytes)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}
	_, pk := PKEKeyGen()

	ciphertext := PKEEncrypt(pk, msg)
	rerandCiphertext := ReRand(pk, ciphertext)
	if rerandCiphertext == nil {
		t.Fatalf("ReRand() returned nil ciphertext")
	}

	if rerandCiphertext.c0 == nil || rerandCiphertext.c1 == nil {
		t.Fatalf("ReRand() returned ciphertext with nil components")
	}
}

func TestIntegration(t *testing.T) {
	msgBytes := make([]byte, 16)
	_, err := rand.Read(msgBytes)
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
	}
	msg, err := NewMessageFromBytes(msgBytes)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}
	sk, pk := PKEKeyGen()

	// Encrypt
	ciphertext := PKEEncrypt(pk, msg)
	if ciphertext == nil {
		t.Fatalf("Encrypt() returned nil ciphertext")
	}

	// Re-randomize
	rerandCiphertext := ReRand(pk, ciphertext)
	if rerandCiphertext == nil {
		t.Fatalf("ReRand() returned nil ciphertext")
	}

	// Decrypt
	decryptedMsg := PKEDecrypt(sk, rerandCiphertext)
	if decryptedMsg == nil {
		t.Fatalf("Decrypt() returned nil message")
	}

	msgbytes, err := decryptedMsg.GetMessageBytes()
	if err != nil {
		t.Fatalf("Failed to get message bytes: %v", err)
	}

	if !bytes.Equal(msgbytes, msgBytes) {
		t.Errorf("Integration test failed: Decrypt() = %v, want %v", msgbytes, msgBytes)
	}

}

func TestPlaintext(t *testing.T) {
	msg_str := "helloworld"
	msg, err := NewMessageFromBytes([]byte(msg_str))
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}
	sk, pk := PKEKeyGen()

	// Encrypt
	ciphertext := PKEEncrypt(pk, msg)
	if ciphertext == nil {
		t.Fatalf("Encrypt() returned nil ciphertext")
	}

	// Re-randomize
	rerandCiphertext := ReRand(pk, ciphertext)
	if rerandCiphertext == nil {
		t.Fatalf("ReRand() returned nil ciphertext")
	}

	// Decrypt
	decryptedMsg := PKEDecrypt(sk, rerandCiphertext)
	if decryptedMsg == nil {
		t.Fatalf("Decrypt() returned nil message")
	}

	msgstr, err := decryptedMsg.GetMessageString()
	if err != nil {
		t.Fatalf("Failed to get message bytes: %v", err)
	}

	if msgstr != msg_str {
		t.Errorf("Integration test failed: Decrypt() = %v, want %v", msgstr, msg_str)
	}
}

func TestPlaintextUUID(t *testing.T) {
	for i := range PAYLOADSIZE + 1 { // byte length of curve modulus
		if i == 0 {
			continue
		}
		msg_str := uuid.New().String()[:i]
		msg, err := NewMessageFromBytes([]byte(msg_str))
		if err != nil {
			t.Fatalf("Failed to create message: %v", err)
		}
		sk, pk := PKEKeyGen()

		// Encrypt
		ciphertext := PKEEncrypt(pk, msg)
		if ciphertext == nil {
			t.Fatalf("Encrypt() returned nil ciphertext")
		}

		// Re-randomize
		rerandCiphertext := ReRand(pk, ciphertext)
		if rerandCiphertext == nil {
			t.Fatalf("ReRand() returned nil ciphertext")
		}

		// Decrypt
		decryptedMsg := PKEDecrypt(sk, rerandCiphertext)
		if decryptedMsg == nil {
			t.Fatalf("Decrypt() returned nil message")
		}

		msgstr, err := decryptedMsg.GetMessageString()
		if err != nil {
			t.Fatalf("Failed to get message bytes: %v", err)
		}

		if i <= PAYLOADSIZE {
			if msgstr != msg_str {
				t.Errorf("Integration test failed: Decrypt() = %v, want %v, iteration %v", msgstr, msg_str, i)
			}
		} else {
			if msgstr == msg_str {
				t.Errorf("Integration test failed: Decrypt() = %v, do not want %v, iteration %v", msgstr, msg_str, i)
			}
		}

	}

}

func TestSerializeDeserializeCiphertexts(t *testing.T) {
	msgBytes := make([]byte, 16)
	_, err := rand.Read(msgBytes)
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
	}
	msg, err := NewMessageFromBytes(msgBytes)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}
	_, pk := PKEKeyGen()

	// Encrypt
	ciphertext := PKEEncrypt(pk, msg)
	if ciphertext == nil {
		t.Fatalf("Encrypt() returned nil ciphertext")
	}

	// Serialize
	serializedCiphertext, err := ciphertext.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	if len(serializedCiphertext) == 0 {
		t.Fatalf("Serialize() returned empty slice")
	}

	// Deserialize
	deserializedCiphertext, err := DeserializeCiphertext(serializedCiphertext)
	if err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	// Check if the deserialized ciphertext matches the original ciphertext
	if !deserializedCiphertext.Equals(ciphertext) {
		t.Errorf("Deserialize() = %v, want %v", deserializedCiphertext, ciphertext)
	}
}

func TestEncryptVector(t *testing.T) {
	msgBytes := make([]byte, 100)
	_, err := rand.Read(msgBytes)
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
	}
	sk, pk := PKEKeyGen()

	// Encrypt
	ciphertexts, err := PKEEncryptVector(pk, msgBytes)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if len(ciphertexts) == 0 {
		t.Fatalf("Encrypt() returned empty ciphertext")
	}

	// Decrypt
	plaintext, err := PKEDecryptVector(sk, ciphertexts)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	// Check if the decrypted plaintext matches the original plaintext
	if string(plaintext) != string(msgBytes) {
		t.Errorf("Decrypt() = %v, want %v", hex.EncodeToString(plaintext), hex.EncodeToString(msgBytes))
	}
}

func TestSerializeDeserializeCiphertextsVector(t *testing.T) {
	msgBytes := make([]byte, 100)
	_, err := rand.Read(msgBytes)
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
	}

	sk, pk := PKEKeyGen()

	// Encrypt
	ciphertexts, err := PKEEncryptVector(pk, msgBytes)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if ciphertexts == nil {
		t.Fatalf("Encrypt() returned nil ciphertext")
	}

	// Serialize
	serializedCiphertext, err := SerializeCiphertexts(ciphertexts)
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	if len(serializedCiphertext) == 0 {
		t.Fatalf("Serialize() returned empty slice")
	}

	// Deserialize
	deserializedCiphertext, err := DeserializeCiphertexts(serializedCiphertext)
	if err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	plaintext, err := PKEDecryptVector(sk, deserializedCiphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	// Check if the deserialized ciphertext matches the original ciphertext
	for i := range ciphertexts {
		if !deserializedCiphertext[i].Equals(ciphertexts[i]) {
			t.Errorf("Deserialize() = %v, want %v", deserializedCiphertext[i], ciphertexts[i])
		}
	}

	// Check if the decrypted plaintext matches the original plaintext
	if string(plaintext) != string(msgBytes) {
		t.Errorf("Decrypt() = %v, want %v", plaintext, msgBytes)
	}

}

func TestPad32(t *testing.T) {
	msgBytes := make([]byte, 100)
	_, err := rand.Read(msgBytes)
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
	}

	padded := pad(msgBytes, 32)
	if len(padded) != 128 {
		t.Fatalf("Pad() returned slice of length %v, want 128", len(padded))
	}

	unpadded, err := unpad(padded)
	if err != nil {
		t.Fatalf("Unpad() error = %v", err)
	}
	if string(unpadded) != string(msgBytes) {
		t.Errorf("Unpad() = %v, want %v", unpadded, msgBytes)
	}
}

func TestPad31(t *testing.T) {
	msgBytes := make([]byte, 100)
	_, err := rand.Read(msgBytes)
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
	}

	padded := pad(msgBytes, 31)
	if len(padded) != 124 {
		t.Fatalf("Pad() returned slice of length %v, want 124", len(padded))
	}

	unpadded, err := unpad(padded)
	if err != nil {
		t.Fatalf("Unpad() error = %v", err)
	}
	if string(unpadded) != string(msgBytes) {
		t.Errorf("Unpad() = %v, want %v", unpadded, msgBytes)
	}
}

func TestPKEEncKeys(t *testing.T) {
	sk, pk := PKEKeyGen()

	for i := range 1000 {
		msgBytes := make([]byte, 16)
		_, err := rand.Read(msgBytes)
		if err != nil {
			t.Fatalf("Failed to generate random bytes: %v", err)
		}

		msg, err := NewMessageFromBytes(msgBytes)
		if err != nil {
			t.Fatalf("Failed to create message: %v", err)
		}

		// Encrypt
		ciphertext := PKEEncrypt(pk, msg)
		if ciphertext == nil {
			t.Fatalf("Encrypt() returned nil ciphertext")
		}

		ciphertext = ReRand(pk, ciphertext)

		// Decrypt
		decryptedMsg := PKEDecrypt(sk, ciphertext)
		if decryptedMsg == nil {
			t.Fatalf("Decrypt() returned nil message")
		}

		msgbytes, err := decryptedMsg.GetMessageBytes()
		if err != nil {
			t.Fatalf("Failed to get message bytes: %v", err)
		}

		// Check if the decrypted message matches the original message
		if hex.EncodeToString(msgbytes) != hex.EncodeToString(msgBytes) {
			t.Errorf("Iteration %d: Decrypt() = %v, want %v", i, hex.EncodeToString(msgbytes), hex.EncodeToString(msgBytes))
		}
	}
}
