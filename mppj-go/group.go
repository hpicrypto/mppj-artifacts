// This provides an abstraction to group operations, being essetially a wrapper around the elliptic curve operations.
// The notation is multiplicative
package mppj

import (
	"crypto/rand"
	"math/big"

	circl "github.com/cloudflare/circl/group"
)

var group = circl.P256

// Scalar represents a scalar value modulo the curve's order
type Scalar struct {
	s circl.Scalar
}

// Point represents a point on the elliptic curve.
type Point struct {
	p circl.Element
}

// SerializePoint serializes a Point into a byte slice.
func (p *Point) MarshalBinary() ([]byte, error) {

	return p.p.MarshalBinaryCompress()
}

// DeserializePoint deserializes a byte slice into a Point.
func (p *Point) UnmarshalBinary(data []byte) error {
	err := p.p.UnmarshalBinary(data)
	return err
}

// NewPoint creates a new Point with coordinates (x, y) modulo the curve's prime P.
func NewPoint() *Point {
	return &Point{p: group.NewElement()}
}

// Equals checks if two points a, b are equal by comparing their coordinates and ensuring they are on the curve.
func (a *Point) Equals(b *Point) bool {

	return a.p.IsEqual(b.p)

}

// Gen returns the generator of the elliptic curve.
func Gen() *Point {
	return &Point{p: group.Generator()}
}

// Identity returns the the identitiy element of the elliptic curve.
func Identity() *Point {
	return &Point{p: group.Identity()}
}

// Mul performs the group operation on two points a, b on the elliptic curve.
func Mul(a, b *Point) *Point {
	return &Point{p: a.p.Copy().Add(a.p, b.p)}
}

// MulBatched performs the group operation on a slice of points.
func MulBatched(pointArr []*Point) *Point {
	if len(pointArr) == 0 {
		return NewPoint()
	}
	result := pointArr[0]
	for _, point := range pointArr[1:] {
		result = Mul(result, point)
	}
	return result
}

// BaseExp exponentiates the generator by a scalar.
func BaseExp(s *Scalar) *Point {
	return &Point{p: group.NewElement().MulGen(s.s)}
}

// Inverts a point a on the elliptic curve.
func (a *Point) Invert() *Point {
	return &Point{p: a.p.Copy().Neg(a.p)}
}

// ScalarExp exponentiates a point a by a scalar b on the elliptic curve.
func (a *Point) ScalarExp(b *Scalar) *Point {
	return &Point{p: a.p.Copy().Mul(a.p, b.s)}
}

// InvertScalar returns the multiplicative inverse of a scalar.
func (s *Scalar) Invert() *Scalar {
	return &Scalar{s: s.s.Copy().Inv(s.s)}
}

// produces a unifromly random point on the curve
func RandomPoint() *Point {
	s := group.RandomScalar(rand.Reader)
	return &Point{p: group.NewElement().MulGen(s)} // faster than  group.RandomElement(rand.Reader)
}

// NewScalar creates a new scalar from value.
func NewScalarEmpty() *Scalar {
	return &Scalar{s: group.NewScalar()}
}

// NewScalar creates a new scalar from value.
func NewScalar(value *big.Int) *Scalar {
	return &Scalar{s: group.NewScalar().SetBigInt(value)}
}

// Adds 2 scalars a, b.
func (a *Scalar) Add(b *Scalar) *Scalar {
	return &Scalar{s: a.s.Copy().Add(a.s, b.s)}
}

// Multiplies 2 scalars a, b.
func (a *Scalar) Mul(b *Scalar) *Scalar {
	return &Scalar{s: a.s.Copy().Mul(a.s, b.s)}
}

func (a *Scalar) Equals(b *Scalar) bool {
	return a.s.IsEqual(b.s)
}

func (a *Scalar) Neg() *Scalar {
	return &Scalar{s: a.s.Copy().Neg(a.s)}
}

func (a *Scalar) Copy() *Scalar {
	return &Scalar{s: a.s.Copy()}
}

// RandomScalar creates a new random scalar.
func RandomScalar() *Scalar {

	return &Scalar{
		s: group.RandomScalar(rand.Reader),
	}
}

// HashToPoint hashes a byte slice to a scalar. See hash to field/group RFC
func HashToPoint(msg, sid []byte) *Point {
	prefix := []byte("hash_to_element")
	dst := append(prefix, sid...)
	return &Point{p: group.HashToElement(msg, dst)}
}
