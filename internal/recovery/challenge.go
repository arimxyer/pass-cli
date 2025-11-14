package recovery

import (
	"crypto/rand"
	"math/big"
	mathrand "math/rand"
	"sort"
	"strings"
)

// selectChallengePositions generates crypto-secure random unique positions
// Parameters: totalWords (24), count (6 for challenge, 3 for verify)
// Returns: sorted array of unique positions, error
func selectChallengePositions(totalWords, count int) ([]int, error) {
	// Validate inputs
	if count <= 0 {
		return nil, ErrInvalidCount
	}
	if count > totalWords {
		return nil, ErrInvalidCount
	}

	// Use crypto-secure random number generation
	positions := make([]int, 0, count)
	seen := make(map[int]bool)

	for len(positions) < count {
		// Generate random number in range [0, totalWords)
		n, err := rand.Int(rand.Reader, big.NewInt(int64(totalWords)))
		if err != nil {
			return nil, ErrRandomGeneration
		}

		pos := int(n.Int64())

		// Only add if not already selected (ensure uniqueness)
		if !seen[pos] {
			positions = append(positions, pos)
			seen[pos] = true
		}
	}

	// Sort positions in ascending order
	sort.Ints(positions)

	return positions, nil
}

// SelectVerifyPositions randomly selects positions for backup verification
// Parameters: count (number of positions to select, e.g., 3)
// Returns: sorted array of random positions [0-23], error
func SelectVerifyPositions(count int) ([]int, error) {
	return selectChallengePositions(MnemonicWords, count)
}

// splitWords splits 24-word mnemonic into challenge words and stored words
// Parameters: mnemonic (full 24 words), challengePos (indices to extract)
// Returns: challenge words (6), stored words (18)
func splitWords(mnemonic string, challengePos []int) (challenge, stored []string) {
	// Split mnemonic into individual words
	words := strings.Fields(mnemonic)

	// Create map of challenge positions for quick lookup
	challengeMap := make(map[int]bool)
	for _, pos := range challengePos {
		challengeMap[pos] = true
	}

	// Separate words into challenge and stored arrays
	challenge = make([]string, 0, len(challengePos))
	stored = make([]string, 0, len(words)-len(challengePos))

	for i, word := range words {
		if challengeMap[i] {
			challenge = append(challenge, word)
		} else {
			stored = append(stored, word)
		}
	}

	return challenge, stored
}

// ShuffleChallengePositions randomizes order of challenge positions for recovery prompts
// Parameters: positions (fixed challenge positions from metadata)
// Returns: shuffled positions (non-destructive, creates copy)
func ShuffleChallengePositions(positions []int) []int {
	// Create copy to avoid modifying input (non-destructive)
	shuffled := make([]int, len(positions))
	copy(shuffled, positions)

	// Fisher-Yates shuffle using math/rand (non-crypto)
	// Note: This is intentionally non-crypto random for UI presentation order only
	// The actual security relies on crypto/rand for position selection in selectChallengePositions
	for i := len(shuffled) - 1; i > 0; i-- {
		j := mathrand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled
}

// reconstructMnemonic combines challenge words + stored words into full 24-word phrase
// Parameters: challengeWords (6), challengePos (indices), storedWords (18)
// Returns: full 24-word mnemonic, error
func reconstructMnemonic(challengeWords []string, challengePos []int, storedWords []string) (string, error) {
	// Validate inputs
	if len(challengeWords) != ChallengeCount {
		return "", ErrInvalidCount
	}
	if len(storedWords) != (MnemonicWords - ChallengeCount) {
		return "", ErrInvalidCount
	}
	if len(challengePos) != ChallengeCount {
		return "", ErrInvalidCount
	}

	// Create map of challenge positions for quick lookup
	challengeMap := make(map[int]string)
	for i, pos := range challengePos {
		challengeMap[pos] = challengeWords[i]
	}

	// Build full 24-word mnemonic
	words := make([]string, MnemonicWords)
	storedIdx := 0

	for i := 0; i < MnemonicWords; i++ {
		if word, isChallenge := challengeMap[i]; isChallenge {
			// This position is a challenge word
			words[i] = word
		} else {
			// This position is a stored word
			if storedIdx >= len(storedWords) {
				return "", ErrInvalidCount
			}
			words[i] = storedWords[storedIdx]
			storedIdx++
		}
	}

	// Join into single string
	mnemonic := strings.Join(words, " ")

	return mnemonic, nil
}
