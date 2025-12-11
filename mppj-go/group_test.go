package mppj

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"testing"

	"go.dedis.ch/kyber/v4/suites"
	"go.dedis.ch/kyber/v4/util/random"
)

func TestKyber(t *testing.T) {

	//fmt.Println(p.EmbedLen())
}

func BenchmarkEmbed(b *testing.B) {

	msg := []byte("DEADBEEFCAFEFACEBAD")
	suite := suites.MustFind("P256")

	rnd := random.New()
	b.Run("Kyber", func(b *testing.B) {
		p := suite.Point()
		p.Embed(msg, rnd)
	})

	b.Run("Ours", func(b *testing.B) {
		msg, err := NewMessageFromBytes(msg)
		if err != nil {
			panic(err)
		}
		_ = msg

	})
}

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name     string
		msgBytes []byte
		wantErr  bool
	}{
		{
			name:     "Valid message",
			msgBytes: []byte{1},
			wantErr:  false,
		},
		{
			name:     "Zero message",
			msgBytes: []byte{0},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := NewMessageFromBytes(tt.msgBytes)
			if err != nil && !tt.wantErr {
				t.Errorf("NewMessage() error = %v, want nil", err)
			}
			if msg == nil && !tt.wantErr {
				t.Errorf("NewMessage() returned nil, expected valid message")
			}
			if msg == nil && !tt.wantErr {
				t.Errorf("NewMessage() returned nil, expected valid message")
			}
			if msg != nil && tt.wantErr {
				t.Errorf("NewMessage() returned valid message, expected error")
			}
		})
	}
}

func TestGetMessage(t *testing.T) {
	msgBytes := make([]byte, 16)
	_, err := rand.Read(msgBytes)
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
	}
	msg, err := NewMessageFromBytes(msgBytes)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}
	got, err := msg.GetMessageStringHex()
	if err != nil {
		t.Fatalf("Failed to get message string: %v", err)
	}
	if hex.EncodeToString(msgBytes) != got {
		t.Errorf("GetMessage() = %v, want %v", got, hex.EncodeToString(msgBytes))
	}
}

func TestRandomMsg(t *testing.T) {
	msg, err := RandomMsg()
	if err != nil {
		t.Errorf("RandomMsg() error = %v", err)
	}
	if msg == nil {
		t.Errorf("RandomMsg() returned nil, expected valid message")
	}
}

func TestScalarMul(t *testing.T) {
	tests := []struct {
		name string
		a    *Point
		b    *Scalar
		want *Point
	}{
		{
			name: "Scalar multiplication with base point",
			a:    Gen(),
			b:    NewScalar(big.NewInt(2)),
			want: Gen().ScalarExp(NewScalar(big.NewInt(2))),
		},
		{
			name: "Scalar multiplication with one",
			a:    Gen(),
			b:    NewScalar(big.NewInt(1)),
			want: Gen(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.ScalarExp(tt.b)
			if !got.Equals(tt.want) {
				t.Errorf("ScalarMul() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRandomPoint(t *testing.T) {
	point := RandomPoint()

	if point == nil {
		t.Fatalf("GetRandomPoint() returned nil point")
	}

}

func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		a, b *Point
		want *Point
	}{
		{
			name: "Add base point to itself",
			a:    Gen(),
			b:    Gen(),
			want: Mul(Gen(), Gen()),
		},
		{
			name: "Add base point to zero point",
			a:    Gen(),
			b:    Identity(),
			want: Gen(),
		},
		{
			name: "Add zero point to base point",
			a:    Identity(),
			b:    Gen(),
			want: Gen(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Mul(tt.a, tt.b); !got.Equals(tt.want) {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInvert(t *testing.T) {
	tests := []struct {
		name  string
		point *Point
		want  *Point
	}{
		{
			name:  "Invert base point",
			point: Gen(),
			want:  BaseExp(NewScalar(big.NewInt(1)).Neg()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.point.Invert(); !got.Equals(tt.want) {
				t.Errorf("Invert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInvert2(t *testing.T) {
	point := RandomPoint()

	inverted := point.Invert()
	if inverted == nil {
		t.Fatalf("Invert() returned nil point")
	}

	if !Mul(Mul(point, inverted), Gen()).Equals(Gen()) {
		t.Errorf("Invert() failed to invert point")
	}
}

func TestInvert3(t *testing.T) {
	point := RandomPoint()

	inverted := point.Invert()
	if inverted == nil {
		t.Fatalf("Invert() returned nil point")
	}

	if !Mul(Mul(point, Gen()), inverted).Equals(Gen()) {
		t.Errorf("Invert() failed to invert point")
	}
}

func TestEncrypt(t *testing.T) {
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
	if ciphertext == nil {
		t.Fatalf("Encrypt() returned nil ciphertext")
	}

	if ciphertext.c0 == nil || ciphertext.c1 == nil {
		t.Fatalf("Encrypt() returned ciphertext with nil components")
	}
}

func TestEqual(t *testing.T) {
	a := Gen()
	b := Gen()
	if !a.Equals(b) {
		t.Errorf("Equals() = false, want true")
	}
}

func TestSerializeDeserializePoint(t *testing.T) {
	// Generate a random point
	point := RandomPoint()

	// Serialize the point
	serializedPoint, err := point.MarshalBinary()
	if err != nil {
		t.Fatalf("SerializePoint() error = %v", err)
	}

	// Deserialize the point
	deserializedPoint := NewPoint()

	err = deserializedPoint.UnmarshalBinary(serializedPoint)
	if err != nil {
		t.Fatalf("DeserializePoint() error = %v", err)
	}

	// Check if the deserialized point matches the original point
	if !point.Equals(deserializedPoint) {
		t.Errorf("DeserializePoint() = %v, want %v", deserializedPoint, point)
	}
}

func TestSerializeDeserializeCiphertext(t *testing.T) {

	point1 := RandomPoint()
	point2 := RandomPoint()

	ciphertext := &Ciphertext{c0: point1, c1: point2}

	// Serialize the ciphertext
	serializedCiphertext, err := ciphertext.Serialize()
	if err != nil {
		t.Fatalf("SerializeCiphertext() error = %v", err)
	}

	// Deserialize the ciphertext
	deserializedCiphertext, err := DeserializeCiphertext(serializedCiphertext)
	if err != nil {
		t.Fatalf("DeserializeCiphertext() error = %v", err)
	}

	// Check if the deserialized ciphertext matches the original ciphertext
	if !deserializedCiphertext.Equals(ciphertext) {
		t.Errorf("DeserializeCiphertext() = %v, want %v", deserializedCiphertext, ciphertext)
	}
}

func TestSerializeDeserializeCiphertext2(t *testing.T) {

	c0 := RandomPoint()

	c1 := RandomPoint()

	ciphertext := &Ciphertext{c0: c0, c1: c1}

	// Serialize the ciphertext
	serializedCiphertext, err := ciphertext.Serialize()
	if err != nil {
		t.Fatalf("SerializeCiphertext() error = %v", err)
	}

	// Deserialize the ciphertext
	deserializedCiphertext, err := DeserializeCiphertext(serializedCiphertext)
	if err != nil {
		t.Fatalf("DeserializeCiphertext() error = %v", err)
	}

	// Check if the deserialized ciphertext matches the original ciphertext
	if !deserializedCiphertext.Equals(ciphertext) {
		t.Errorf("DeserializeCiphertext() = %v, want %v", deserializedCiphertext, ciphertext)
	}
}

func TestScalarAddition(t *testing.T) {
	s1 := RandomScalar()

	s2 := RandomScalar()

	s3 := s1.Add(s2)

	if !s3.Equals(s1.Add(s2)) {
		t.Errorf("Addition failed: %s + %s != %s", s1, s2, s3)
	}
}

func TestScalarAdditionExp(t *testing.T) {

	for i := 0; i < 100; i++ {
		s1 := RandomScalar()
		s2 := RandomScalar()

		s3 := s1.Add(s2)

		if !BaseExp(s3).Equals(Mul(BaseExp(s2), BaseExp(s1))) {
			t.Errorf("Addition %d failed: %s + %s != %s", i, BaseExp(s3), BaseExp(s2), BaseExp(s1))
		}
	}
}

func TestSecretExponentiationGen(t *testing.T) {

	num_shares := 10
	nonceSum := NewScalar(big.NewInt(0))
	blind_shares := make([]*Point, num_shares)

	nonces := make([]*Scalar, num_shares)
	for i := range num_shares {
		s := RandomScalar()

		nonces[i] = s.Copy()
		blind_shares[i] = BaseExp(nonces[i].Copy())
		nonceSum = nonceSum.Add(nonces[i].Copy())
	}

	blind := BaseExp(nonceSum)
	decKeyTemp := MulBatched(blind_shares)

	if !blind.Equals(decKeyTemp) {
		t.Errorf("Secrets do not match: %s != %s", blind, decKeyTemp)
	}
}

func TestSecretExponentiation(t *testing.T) {

	num_shares := 10
	blind_base := RandomPoint()
	nonceSum := NewScalar(big.NewInt(0))
	blind_shares := make([]*Point, num_shares)

	nonces := make([]*Scalar, num_shares)
	for i := range num_shares {
		s := RandomScalar()

		nonces[i] = s
		blind_shares[i] = blind_base.ScalarExp(s)
		nonceSum = nonceSum.Add(s)
	}

	blind := blind_base.ScalarExp(nonceSum)
	decKeyTemp := MulBatched(blind_shares)

	if !blind.Equals(decKeyTemp) {
		t.Errorf("Secrets do not match: %s != %s", blind, decKeyTemp)
	}
}

func TestPlantextSecretSharing(t *testing.T) {

	rp := RandomPoint()
	num_shares := 10
	blind_base := RandomPoint()
	nonceSum := NewScalar(big.NewInt(0))

	nonces := make([]*Scalar, num_shares)
	for i := range num_shares {
		s := RandomScalar()

		nonces[i] = s
		nonceSum = nonceSum.Add(s)
	}

	blind := blind_base.ScalarExp(nonceSum)
	blinded_point := Mul(rp, blind)

	decKeyTemp := Gen() // no identity in this lib for now :(
	for _, nonce := range nonces {
		decKeyTemp = Mul(decKeyTemp, blind_base.ScalarExp(nonce))
	}

	decKeyTemp = Mul(decKeyTemp, Gen().Invert()).Invert()
	recovered_rp := Mul(blinded_point, decKeyTemp)

	if !rp.Equals(recovered_rp) {
		t.Errorf("Secrets do not match: %s != %s", rp, recovered_rp)
	}
}
