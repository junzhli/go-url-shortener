package util

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	base62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

// NewUUID generates a random UUID according to RFC 4122
func NewUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

func Base10FromBase62(str string) (uint64, error) {
	power := uint64(1)
	res := uint64(0)
	for _, ch := range str {
		index := strings.Index(base62, string(ch))
		if index == -1 {
			return 0, errors.New("invalid base62 format string")
		}

		res += power * uint64(index)
		power *= 62
	}
	return res, nil
}

func Base62FromBase10(n uint64) string {
	var str strings.Builder
	for n > 62 {
		rem := n % 62
		n /= 62
		str.WriteString(string(base62[rem]))
	}
	str.WriteString(string(base62[n%62]))
	return str.String()
}
