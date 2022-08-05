package crypto

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"
)

// Range of characters for the different types of string generations
const charRange = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
const charRangeSpec = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
const charRangeHuman = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ23456789"
const charRangeSpecHuman = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ23456789\"#%*+-:;="

// Bitmask sizes for the string generators (based on 93 chars total)
const (
	letterIdxBits = 7                    // 7 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// RandomBytes uses the crypto/rand generator to generate random bytes. The amount of provided
// random bytes is controlled by the n argument
func RandomBytes(n int64) ([]byte, error) {
	if n < 1 {
		return []byte{}, fmt.Errorf("negative or zero byte size given")
	}
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// RandomStringSecure returns a random, n long string of characters. The character set is based
// on the s (special chars) and h (human readable) boolean arguments. This method uses the
// crypto/random package and therfore is cryptographically secure
func RandomStringSecure(n int, s, h bool) (string, error) {
	rs := strings.Builder{}
	rs.Grow(n)
	cr := getCharRange(s, h)
	crl := len(cr)

	rp := make([]byte, 8)
	_, err := rand.Read(rp)
	if err != nil {
		return rs.String(), err
	}
	for i, c, r := n-1, binary.BigEndian.Uint64(rp), letterIdxMax; i >= 0; {
		if r == 0 {
			_, err := rand.Read(rp)
			if err != nil {
				return rs.String(), err
			}
			c, r = binary.BigEndian.Uint64(rp), letterIdxMax
		}
		if idx := int(c & letterIdxMask); idx < crl {
			rs.WriteByte(cr[idx])
			i--
		}
		c >>= letterIdxBits
		r--
	}

	return rs.String(), nil
}

// RandNum returns a random number with a maximum value of n
func RandNum(n int) (int, error) {
	if n <= 0 {
		return 0, fmt.Errorf("provided number is <= 0: %d", n)
	}
	mbi := big.NewInt(int64(n))
	if !mbi.IsUint64() {
		return 0, fmt.Errorf("big.NewInt() generation returned negative value: %d", mbi)
	}
	rn64, err := rand.Int(rand.Reader, mbi)
	if err != nil {
		return 0, err
	}
	rn := int(rn64.Int64())
	if rn < 0 {
		return 0, fmt.Errorf("generated random number does not fit as int64: %d", rn64)
	}
	return rn, nil
}

// getCharRange returns the range of characters as controlled by the s and h bools
func getCharRange(s, h bool) string {
	var cr string
	if h {
		cr = charRangeHuman
		if s {
			cr = charRangeSpecHuman
		}
	}
	if !h {
		cr = charRange
		if s {
			cr = charRangeSpec
		}
	}
	return cr
}
