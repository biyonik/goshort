package shortener

// Shuffle encrypts a number using Feistel cipher
// Makes sequential IDs unpredictable
func (e *Base62Encoder) Shuffle(num uint64) uint64 {
	rounds := 3
	key := uint64(31337)
	mod := uint64(10000)

	l := num / mod
	r := num % mod

	for i := 0; i < rounds; i++ {
		fR := (r*31 + key) % mod
		l, r = r, l^fR
	}

	return l*mod + r
}

// Unshuffle decrypts a number back to original
func (e *Base62Encoder) Unshuffle(num uint64) uint64 {
	rounds := 3
	key := uint64(31337)
	mod := uint64(10000)

	l := num / mod
	r := num % mod

	for i := rounds - 1; i >= 0; i-- {
		fL := (l*31 + key) % mod
		l, r = r^fL, l
	}

	return l*mod + r
}
