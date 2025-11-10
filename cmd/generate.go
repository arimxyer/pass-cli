package cmd

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var (
	genLength      int
	genNoLower     bool
	genNoUpper     bool
	genNoDigits    bool
	genNoSymbols   bool
	genNoClipboard bool
)

const (
	lowerChars    = "abcdefghijklmnopqrstuvwxyz"
	upperChars    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars    = "0123456789"
	symbolChars   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	minLength     = 8
	maxLength     = 128
	defaultLength = 20
)

var generateCmd = &cobra.Command{
	Use:     "generate",
	GroupID: "security",
	Aliases: []string{"gen", "pwd"},
	Short:   "Generate a cryptographically secure password",
	Long: `Generate creates a strong, random password using cryptographic randomness.

By default, generates a 20-character password with lowercase, uppercase,
digits, and symbols. You can customize the length and character sets.

The generated password is automatically copied to the clipboard and
displayed on screen.`,
	Example: `  # Generate default password (20 chars, all character types)
  pass-cli generate

  # Generate 32-character password
  pass-cli generate --length 32

  # Generate password without symbols (alphanumeric only)
  pass-cli generate --no-symbols

  # Generate digits-only PIN (8 digits)
  pass-cli gen --length 8 --no-lower --no-upper --no-symbols

  # Generate without clipboard
  pass-cli generate --no-clipboard`,
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().IntVarP(&genLength, "length", "l", defaultLength, "password length")
	generateCmd.Flags().BoolVar(&genNoLower, "no-lower", false, "exclude lowercase letters")
	generateCmd.Flags().BoolVar(&genNoUpper, "no-upper", false, "exclude uppercase letters")
	generateCmd.Flags().BoolVar(&genNoDigits, "no-digits", false, "exclude digits")
	generateCmd.Flags().BoolVar(&genNoSymbols, "no-symbols", false, "exclude symbols")
	generateCmd.Flags().BoolVar(&genNoClipboard, "no-clipboard", false, "do not copy to clipboard")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Validate length
	if genLength < minLength {
		return fmt.Errorf("password length must be at least %d characters", minLength)
	}
	if genLength > maxLength {
		return fmt.Errorf("password length cannot exceed %d characters", maxLength)
	}

	// Build character set
	charset := ""
	if !genNoLower {
		charset += lowerChars
	}
	if !genNoUpper {
		charset += upperChars
	}
	if !genNoDigits {
		charset += digitChars
	}
	if !genNoSymbols {
		charset += symbolChars
	}

	// Validate character set
	if charset == "" {
		return fmt.Errorf("must include at least one character type")
	}

	// Generate password
	password, err := generatePassword(genLength, charset)
	if err != nil {
		return fmt.Errorf("failed to generate password: %w", err)
	}

	// Display password
	fmt.Printf("🔐 Generated Password:\n")
	fmt.Printf("   %s\n\n", password)

	// Show password strength info
	entropy := calculateEntropy(genLength, len(charset))
	fmt.Printf("📊 Strength: %.1f bits of entropy\n", entropy)
	fmt.Printf("📏 Length: %d characters\n", genLength)
	fmt.Printf("🔤 Character types: ")

	var types []string
	if !genNoLower {
		types = append(types, "lowercase")
	}
	if !genNoUpper {
		types = append(types, "uppercase")
	}
	if !genNoDigits {
		types = append(types, "digits")
	}
	if !genNoSymbols {
		types = append(types, "symbols")
	}
	fmt.Printf("%s\n", joinWithCommas(types))

	// Copy to clipboard
	if !genNoClipboard {
		if err := clipboard.WriteAll(password); err != nil {
			fmt.Fprintf(os.Stderr, "\n⚠️  Warning: failed to copy to clipboard: %v\n", err)
		} else {
			fmt.Println("\n✅ Password copied to clipboard!")
		}
	}

	return nil
}

// generatePassword creates a cryptographically secure random password
// Guarantees at least one character from each enabled character type
func generatePassword(length int, charset string) (string, error) {
	password := make([]byte, length)

	// Build list of required character sets
	var requiredSets []string
	if !genNoLower {
		requiredSets = append(requiredSets, lowerChars)
	}
	if !genNoUpper {
		requiredSets = append(requiredSets, upperChars)
	}
	if !genNoDigits {
		requiredSets = append(requiredSets, digitChars)
	}
	if !genNoSymbols {
		requiredSets = append(requiredSets, symbolChars)
	}

	// Step 1: Ensure at least one char from each required set
	for i, reqSet := range requiredSets {
		if i >= length {
			break // Not enough space for all required types
		}
		setLen := big.NewInt(int64(len(reqSet)))
		randomIndex, err := rand.Int(rand.Reader, setLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		password[i] = reqSet[randomIndex.Int64()]
	}

	// Step 2: Fill remaining positions with random chars from full charset
	charsetLen := big.NewInt(int64(len(charset)))
	for i := len(requiredSets); i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		password[i] = charset[randomIndex.Int64()]
	}

	// Step 3: Shuffle password to avoid predictable positions
	for i := length - 1; i > 0; i-- {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		j := randomIndex.Int64()
		password[i], password[j] = password[j], password[i]
	}

	return string(password), nil
}

// calculateEntropy calculates password entropy in bits
func calculateEntropy(length int, charsetSize int) float64 {
	if charsetSize <= 0 || length <= 0 {
		return 0
	}
	// Entropy = log2(charsetSize^length) = length * log2(charsetSize)
	// Using bit length as approximation for log2
	bits := 0
	n := charsetSize
	for n > 1 {
		n >>= 1
		bits++
	}
	return float64(length) * float64(bits)
}

func joinWithCommas(items []string) string {
	if len(items) == 0 {
		return ""
	}
	if len(items) == 1 {
		return items[0]
	}
	if len(items) == 2 {
		return items[0] + " and " + items[1]
	}
	result := ""
	for i, item := range items {
		if i == len(items)-1 {
			result += "and " + item
		} else {
			result += item + ", "
		}
	}
	return result
}
