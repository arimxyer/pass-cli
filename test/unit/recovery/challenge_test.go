package recovery_test

import (
	"testing"

	"pass-cli/internal/recovery"
)

// T014: Unit test for SelectVerifyPositions()
// Tests: randomness verified over 10+ attempts (SC-009), uniqueness, count

func TestSelectVerifyPositions(t *testing.T) {
	t.Run("returns correct count of positions", func(t *testing.T) {
		positions, err := recovery.SelectVerifyPositions(3)
		if err != nil {
			t.Fatalf("SelectVerifyPositions(3) failed: %v", err)
		}

		if len(positions) != 3 {
			t.Errorf("Expected 3 positions, got %d", len(positions))
		}
	})

	t.Run("positions are within valid range [0-23]", func(t *testing.T) {
		positions, err := recovery.SelectVerifyPositions(6)
		if err != nil {
			t.Fatalf("SelectVerifyPositions(6) failed: %v", err)
		}

		for i, pos := range positions {
			if pos < 0 || pos >= 24 {
				t.Errorf("Position %d has invalid value %d (must be 0-23)", i, pos)
			}
		}
	})

	t.Run("positions are unique (no duplicates)", func(t *testing.T) {
		positions, err := recovery.SelectVerifyPositions(6)
		if err != nil {
			t.Fatalf("SelectVerifyPositions(6) failed: %v", err)
		}

		seen := make(map[int]bool)
		for _, pos := range positions {
			if seen[pos] {
				t.Errorf("Duplicate position found: %d", pos)
			}
			seen[pos] = true
		}
	})

	t.Run("positions are sorted in ascending order", func(t *testing.T) {
		positions, err := recovery.SelectVerifyPositions(6)
		if err != nil {
			t.Fatalf("SelectVerifyPositions(6) failed: %v", err)
		}

		for i := 1; i < len(positions); i++ {
			if positions[i] <= positions[i-1] {
				t.Errorf("Positions not sorted: %v", positions)
				break
			}
		}
	})

	t.Run("produces different positions across multiple calls (randomness - SC-009)", func(t *testing.T) {
		// Per SC-009: verify randomness over 10+ attempts
		const attempts = 15
		results := make([][]int, attempts)

		for i := 0; i < attempts; i++ {
			positions, err := recovery.SelectVerifyPositions(6)
			if err != nil {
				t.Fatalf("Attempt %d failed: %v", i+1, err)
			}
			results[i] = positions
		}

		// Verify at least some results are different
		allSame := true
		for i := 1; i < attempts; i++ {
			if !slicesEqual(results[0], results[i]) {
				allSame = false
				break
			}
		}

		if allSame {
			t.Error("All 15 attempts produced identical positions (randomness failure)")
		}

		// Count unique first positions (should have variety)
		firstPositions := make(map[int]int)
		for _, result := range results {
			firstPositions[result[0]]++
		}

		if len(firstPositions) < 3 {
			t.Errorf("Insufficient randomness: only %d unique first positions across %d attempts", len(firstPositions), attempts)
		}
	})

	t.Run("handles edge case: requesting all 24 positions", func(t *testing.T) {
		positions, err := recovery.SelectVerifyPositions(24)
		if err != nil {
			t.Fatalf("SelectVerifyPositions(24) failed: %v", err)
		}

		if len(positions) != 24 {
			t.Errorf("Expected 24 positions, got %d", len(positions))
		}

		// Should be [0, 1, 2, ..., 23] when sorted
		for i := 0; i < 24; i++ {
			if positions[i] != i {
				t.Errorf("Position %d has value %d (expected %d)", i, positions[i], i)
			}
		}
	})

	t.Run("returns error for invalid count (> 24)", func(t *testing.T) {
		_, err := recovery.SelectVerifyPositions(25)
		if err == nil {
			t.Error("Expected error for count > 24, got nil")
		}
	})

	t.Run("returns error for invalid count (0)", func(t *testing.T) {
		_, err := recovery.SelectVerifyPositions(0)
		if err == nil {
			t.Error("Expected error for count = 0, got nil")
		}
	})

	t.Run("returns error for negative count", func(t *testing.T) {
		_, err := recovery.SelectVerifyPositions(-1)
		if err == nil {
			t.Error("Expected error for negative count, got nil")
		}
	})
}

// Helper function to compare two slices
func slicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// T032: Unit test for ShuffleChallengePositions()
// Tests: non-destructive shuffling, randomness verified over 10+ attempts (SC-009), preserves all positions

func TestShuffleChallengePositions(t *testing.T) {
	t.Run("returns same length as input", func(t *testing.T) {
		input := []int{0, 5, 10, 15, 20, 23}
		shuffled := recovery.ShuffleChallengePositions(input)

		if len(shuffled) != len(input) {
			t.Errorf("Expected length %d, got %d", len(input), len(shuffled))
		}
	})

	t.Run("preserves all positions (no data loss)", func(t *testing.T) {
		input := []int{0, 5, 10, 15, 20, 23}
		shuffled := recovery.ShuffleChallengePositions(input)

		// Convert to maps for comparison
		inputSet := make(map[int]bool)
		shuffledSet := make(map[int]bool)

		for _, pos := range input {
			inputSet[pos] = true
		}
		for _, pos := range shuffled {
			shuffledSet[pos] = true
		}

		// Verify all input positions are in shuffled
		for _, pos := range input {
			if !shuffledSet[pos] {
				t.Errorf("Position %d from input missing in shuffled result", pos)
			}
		}

		// Verify no extra positions in shuffled
		for _, pos := range shuffled {
			if !inputSet[pos] {
				t.Errorf("Position %d in shuffled result not in input", pos)
			}
		}
	})

	t.Run("is non-destructive (does not modify input slice)", func(t *testing.T) {
		input := []int{0, 5, 10, 15, 20, 23}
		original := make([]int, len(input))
		copy(original, input)

		_ = recovery.ShuffleChallengePositions(input)

		// Verify input slice unchanged
		if !slicesEqual(input, original) {
			t.Errorf("Input slice was modified: original=%v, after=%v", original, input)
		}
	})

	t.Run("produces different orderings across multiple calls (randomness - SC-009)", func(t *testing.T) {
		// Per SC-009: verify randomness over 10+ attempts
		const attempts = 15
		input := []int{0, 5, 10, 15, 20, 23}
		results := make([][]int, attempts)

		for i := 0; i < attempts; i++ {
			shuffled := recovery.ShuffleChallengePositions(input)
			results[i] = shuffled
		}

		// Verify at least some results are different
		allSame := true
		for i := 1; i < attempts; i++ {
			if !slicesEqual(results[0], results[i]) {
				allSame = false
				break
			}
		}

		if allSame {
			t.Error("All 15 attempts produced identical shuffles (randomness failure)")
		}

		// Count unique first positions (should have variety)
		firstPositions := make(map[int]int)
		for _, result := range results {
			if len(result) > 0 {
				firstPositions[result[0]]++
			}
		}

		if len(firstPositions) < 3 {
			t.Errorf("Insufficient randomness: only %d unique first positions across %d attempts", len(firstPositions), attempts)
		}
	})

	t.Run("handles single-element slice", func(t *testing.T) {
		input := []int{42}
		shuffled := recovery.ShuffleChallengePositions(input)

		if len(shuffled) != 1 {
			t.Errorf("Expected length 1, got %d", len(shuffled))
		}

		if shuffled[0] != 42 {
			t.Errorf("Expected [42], got %v", shuffled)
		}
	})

	t.Run("handles empty slice", func(t *testing.T) {
		input := []int{}
		shuffled := recovery.ShuffleChallengePositions(input)

		if len(shuffled) != 0 {
			t.Errorf("Expected length 0, got %d", len(shuffled))
		}
	})

	t.Run("handles two-element slice", func(t *testing.T) {
		input := []int{1, 2}

		// Run multiple times to check we get both orderings
		const attempts = 20
		foundOriginal := false
		foundReversed := false

		for i := 0; i < attempts; i++ {
			shuffled := recovery.ShuffleChallengePositions(input)

			if slicesEqual(shuffled, []int{1, 2}) {
				foundOriginal = true
			}
			if slicesEqual(shuffled, []int{2, 1}) {
				foundReversed = true
			}
		}

		if !foundOriginal && !foundReversed {
			t.Error("No valid shuffling found over 20 attempts")
		}
	})
}
