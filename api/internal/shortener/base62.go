package shortener

import (
	"errors"
	"strings"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const base = 62

var ErrInvalidCode = errors.New("invalid short code")

type Base62Encoder struct {
	minLength int
}

func NewBase62Encoder(minLength int) *Base62Encoder {
	return &Base62Encoder{
		minLength: minLength,
	}
}

func (e *Base62Encoder) Encode(num uint64) string {
	if num == 0 {
		return e.padLeft("0")
	}

	var result strings.Builder

	for num > 0 {
		remainder := num % base
		result.WriteByte(base62Chars[remainder])
		num = num / base
	}

	encoded := e.reverse(result.String())
	return e.padLeft(encoded)
}

func (e *Base62Encoder) Decode(code string) (uint64, error) {
	if len(code) == 0 {
		return 0, ErrInvalidCode
	}

	var result uint64 = 0

	for _, char := range code {
		index := strings.IndexRune(base62Chars, char)
		if index == -1 {
			return 0, ErrInvalidCode
		}
		result = result*base + uint64(index)
	}

	return result, nil
}

func (e *Base62Encoder) reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func (e *Base62Encoder) padLeft(s string) string {
	if len(s) >= e.minLength {
		return s
	}
	return strings.Repeat("0", e.minLength-len(s)) + s
}
