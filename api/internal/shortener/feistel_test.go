package shortener

import "testing"

func TestShuffleUnshuffle(t *testing.T) {
	encoder := NewBase62Encoder(7)

	testNumbers := []uint64{1000001, 1000002, 1000003, 1234567, 9999999}

	for _, num := range testNumbers {
		shuffled := encoder.Shuffle(num)
		unshuffled := encoder.Unshuffle(shuffled)

		if unshuffled != num {
			t.Errorf("Round-trip failed: %d -> %d -> %d", num, shuffled, unshuffled)
		}

		// Shuffled should be different from original
		if shuffled == num {
			t.Errorf("Shuffle didn't change the number: %d", num)
		}
	}
}

func TestShuffleUnpredictable(t *testing.T) {
	encoder := NewBase62Encoder(7)

	// Sequential numbers should produce non-sequential results
	s1 := encoder.Shuffle(1000001)
	s2 := encoder.Shuffle(1000002)
	s3 := encoder.Shuffle(1000003)

	// Check they're all different
	if s1 == s2 || s2 == s3 || s1 == s3 {
		t.Errorf("Shuffle produced duplicate values")
	}

	// Check they're not sequential
	if s2 == s1+1 || s3 == s2+1 {
		t.Errorf("Shuffle produced sequential values: %d, %d, %d", s1, s2, s3)
	}

	t.Logf("1000001 -> %d", s1)
	t.Logf("1000002 -> %d", s2)
	t.Logf("1000003 -> %d", s3)
}
