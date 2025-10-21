package security

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"unicode"
)

// T041 [US3]: PasswordPolicy struct defines password requirements
type PasswordPolicy struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireDigit     bool
	RequireSymbol    bool
}

// T042 [US3]: DefaultPasswordPolicy constant (12 chars, all requirements true)
// FR-016: Minimum 12 characters with uppercase, lowercase, digit, and symbol
var DefaultPasswordPolicy = PasswordPolicy{
	MinLength:        12,
	RequireUppercase: true,
	RequireLowercase: true,
	RequireDigit:     true,
	RequireSymbol:    true,
}

// PasswordStrength represents the strength level of a password
type PasswordStrength int

const (
	PasswordStrengthWeak PasswordStrength = iota
	PasswordStrengthMedium
	PasswordStrengthStrong
)

func (s PasswordStrength) String() string {
	switch s {
	case PasswordStrengthWeak:
		return "Weak"
	case PasswordStrengthMedium:
		return "Medium"
	case PasswordStrengthStrong:
		return "Strong"
	default:
		return "Unknown"
	}
}

// T043 [US3]: Validate method validates password against policy
// FR-016: Return descriptive error messages for each failed requirement
func (p *PasswordPolicy) Validate(password []byte) error {
	if password == nil {
		return fmt.Errorf("password cannot be empty (must be at least %d characters)", p.MinLength)
	}

	// Convert to rune slice for proper Unicode handling
	runes := []rune(string(password))

	// Check minimum length (count runes, not bytes)
	if len(runes) < p.MinLength {
		return fmt.Errorf("password must be at least %d characters long (got %d)", p.MinLength, len(runes))
	}

	// Track which requirements are met
	var hasUpper, hasLower, hasDigit, hasSymbol bool

	for _, r := range runes {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			hasSymbol = true
		}
	}

	// Check requirements and return descriptive errors
	if p.RequireUppercase && !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if p.RequireLowercase && !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if p.RequireDigit && !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	if p.RequireSymbol && !hasSymbol {
		return errors.New("password must contain at least one special character or symbol")
	}

	return nil
}

// T044 [US3]: Strength method calculates password strength
// FR-017: Calculate weak/medium/strong based on length and character variety
// Algorithm per data-model.md:186-238
func (p *PasswordPolicy) Strength(password []byte) PasswordStrength {
	if len(password) == 0 {
		return PasswordStrengthWeak
	}

	// Convert to rune slice for proper Unicode handling
	runes := []rune(string(password))
	length := len(runes)

	// Calculate character variety score
	var hasUpper, hasLower, hasDigit, hasSymbol bool
	symbolCount := 0

	for _, r := range runes {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			hasSymbol = true
			symbolCount++
		}
	}

	// Count character types present
	typeCount := 0
	if hasUpper {
		typeCount++
	}
	if hasLower {
		typeCount++
	}
	if hasDigit {
		typeCount++
	}
	if hasSymbol {
		typeCount++
	}

	// Strength calculation based on length and variety
	// Weak: < 12 characters OR missing required types
	// Medium: 12-19 characters with all required types
	// Strong: 20+ characters with all required types OR exceptional variety

	if length < 12 || typeCount < 4 {
		return PasswordStrengthWeak
	}

	if length >= 25 || (length >= 20 && symbolCount >= 3) {
		return PasswordStrengthStrong
	}

	if length >= 16 && typeCount == 4 {
		return PasswordStrengthMedium
	}

	return PasswordStrengthWeak
}

// T051a [US3]: ValidationRateLimiter prevents brute-force password guessing
// FR-024: Enforce 5-second cooldown after 3rd validation failure
type ValidationRateLimiter struct {
	mu            sync.Mutex
	failureCount  int
	lastFailure   time.Time
	cooldownUntil time.Time
}

// NewValidationRateLimiter creates a new rate limiter
func NewValidationRateLimiter() *ValidationRateLimiter {
	return &ValidationRateLimiter{}
}

// CheckAndRecordFailure checks if rate limiting is active and records a failure
// Returns error if in cooldown period, nil otherwise
func (rl *ValidationRateLimiter) CheckAndRecordFailure() error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Check if currently in cooldown
	if now.Before(rl.cooldownUntil) {
		remaining := time.Until(rl.cooldownUntil).Round(time.Second)
		return fmt.Errorf("too many failed attempts - please wait %v before trying again", remaining)
	}

	// Reset counter if last failure was > 30 seconds ago
	if now.Sub(rl.lastFailure) > 30*time.Second {
		rl.failureCount = 0
	}

	// Increment failure count
	rl.failureCount++
	rl.lastFailure = now

	// Trigger cooldown after 3rd failure
	if rl.failureCount >= 3 {
		rl.cooldownUntil = now.Add(5 * time.Second)
		rl.failureCount = 0 // Reset after triggering cooldown
		return fmt.Errorf("too many failed attempts - please wait 5 seconds before trying again")
	}

	return nil
}

// Reset clears the rate limiter state (for successful validation)
func (rl *ValidationRateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.failureCount = 0
	rl.cooldownUntil = time.Time{}
}
