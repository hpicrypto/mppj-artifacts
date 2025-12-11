package mppj

import (
	"crypto/hkdf"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
)

type OPRFKey Scalar

// OPRFKeyGen generates a new random key for the DH-OPRF.
func OPRFKeyGen() *OPRFKey {
	k := RandomScalar()
	return (*OPRFKey)(k)
}

// OPRFBlind computes the encryption of m using the public key bpk.
func OPRFBlind(bpk *PublicKey, msg, sid []byte) *Ciphertext {
	hmsg := HashToMessage(msg, sid)
	return PKEEncrypt(bpk, hmsg)
}

// OPRFUnblind computes the decryption of the ciphertext using the secret key bsk.
func OPRFUnblind(bsk *SecretKey, ciphertext *Ciphertext) *Message {
	return PKEDecrypt(bsk, ciphertext)
}

// OPRFEval computes the encryption of m^k. Computes ReRand internally.
func OPRFEval(key *OPRFKey, bpk *PublicKey, ciphertext *Ciphertext) *Ciphertext {
	c0 := ciphertext.c0.ScalarExp((*Scalar)(key))
	c1 := ciphertext.c1.ScalarExp((*Scalar)(key))

	return ReRand(bpk, &Ciphertext{c0: c0, c1: c1})
}

// NewSessionID generates a new session ID based on session participants and randomness.
func NewSessionID(numsources int, helper, receiver string, datasources []SourceID) []byte {

	sidprime := uuid.New().String()

	info := fmt.Sprintf("%d", numsources) + "|" + helper + "|" + receiver
	for _, ds := range datasources {
		info += "|" + string(ds)
	}

	sid, err := hkdf.Key(sha256.New, []byte(sidprime), nil, info, sha256.New().Size())
	if err != nil {
		panic(err)
	}

	return sid
}
