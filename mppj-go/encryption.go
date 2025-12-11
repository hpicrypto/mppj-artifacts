package mppj

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"math/big"

	"crypto/elliptic"
	"crypto/hkdf"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/crypto/blake2b"
)

var curve = elliptic.P256()

const KEYSIZE = 16
const PAYLOADSIZE = 30

var ZeroNonce = make([]byte, aes.BlockSize)

// *********************** Types ************************

type PublicKey Point

func (pk *PublicKey) String() string {
	pb, _ := pk.p.MarshalBinary()
	return hex.EncodeToString(pb)
}

type SecretKey Scalar

type SymmetricCiphertext []byte

type SecretKeyTuple struct {
	bsk *SecretKey
	esk *SecretKey
}

type PublicKeyTuple struct {
	bpk *PublicKey
	epk *PublicKey
}

func (pkt *PublicKeyTuple) String() string {
	return fmt.Sprintf("bpk: %v,\nepk: %v", pkt.bpk, pkt.epk)
}

// Message represents a message point on the elliptic curve.
type Message struct {
	m Point
}

// Ciphertext represents an ElGamal ciphertext. c0 = g^r, c1 = m * pk^r
type Ciphertext struct {
	c0 *Point // g^r
	c1 *Point // m * pk^r
}

// pad pads the input byte slice to the next multiple of blockSize.
func pad(data []byte, blockSize int) []byte {

	padding := blockSize - len(data)%blockSize
	if padding == 0 {
		padding = blockSize
	}
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)

}

// unpad removes the padding from the input byte slice.
func unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("invalid padding size, empty data")
	}
	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return nil, errors.New("invalid padding size")
	}
	for _, v := range data[len(data)-padding:] {
		if int(v) != padding {
			return nil, errors.New("invalid padding")
		}
	}
	return data[:len(data)-padding], nil
}

// *********************** PKE ************************

// PKEEncrypt encrypts a message msg using the public key pk.
func PKEEncrypt(pk *PublicKey, msg *Message) *Ciphertext {
	r := RandomScalar()

	c0 := BaseExp(r)
	c1 := Mul(&msg.m, (*Point)(pk).ScalarExp(r))

	return &Ciphertext{
		c0: c0,
		c1: c1,
	}

}

// PKEEncryptVector encrypts a byte slice PAYLOADSIZE bytes at a time using the public key pk. ( due to the 256-bit curve)
func PKEEncryptVector(pk *PublicKey, msg []byte) ([]*Ciphertext, error) {

	ciphertexts := make([]*Ciphertext, len(pad(msg, PAYLOADSIZE))/PAYLOADSIZE)
	msg_padded := pad(msg, PAYLOADSIZE)

	for i := 0; i < len(msg_padded); i += PAYLOADSIZE {
		end := i + PAYLOADSIZE
		chunk := make([]byte, PAYLOADSIZE)
		copy(chunk, msg_padded[i:end])
		idx := i / PAYLOADSIZE

		msg, err := NewMessageFromBytes(chunk)
		if err != nil {
			return nil, err
		}
		ciphertexts[idx] = PKEEncrypt(pk, msg)
	}

	return ciphertexts, nil

}

// PKEDecrypt decrypts a ciphertext using the secret key sk.
func PKEDecrypt(sk *SecretKey, ciphertext *Ciphertext) *Message {
	// Calculate s = (g ^ r) ^ -sk
	s := ciphertext.c0.ScalarExp((*Scalar)(sk))

	m := Mul(ciphertext.c1, s)

	return &Message{m: *m}
}

// PKEDecryptVector decrypts a slice of ciphertexts using the secret key sk.
func PKEDecryptVector(sk *SecretKey, ciphertexts []*Ciphertext) ([]byte, error) {
	msgBytes := make([]byte, 0)
	msgByteshelper := make([][]byte, len(ciphertexts))
	var wg sync.WaitGroup
	errCh := make(chan error, len(ciphertexts))

	for i, ct := range ciphertexts {
		wg.Add(1)
		go func(i int, ct *Ciphertext) {
			defer wg.Done()
			msg, err := PKEDecrypt(sk, ct).GetMessageBytes()
			if err != nil {
				errCh <- err
				return
			}
			msgByteshelper[i] = msg
		}(i, ct)
	}
	wg.Wait()
	close(errCh)

	// Check for errors
	if len(errCh) > 0 {
		return nil, <-errCh
	}

	for _, msg := range msgByteshelper {
		msgBytes = append(msgBytes, msg...)
	}

	msgBytes, err := unpad(msgBytes) // hom. PKE cannot be CCA secure anyway, so padding oracles are not a concern
	if err != nil {
		return nil, err
	}
	return msgBytes, nil
}

// ReRand re-randomizes a ciphertext using pk.
func ReRand(pk *PublicKey, ciphertext *Ciphertext) *Ciphertext {
	r := RandomScalar()

	c0 := Mul(ciphertext.c0, BaseExp(r))
	c1 := Mul(ciphertext.c1, (*Point)(pk).ScalarExp(r))

	return &Ciphertext{
		c0: c0,
		c1: c1,
	}

}

// ReRandVector re-randomizes a slice of ciphertexts using pk.
func ReRandVector(pk *PublicKey, ciphertexts []*Ciphertext) []*Ciphertext {
	ciphertextsout := make([]*Ciphertext, len(ciphertexts))
	var wg sync.WaitGroup
	for i, ct := range ciphertexts {
		wg.Add(1)
		go func(i int, ct *Ciphertext) {
			defer wg.Done()
			ciphertextsout[i] = ReRand(pk, ct)
		}(i, ct)
	}
	wg.Wait()
	return ciphertextsout
}

// PKEKeyGen generates a new public/private key pair. (scalar, point)
func PKEKeyGen() (*SecretKey, *PublicKey) {
	sk := RandomScalar()

	pk := BaseExp(sk)
	return (*SecretKey)(sk.Neg()), (*PublicKey)(pk) // Negate the scalar for efficiency
}

// Serialize serializes a Ciphertext into a byte slice.
func (ct *Ciphertext) Serialize() ([]byte, error) {
	c0Bytes, err := ct.c0.MarshalBinary()
	if err != nil {
		return nil, err
	}
	c1Bytes, err := ct.c1.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return append(c0Bytes, c1Bytes...), nil
}

func SerializeCiphertexts(cts []*Ciphertext) ([]byte, error) {
	serialized := make([]byte, 0)
	for _, ct := range cts {
		serializedct, err := ct.Serialize()
		if err != nil {
			return nil, err
		}

		serialized = append(serialized, serializedct...)
	}
	return serialized, nil
}

// DeserializeCiphertexts deserializes a byte slice into a slice of Ciphertexts.
func DeserializeCiphertexts(data []byte) ([]*Ciphertext, error) {

	byteLen := int(group.Params().CompressedElementLength)
	ciphertextlen := 2 * byteLen
	if len(data)%(ciphertextlen) != 0 {
		return nil, errors.New("invalid byte slice length for deserialization of array")
	}
	ciphertexts := make([]*Ciphertext, 0)
	for i := 0; i < len(data); i += ciphertextlen {
		ciphertext, err := DeserializeCiphertext(data[i : i+ciphertextlen])
		if err != nil {
			return nil, err
		}

		ciphertexts = append(ciphertexts, ciphertext)
	}
	return ciphertexts, nil
}

// DeserializeCiphertext deserializes a byte slice into a Ciphertext.
func DeserializeCiphertext(data []byte) (*Ciphertext, error) {
	byteLen := int(group.Params().CompressedElementLength)
	pointLen := 2 * byteLen

	if len(data) != pointLen {
		return nil, errors.New("invalid byte slice length for deserialization")
	}

	c0 := NewPoint()
	c1 := NewPoint()

	err := c0.UnmarshalBinary(data[:byteLen])
	if err != nil {
		return nil, err
	}

	err = c1.UnmarshalBinary(data[byteLen:])
	if err != nil {
		return nil, err
	}

	return &Ciphertext{
		c0: c0,
		c1: c1,
	}, nil
}

func (msg *Message) String() string {
	msgstr, err := msg.GetMessageString()
	if err != nil {
		return "Invalid message"
	}
	return fmt.Sprintf("Message(%s)", msgstr)
}

func (msg *Message) Equals(other *Message) bool {
	if msg == nil || other == nil {
		return msg == other
	}
	return msg.m.Equals(&other.m)
}

// Equals checks if two Ciphertexts are equal.
func (ct *Ciphertext) Equals(other *Ciphertext) bool {
	if ct == nil || other == nil {
		return ct == other
	}
	return ct.c0.Equals(other.c0) && ct.c1.Equals(other.c1)
}

func NewMessageFromBytes(msgBytesin []byte) (*Message, error) {
	params := curve.Params()

	if len(msgBytesin) == 0 {
		return nil, errors.New("Empty message unsupported")
	}

	msgBytes := make([]byte, len(msgBytesin))

	copy(msgBytes, msgBytesin)

	// Prefix msgInt with one LSB byte
	msgBytes = append(msgBytes, 0x02)            // Prefix LSB byte
	msgBytes = append([]byte{0x04}, msgBytes...) // Postfix MSB byte
	msgInt := new(big.Int).SetBytes(msgBytes)

	i := 1

	y := new(big.Int)
	for {
		// adapted from elliptic.polynomial
		x3 := new(big.Int).Mul(msgInt, msgInt)
		x3.Mul(x3, msgInt)

		threeX := new(big.Int).Lsh(msgInt, 1)
		threeX.Add(threeX, msgInt)

		x3.Sub(x3, threeX)
		x3.Add(x3, params.B)
		x3.Mod(x3, params.P)

		// Try to calculate the square root mod p (y = sqrt(y^2) mod p)
		y = new(big.Int).ModSqrt(x3, params.P)
		if y != nil {
			break
		}

		if i == 255 { // there is only one byte of space for the counter
			return nil, errors.New("Failed to find a valid message point")

		}
		i++

		// Update msgInt for the next iteration if not valid
		msgInt.Add(msgInt, big.NewInt(1))
	}

	pointBytes := elliptic.Marshal(curve, msgInt, y)
	result := NewPoint()
	err := result.UnmarshalBinary(pointBytes)
	if err != nil {
		return nil, err // when using a compatible curve, this should never happen
	}

	return &Message{m: *result}, nil
}

// GetMessageBytes returns the message as a byte slice.
func (msg *Message) GetMessageBytes() ([]byte, error) {
	serialized, err := msg.m.MarshalBinary()
	if err != nil {
		return nil, err
	}
	x, _ := elliptic.UnmarshalCompressed(curve, serialized)

	if x == nil {
		return nil, fmt.Errorf("failed to unmarshal message point")
	}

	msgBytes := x.Bytes()

	msgBytes = msgBytes[1 : len(msgBytes)-1] // Remove the prefix and LSB

	return msgBytes, nil
}

func (msg *Message) GetMessageString() (string, error) {
	bytes, err := msg.GetMessageBytes()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (msg *Message) GetMessageStringHex() (string, error) {
	bytes, err := msg.GetMessageBytes()
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// RandomMsg creates a new random message point.
func RandomMsg() (*Message, error) {

	randomPoint := RandomPoint()

	return &Message{m: *randomPoint}, nil
}

// HashToPoint hashes a byte slice to a point on the curve. Uses the secure hash-to-group approach from the underlying group
func HashToMessage(msg, sid []byte) *Message {
	return &Message{m: *HashToPoint(msg, sid)}
}

// *********************** Symmetric ************************

// RandomKeyFromPoint generates a random 16-byte key from a random point on the curve
func RandomKeyFromPoint(sid []byte) (*Point, []byte) {
	rp := RandomPoint()

	key, err := KeyFromPoint(rp, sid)
	if err != nil {
		panic(err) // Random points are assumed to be "correct"
	}

	return rp, key
}

func KeyFromPoint(rp *Point, sid []byte) ([]byte, error) {
	info := "ephemeral associated data val key"

	serialized, err := rp.MarshalBinary()
	if err != nil {
		return nil, err
	}

	key, err := hkdf.Key(sha256.New, serialized, sid, info, KEYSIZE)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// ctr encrypts/decrypts the plaintext using AES-CTR with the given key and nonce.
func ctr(key, plaintext []byte) (SymmetricCiphertext, error) {

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, len(plaintext))
	cipher.NewCTR(block, ZeroNonce).XORKeyStream(ciphertext, plaintext) // nonce = 16 * 0x00 since each key is used only once

	return ciphertext, nil

}

// Encrypt the plaintext bytes with the symmetric key using AES-CTR
func SymmetricEncrypt(key []byte, plaintext []byte) (SymmetricCiphertext, error) {
	return ctr(key, plaintext)
}

// Decrypt the ciphertext bytes with the symmetric key using AES-CTR
func SymmetricDecrypt(key []byte, ciphertext SymmetricCiphertext) ([]byte, error) {
	return ctr(key, ciphertext)
}

// Generates keys *deterministically* from a seed
func GetTestKeys(seed []byte) (SecretKeyTuple, PublicKeyTuple) {

	xof, err := blake2b.NewXOF(blake2b.OutputLengthUnknown, seed)
	if err != nil {
		panic(err)
	}

	esk := &Scalar{s: group.RandomScalar(xof)}
	bsk := &Scalar{s: group.RandomScalar(xof)}
	rsk := SecretKeyTuple{esk: (*SecretKey)(esk.Neg()), bsk: (*SecretKey)(bsk.Neg())}
	rpk := PublicKeyTuple{epk: (*PublicKey)(BaseExp(esk)), bpk: (*PublicKey)(BaseExp(bsk))}

	return rsk, rpk
}
