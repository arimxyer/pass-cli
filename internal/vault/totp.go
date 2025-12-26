package vault

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

// TimeSyncThreshold is the maximum acceptable clock drift in seconds
const TimeSyncThreshold = 30

// TimeSyncResult holds the result of a time sync check
type TimeSyncResult struct {
	InSync     bool          // Whether the clock is within acceptable drift
	Drift      time.Duration // Estimated drift from server time
	Checked    bool          // Whether the check was performed
	Error      error         // Any error during the check
	ServerTime time.Time     // The server's reported time
	LocalTime  time.Time     // Local time when check was performed
}

// CheckTimeSync verifies that the local system clock is reasonably accurate
// by comparing against an HTTP server's Date header.
// Returns a TimeSyncResult with drift information.
func CheckTimeSync() TimeSyncResult {
	result := TimeSyncResult{
		LocalTime: time.Now(),
	}

	// Use a reliable, fast endpoint - just need the Date header
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// HEAD request to minimize data transfer
	resp, err := client.Head("https://www.google.com")
	if err != nil {
		result.Error = err
		return result
	}
	defer func() { _ = resp.Body.Close() }()

	result.Checked = true

	// Parse the Date header
	dateHeader := resp.Header.Get("Date")
	if dateHeader == "" {
		result.Error = fmt.Errorf("no Date header in response")
		return result
	}

	serverTime, err := http.ParseTime(dateHeader)
	if err != nil {
		result.Error = fmt.Errorf("failed to parse Date header: %w", err)
		return result
	}

	result.ServerTime = serverTime
	result.Drift = result.LocalTime.Sub(serverTime)

	// Check if drift is within acceptable range
	if result.Drift < 0 {
		result.Drift = -result.Drift // Absolute value
	}
	result.InSync = result.Drift <= TimeSyncThreshold*time.Second

	return result
}

// FormatTimeSyncWarning returns a user-friendly warning message if time is out of sync
func FormatTimeSyncWarning(result TimeSyncResult) string {
	if !result.Checked {
		if result.Error != nil {
			return fmt.Sprintf("⚠️  Could not verify time sync: %v", result.Error)
		}
		return ""
	}

	if result.InSync {
		return "" // No warning needed
	}

	direction := "ahead"
	if result.LocalTime.Before(result.ServerTime) {
		direction = "behind"
	}

	return fmt.Sprintf("⚠️  Warning: System clock is %v %s. TOTP codes may not work!\n"+
		"   Please sync your system time.", result.Drift.Round(time.Second), direction)
}

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

	// Safely convert period - cap to reasonable range then convert
	// TOTP periods are typically 30s, max 5 minutes is reasonable
	keyPeriod := key.Period()
	periodInt := 30 // Default to standard period
	if keyPeriod > 0 && keyPeriod <= 300 {
		// #nosec G115 -- keyPeriod is bounds-checked to max 300, safe for int
		periodInt = int(keyPeriod)
	}

	config := &TOTPConfig{
		Secret:  key.Secret(),
		Issuer:  key.Issuer(),
		Account: key.AccountName(),
		Period:  periodInt,
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

	// Check for valid base32 characters (A-Z, 2-7, =)
	for _, c := range secret {
		isUpperAlpha := c >= 'A' && c <= 'Z'
		isValidDigit := c >= '2' && c <= '7'
		isPadding := c == '='
		if !isUpperAlpha && !isValidDigit && !isPadding {
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

	// Determine period with bounds check (TOTP periods are typically 30s, max 5 min)
	period := uint(30)
	if cred.TOTPPeriod > 0 && cred.TOTPPeriod <= 300 {
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
	// period is bounds-checked above (max 300), safe to convert
	epoch := now.Unix()
	// #nosec G115 -- period is bounds-checked to max 300, safe for int
	periodInt := int(period)
	remaining := periodInt - int(epoch%int64(periodInt))

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
//
// The URI follows the otpauth:// format:
// otpauth://totp/ISSUER:ACCOUNT?secret=SECRET&issuer=ISSUER&algorithm=ALG&digits=N&period=N
func (c *Credential) BuildTOTPURI() (string, error) {
	if c.TOTPSecret == "" {
		return "", fmt.Errorf("no TOTP configured for this credential")
	}

	// Get parameters with defaults
	algorithm := strings.ToUpper(c.TOTPAlgorithm)
	if algorithm == "" {
		algorithm = "SHA1"
	}
	digits := c.TOTPDigits
	if digits == 0 {
		digits = 6
	}
	period := c.TOTPPeriod
	if period == 0 {
		period = 30
	}
	issuer := c.TOTPIssuer
	if issuer == "" {
		issuer = c.Service
	}
	account := c.Username
	if account == "" {
		account = c.Service
	}

	// Build label (issuer:account or just account)
	// The label appears in the authenticator app as the credential name
	label := url.PathEscape(account)
	if issuer != "" {
		label = url.PathEscape(issuer) + ":" + url.PathEscape(account)
	}

	// Build query parameters
	// secret is already base32 encoded - use directly without re-encoding
	params := url.Values{}
	params.Set("secret", c.TOTPSecret)
	if issuer != "" {
		params.Set("issuer", issuer)
	}
	// Only include non-default values to keep URI clean
	if algorithm != "SHA1" {
		params.Set("algorithm", algorithm)
	}
	if digits != 6 {
		params.Set("digits", strconv.Itoa(digits))
	}
	if period != 30 {
		params.Set("period", strconv.Itoa(period))
	}

	return fmt.Sprintf("otpauth://totp/%s?%s", label, params.Encode()), nil
}

// DisplayQRCode displays a QR code in the terminal for the credential's TOTP configuration
// This allows users to scan the code with their authenticator app
func (c *Credential) DisplayQRCode(writer *os.File) error {
	uri, err := c.BuildTOTPURI()
	if err != nil {
		return err
	}

	config := qrterminal.Config{
		Level:     qrterminal.L,
		Writer:    writer,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
		QuietZone: 1,
	}

	qrterminal.GenerateWithConfig(uri, config)
	return nil
}

// ExportQRCode exports a QR code to a PNG file for the credential's TOTP configuration
// size is the image width/height in pixels (e.g., 256)
func (c *Credential) ExportQRCode(filename string, size int) error {
	uri, err := c.BuildTOTPURI()
	if err != nil {
		return err
	}

	return qrcode.WriteFile(uri, qrcode.Medium, size, filename)
}
