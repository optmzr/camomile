package node

import (
	"bytes"
	"crypto/rand"
	"errors"
	"math/bits"
	"testing"
)

func TestEqual(t *testing.T) {
	id1 := NewID()
	id2 := NewID()

	if id1.Equal(id2) {
		t.Error("two instances of ids must not equal")
	}

	if !id1.Equal(id1) {
		t.Error("same instance of id must equal itself")
	}

	rng = func(_ []byte) (int, error) {
		return 0, errors.New("error")
	}

	defer func() {
		rng = rand.Read
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()

	NewID() // Should panic.
}

func TestIDFromStringValid(t *testing.T) {
	id, err := IDFromString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	if err != nil {
		t.Errorf("unexpected error: %w", err)
	}

	for _, b := range id {
		if byte(0xff) != b {
			t.Errorf("unexpected byte: %b", b)
		}
	}
}

func TestIDFromStringInvalidLength(t *testing.T) {
	_, err := IDFromString("ffffffffffffffffffffffffffffffffffff")
	if err.Error() != "hex string must be 32 bytes" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestIDFromStringInvalidHex(t *testing.T) {
	_, err := IDFromString("ABC, du är mina tankar")
	if err.Error() != "cannot decode hex string as ID: encoding/hex: invalid byte: U+002C ','" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestIDFromString(t *testing.T) {
	str1 := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	id, _ := IDFromString(str1)
	str2 := id.String()
	if str2 != str1 {
		t.Errorf("unexpected string, got: %s, exp: %s", str2, str1)
	}
}

func TestIDFromBytes(t *testing.T) {
	b := []byte{123, 123, 123}
	exp := [32]byte{123, 123, 123}

	id := IDFromBytes(b)
	if !bytes.Equal(id[:], exp[:]) {
		t.Errorf("unexpected id, got:\n\t%v,\nexp:\n\t %v", id[:], exp)
	}
}

func TestIDWithPrefixGenerator_uniqueDistances(t *testing.T) {
	str := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	id, _ := IDFromString(str)

	i := 0
	for nextID := range IDWithPrefixGenerator(id) {
		numLeadingOnes := 0

		for _, b := range nextID {
			l := bits.LeadingZeros8(^uint8(b))

			numLeadingOnes += l
			if l != 8 {
				break
			}
		}

		if numLeadingOnes != i {
			t.Errorf("unexpected number of leading ones, got: %d, exp: %d",
				numLeadingOnes, i)
		}

		i++
	}
}
