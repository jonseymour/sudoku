package model

import (
	"encoding/hex"
	"fmt"
)

// A mapping from integers between 0 and 255 and the
// number of bits in the binary representation of that integer
//
// ref: OEIS A000120
//
var hammingWeight = [256]int{}

// The mask used to ensure that bits outside
// the range 0-80 are not set.
var mask = [11]byte{
	0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff,
	0x01,
}

// Maps each pattern of eight bits to the set of
// integers representing the set bit positions in those
// 8 bits.
var decoding = [256][]int{}

// A type that can efficiently perform logical
// operations on sets of 81 bits.
type BitSet struct {
	bits  [11]byte
	count int
}

// initialise the hamming weight and decodings for the integers 0-255
func init() {
	hammingWeight[0] = 0
	b := 0
	c := 1 << uint(b)
	for c < 256 {
		hammingWeight[c] = 1
		decoding[c] = []int{b}
		for i := 1; i < c; i++ {
			hammingWeight[c+i] = hammingWeight[i] + 1
			decoding[c+i] = append(decoding[i], b)
		}
		c <<= 1
		b++
	}
}

// The number of set bits.
func (s *BitSet) Size() int {
	return s.count
}

// True if the bit set contains the specified bit.
func (s *BitSet) Test(pos int) bool {
	return (s.bits[pos/8] & (1 << uint(pos%8))) != 0
}

// Set the specified bit of the receiver.
func (s *BitSet) Set(pos int) *BitSet {
	t := s.bits[pos/8]
	s.bits[pos/8] |= (1 << uint(pos%8))
	if t != s.bits[pos/8] {
		s.count++
	}
	return s
}

// Clear the specified but of the receiver
func (s *BitSet) Clear(pos int) *BitSet {
	t := s.bits[pos/8]
	s.bits[pos/8] &^= (1 << uint(pos%8))
	if t != s.bits[pos/8] {
		s.count--
	}
	return s
}

// Answer a new bit set which contains the
// the intersection of the receiver and the
// specified set.
func (s *BitSet) And(o *BitSet) *BitSet {
	r := &BitSet{}
	for i, b := range s.bits {
		r.bits[i] = b & o.bits[i]
		r.count += hammingWeight[r.bits[i]]
	}
	return r
}

// Answer a new bit set which includes only
// those bits of the receiver's set which are
// not in the specified set.
func (s *BitSet) AndNot(o *BitSet) *BitSet {
	r := &BitSet{}
	for i, b := range s.bits {
		r.bits[i] = (b &^ o.bits[i]) & mask[i]
		r.count += hammingWeight[r.bits[i]]
	}
	return r
}

// Answer the complement of the receiving set.
func (s *BitSet) Not() *BitSet {
	r := &BitSet{}
	for i, b := range s.bits {
		r.bits[i] = (0xff &^ b) & mask[i]
	}
	r.count = 81 - s.count
	return r
}

// Answer a new bit set which includes bits
// that are set in either of the receiver
// or the specified bit set.
func (s *BitSet) Or(o *BitSet) *BitSet {
	r := &BitSet{}
	for i, b := range s.bits {
		r.bits[i] = b | o.bits[i]
		r.count += hammingWeight[r.bits[i]]
	}
	return r
}

// Clone the bit set.
func (s *BitSet) Clone() *BitSet {
	r := &BitSet{
		count: s.count,
	}
	for i, c := range s.bits {
		r.bits[i] = c
	}
	return r
}

// Outputs the bit set as an array of the set bit positions
func (s *BitSet) AsInts() []int {
	r := make([]int, s.count, s.count)
	x := 0
	for i, b := range s.bits {
		offset := i << 3
		for _, c := range decoding[b] {
			r[x] = c + offset
			x++
		}
	}
	return r
}

// Converts a bit set to a hex string in most-significant-bit first order.
func (s *BitSet) AsHex() string {
	reversed := [11]byte{}
	for i, b := range s.bits {
		reversed[10-i] = b
	}
	return hex.EncodeToString(reversed[0:])
}

// Converts a hex representation into a bit set.
func FromHex(s string) *BitSet {
	r := &BitSet{}
	decoded, err := hex.DecodeString(s)
	l := len(decoded)
	if l > 11 || err != nil {
		panic(fmt.Errorf("illegal argument: %s, %v", s, err))
	}
	for i, b := range decoded {
		r.bits[l-i] = b
		r.count += hammingWeight[b]
	}
	return r
}

// Outputs the bit set as a string of the set bit positions
func (s *BitSet) String() string {
	return fmt.Sprintf("%v", s.AsInts())
}
