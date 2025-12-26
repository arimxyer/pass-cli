package vault

import (
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPConfig holds parsed TOTP configuration from an otpauth:// URI
type TOTPConfig struct {
	Secret    string // Base32 encoded secret
	Algorithm string // SHA1, SHA256, SHA512
	Digits    int    // 6 or 8
	Period    int    // seconds
	Issuer    string // Service/issuer name
	Account   string // Account name (usually email)
}

// DefaultTOTPConfig returns default TOTP configuration values
func DefaultTOTPConfig() TOTPConfig {
	return TOTPConfig{
		Algorithm: "SHA1",
		Digits:    6,
		Period:    30,
	}
}

// ParseTOTPURI parses an otpauth:// URI and returns TOTP configuration
// Supports both otpauth://totp/... and raw base32 secrets
func ParseTOTPURI(uri string) (*TOTPConfig, error) {
	uri = strings.TrimSpace(uri)

	// Check if it's a raw base32 secret (no otpauth:// prefix)
	if !strings.HasPrefix(strings.ToLower(uri), "otpauth://") {
		// Validate as base32 secret
		if err := ValidateTOTPSecret(uri); err != nil {
			return nil, err
		}
		config := DefaultTOTPConfig()
		config.Secret = strings.ToUpper(uri)
		return &config, nil
	}

	// Parse as otpauth:// URI using the library
	key, err := otp.NewKeyFromURL(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid otpauth URI: %w", err)
	}

	// Validate it's a TOTP URI (not HOTP)
	if key.Type() != "totp" {
		return nil, fmt.Errorf("unsupported OTP type: %s (only totp is supported)", key.Type())
	}

	config := &TOTPConfig{
		Secret:  key.Secret(),
		Issuer:  key.Issuer(),
		Account: key.AccountName(),
		Period:  int(key.Period()),
		Digits:  key.Digits().Length(),
	}

	// Map algorithm
	switch key.Algorithm() {
	case otp.AlgorithmSHA1:
		config.Algorithm = "SHA1"
	case otp.AlgorithmSHA256:
		config.Algorithm = "SHA256"
	case otp.AlgorithmSHA512:
		config.Algorithm = "SHA512"
	default:
		config.Algorithm = "SHA1"
	}

	// Apply defaults for zero values
	if config.Period == 0 {
		config.Period = 30
	}
	if config.Digits == 0 {
		config.Digits = 6
	}

	return config, nil
}

// ValidateTOTPSecret validates that a string is a valid base32 TOTP secret
func ValidateTOTPSecret(secret string) error {
	secret = strings.TrimSpace(strings.ToUpper(secret))
	if secret == "" {
		return fmt.Errorf("TOTP secret cannot be empty")
	}

	// Check for valid base32 characters (A-Z, 2-7)
	for _, c := range secret {
		if !((c >= 'A' && c <= 'Z') || (c >= '2' && c <= '7') || c == '=') {
			return fmt.Errorf("invalid base32 character in TOTP secret: %c", c)
		}
	}

	// Try to generate a code to fully validate the secret
	_, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		return fmt.Errorf("invalid TOTP secret: %w", err)
	}

	return nil
}

// GenerateTOTPCode generates a TOTP code for the given credential
// Returns the code and remaining validity in seconds
func GenerateTOTPCode(cred *Credential) (string, int, error) {
	if cred.TOTPSecret == "" {
		return "", 0, fmt.Errorf("no TOTP configured for this credential")
	}

	// Determine algorithm
	algo := otp.AlgorithmSHA1
	switch strings.ToUpper(cred.TOTPAlgorithm) {
	case "SHA256":
		algo = otp.AlgorithmSHA256
	case "SHA512":
		algo = otp.AlgorithmSHA512
	}

	// Determine digits
	digits := otp.DigitsSix
	if cred.TOTPDigits == 8 {
		digits = otp.DigitsEight
	}

	// Determine period
	period := uint(30)
	if cred.TOTPPeriod > 0 {
		period = uint(cred.TOTPPeriod)
	}

	// Generate code
	now := time.Now()
	code, err := totp.GenerateCodeCustom(cred.TOTPSecret, now, totp.ValidateOpts{
		Period:    period,
		Digits:    digits,
		Algorithm: algo,
	})
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate TOTP code: %w", err)
	}

	// Calculate remaining validity
	epoch := now.Unix()
	remaining := int(period) - int(epoch%int64(period))

	return code, remaining, nil
}

// HasTOTP returns true if the credential has TOTP configured
func (c *Credential) HasTOTP() bool {
	return c.TOTPSecret != ""
}

// GetTOTPCode generates and returns the current TOTP code for this credential
func (c *Credential) GetTOTPCode() (string, int, error) {
	return GenerateTOTPCode(c)
}

// SetTOTPFromURI parses a TOTP URI and sets the credential's TOTP fields
func (c *Credential) SetTOTPFromURI(uri string) error {
	config, err := ParseTOTPURI(uri)
	if err != nil {
		return err
	}

	c.TOTPSecret = config.Secret
	c.TOTPAlgorithm = config.Algorithm
	c.TOTPDigits = config.Digits
	c.TOTPPeriod = config.Period
	if config.Issuer != "" {
		c.TOTPIssuer = config.Issuer
	}

	return nil
}

// ClearTOTP removes all TOTP configuration from the credential
func (c *Credential) ClearTOTP() {
	c.TOTPSecret = ""
	c.TOTPAlgorithm = ""
	c.TOTPDigits = 0
	c.TOTPPeriod = 0
	c.TOTPIssuer = ""
}

// BuildTOTPURI constructs an otpauth:// URI from the credential's TOTP config
// Useful for exporting or displaying QR codes
func (c *Credential) BuildTOTPURI() (string, error) {
	if c.TOTPSecret == "" {
		return "", fmt.Errorf("no TOTP configured for this credential")
	}

	// Determine algorithm
	algo := otp.AlgorithmSHA1
	switch strings.ToUpper(c.TOTPAlgorithm) {
	case "SHA256":
		algo = otp.AlgorithmSHA256
	case "SHA512":
		algo = otp.AlgorithmSHA512
	}

	// Determine digits
	digits := otp.DigitsSix
	if c.TOTPDigits == 8 {
		digits = otp.DigitsEight
	}

	// Determine period
	period := uint(30)
	if c.TOTPPeriod > 0 {
		period = uint(c.TOTPPeriod)
	}

	// Determine issuer and account
	issuer := c.TOTPIssuer
	if issuer == "" {
		issuer = c.Service
	}
	account := c.Username
	if account == "" {
		account = c.Service
	}

	// Generate key with the library
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: account,
		Period:      period,
		Digits:      digits,
		Algorithm:   algo,
		Secret:      []byte(c.TOTPSecret), // Will be re-encoded
	})
	if err != nil {
		return "", fmt.Errorf("failed to build TOTP URI: %w", err)
	}

	return key.URL(), nil
}
