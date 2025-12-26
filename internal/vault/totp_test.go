package vault

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func TestParseTOTPURI_ValidFullURI(t *testing.T) {
	uri := "otpauth://totp/GitHub:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=GitHub&algorithm=SHA256&digits=6&period=30"

	config, err := ParseTOTPURI(uri)
	if err != nil {
		t.Fatalf("ParseTOTPURI failed: %v", err)
	}

	if config.Secret != "JBSWY3DPEHPK3PXP" {
		t.Errorf("expected secret JBSWY3DPEHPK3PXP, got %s", config.Secret)
	}
	if config.Issuer != "GitHub" {
		t.Errorf("expected issuer GitHub, got %s", config.Issuer)
	}
	if config.Algorithm != "SHA256" {
		t.Errorf("expected algorithm SHA256, got %s", config.Algorithm)
	}
	if config.Digits != 6 {
		t.Errorf("expected digits 6, got %d", config.Digits)
	}
	if config.Period != 30 {
		t.Errorf("expected period 30, got %d", config.Period)
	}
}

func TestParseTOTPURI_MinimalURI(t *testing.T) {
	uri := "otpauth://totp/service?secret=JBSWY3DPEHPK3PXP"

	config, err := ParseTOTPURI(uri)
	if err != nil {
		t.Fatalf("ParseTOTPURI failed: %v", err)
	}

	// Check defaults are applied
	if config.Algorithm != "SHA1" {
		t.Errorf("expected default algorithm SHA1, got %s", config.Algorithm)
	}
	if config.Digits != 6 {
		t.Errorf("expected default digits 6, got %d", config.Digits)
	}
	if config.Period != 30 {
		t.Errorf("expected default period 30, got %d", config.Period)
	}
}

func TestParseTOTPURI_RawBase32Secret(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"

	config, err := ParseTOTPURI(secret)
	if err != nil {
		t.Fatalf("ParseTOTPURI failed: %v", err)
	}

	if config.Secret != "JBSWY3DPEHPK3PXP" {
		t.Errorf("expected secret JBSWY3DPEHPK3PXP, got %s", config.Secret)
	}
	if config.Algorithm != "SHA1" {
		t.Errorf("expected default algorithm SHA1, got %s", config.Algorithm)
	}
}

func TestParseTOTPURI_LowercaseSecret(t *testing.T) {
	// Google Authenticator sometimes uses lowercase
	secret := "jbswy3dpehpk3pxp"

	config, err := ParseTOTPURI(secret)
	if err != nil {
		t.Fatalf("ParseTOTPURI failed: %v", err)
	}

	// Should be normalized to uppercase
	if config.Secret != "JBSWY3DPEHPK3PXP" {
		t.Errorf("expected uppercase secret JBSWY3DPEHPK3PXP, got %s", config.Secret)
	}
}

func TestParseTOTPURI_InvalidScheme(t *testing.T) {
	uri := "https://totp/service?secret=JBSWY3DPEHPK3PXP"

	_, err := ParseTOTPURI(uri)
	if err == nil {
		t.Error("expected error for invalid scheme, got nil")
	}
}

func TestParseTOTPURI_HOTPNotSupported(t *testing.T) {
	uri := "otpauth://hotp/service?secret=JBSWY3DPEHPK3PXP&counter=0"

	_, err := ParseTOTPURI(uri)
	if err == nil {
		t.Error("expected error for HOTP type, got nil")
	}
	if !strings.Contains(err.Error(), "hotp") {
		t.Errorf("expected error to mention hotp, got: %v", err)
	}
}

func TestParseTOTPURI_InvalidBase32(t *testing.T) {
	secret := "INVALID!@#SECRET"

	_, err := ParseTOTPURI(secret)
	if err == nil {
		t.Error("expected error for invalid base32, got nil")
	}
}

func TestValidateTOTPSecret_Valid(t *testing.T) {
	secrets := []string{
		"JBSWY3DPEHPK3PXP",
		"GEZDGNBVGY3TQOJQ",
		"jbswy3dpehpk3pxp", // lowercase should be accepted
	}

	for _, secret := range secrets {
		err := ValidateTOTPSecret(secret)
		if err != nil {
			t.Errorf("ValidateTOTPSecret(%q) failed: %v", secret, err)
		}
	}
}

func TestValidateTOTPSecret_Empty(t *testing.T) {
	err := ValidateTOTPSecret("")
	if err == nil {
		t.Error("expected error for empty secret, got nil")
	}
}

func TestValidateTOTPSecret_InvalidChars(t *testing.T) {
	err := ValidateTOTPSecret("INVALID!SECRET")
	if err == nil {
		t.Error("expected error for invalid characters, got nil")
	}
}

func TestGenerateTOTPCode_Success(t *testing.T) {
	cred := &Credential{
		Service:       "test",
		TOTPSecret:    "JBSWY3DPEHPK3PXP",
		TOTPAlgorithm: "SHA1",
		TOTPDigits:    6,
		TOTPPeriod:    30,
	}

	code, remaining, err := GenerateTOTPCode(cred)
	if err != nil {
		t.Fatalf("GenerateTOTPCode failed: %v", err)
	}

	// Code should be 6 digits
	if len(code) != 6 {
		t.Errorf("expected 6-digit code, got %d digits: %s", len(code), code)
	}

	// Remaining should be between 1 and 30
	if remaining < 1 || remaining > 30 {
		t.Errorf("expected remaining between 1-30, got %d", remaining)
	}
}

func TestGenerateTOTPCode_NoTOTPConfigured(t *testing.T) {
	cred := &Credential{
		Service: "test",
	}

	_, _, err := GenerateTOTPCode(cred)
	if err == nil {
		t.Error("expected error for no TOTP configured, got nil")
	}
}

func TestGenerateTOTPCode_8Digits(t *testing.T) {
	cred := &Credential{
		Service:       "test",
		TOTPSecret:    "JBSWY3DPEHPK3PXP",
		TOTPAlgorithm: "SHA1",
		TOTPDigits:    8,
		TOTPPeriod:    30,
	}

	code, _, err := GenerateTOTPCode(cred)
	if err != nil {
		t.Fatalf("GenerateTOTPCode failed: %v", err)
	}

	if len(code) != 8 {
		t.Errorf("expected 8-digit code, got %d digits: %s", len(code), code)
	}
}

func TestGenerateTOTPCode_SHA256(t *testing.T) {
	cred := &Credential{
		Service:       "test",
		TOTPSecret:    "JBSWY3DPEHPK3PXP",
		TOTPAlgorithm: "SHA256",
		TOTPDigits:    6,
		TOTPPeriod:    30,
	}

	code, _, err := GenerateTOTPCode(cred)
	if err != nil {
		t.Fatalf("GenerateTOTPCode failed: %v", err)
	}

	if len(code) != 6 {
		t.Errorf("expected 6-digit code, got %d digits: %s", len(code), code)
	}
}

func TestCredential_HasTOTP(t *testing.T) {
	tests := []struct {
		name     string
		cred     *Credential
		expected bool
	}{
		{
			name:     "with TOTP",
			cred:     &Credential{TOTPSecret: "JBSWY3DPEHPK3PXP"},
			expected: true,
		},
		{
			name:     "without TOTP",
			cred:     &Credential{},
			expected: false,
		},
		{
			name:     "empty secret",
			cred:     &Credential{TOTPSecret: ""},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cred.HasTOTP(); got != tt.expected {
				t.Errorf("HasTOTP() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCredential_SetTOTPFromURI(t *testing.T) {
	cred := &Credential{Service: "github"}
	uri := "otpauth://totp/GitHub:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=GitHub&algorithm=SHA256&digits=8&period=60"

	err := cred.SetTOTPFromURI(uri)
	if err != nil {
		t.Fatalf("SetTOTPFromURI failed: %v", err)
	}

	if cred.TOTPSecret != "JBSWY3DPEHPK3PXP" {
		t.Errorf("expected secret JBSWY3DPEHPK3PXP, got %s", cred.TOTPSecret)
	}
	if cred.TOTPAlgorithm != "SHA256" {
		t.Errorf("expected algorithm SHA256, got %s", cred.TOTPAlgorithm)
	}
	if cred.TOTPDigits != 8 {
		t.Errorf("expected digits 8, got %d", cred.TOTPDigits)
	}
	if cred.TOTPPeriod != 60 {
		t.Errorf("expected period 60, got %d", cred.TOTPPeriod)
	}
	if cred.TOTPIssuer != "GitHub" {
		t.Errorf("expected issuer GitHub, got %s", cred.TOTPIssuer)
	}
}

func TestCredential_ClearTOTP(t *testing.T) {
	cred := &Credential{
		Service:       "github",
		TOTPSecret:    "JBSWY3DPEHPK3PXP",
		TOTPAlgorithm: "SHA256",
		TOTPDigits:    8,
		TOTPPeriod:    60,
		TOTPIssuer:    "GitHub",
	}

	cred.ClearTOTP()

	if cred.TOTPSecret != "" {
		t.Errorf("expected empty secret, got %s", cred.TOTPSecret)
	}
	if cred.TOTPAlgorithm != "" {
		t.Errorf("expected empty algorithm, got %s", cred.TOTPAlgorithm)
	}
	if cred.TOTPDigits != 0 {
		t.Errorf("expected digits 0, got %d", cred.TOTPDigits)
	}
	if cred.TOTPPeriod != 0 {
		t.Errorf("expected period 0, got %d", cred.TOTPPeriod)
	}
	if cred.TOTPIssuer != "" {
		t.Errorf("expected empty issuer, got %s", cred.TOTPIssuer)
	}
}

func TestDefaultTOTPConfig(t *testing.T) {
	config := DefaultTOTPConfig()

	if config.Algorithm != "SHA1" {
		t.Errorf("expected default algorithm SHA1, got %s", config.Algorithm)
	}
	if config.Digits != 6 {
		t.Errorf("expected default digits 6, got %d", config.Digits)
	}
	if config.Period != 30 {
		t.Errorf("expected default period 30, got %d", config.Period)
	}
}

func TestFormatTimeSyncWarning_InSync(t *testing.T) {
	result := TimeSyncResult{
		Checked: true,
		InSync:  true,
		Drift:   5 * time.Second,
	}

	warning := FormatTimeSyncWarning(result)
	if warning != "" {
		t.Errorf("expected no warning for in-sync time, got: %s", warning)
	}
}

func TestFormatTimeSyncWarning_OutOfSync(t *testing.T) {
	result := TimeSyncResult{
		Checked:    true,
		InSync:     false,
		Drift:      2 * time.Minute,
		LocalTime:  time.Now(),
		ServerTime: time.Now().Add(-2 * time.Minute),
	}

	warning := FormatTimeSyncWarning(result)
	if warning == "" {
		t.Error("expected warning for out-of-sync time")
	}
	if !strings.Contains(warning, "Warning") {
		t.Errorf("expected warning message to contain 'Warning', got: %s", warning)
	}
}

func TestFormatTimeSyncWarning_CheckFailed(t *testing.T) {
	result := TimeSyncResult{
		Checked: false,
		Error:   fmt.Errorf("network error"),
	}

	warning := FormatTimeSyncWarning(result)
	if warning == "" {
		t.Error("expected warning for failed check")
	}
	if !strings.Contains(warning, "Could not verify") {
		t.Errorf("expected 'Could not verify' in warning, got: %s", warning)
	}
}

func TestFormatTimeSyncWarning_NotChecked(t *testing.T) {
	result := TimeSyncResult{
		Checked: false,
	}

	warning := FormatTimeSyncWarning(result)
	if warning != "" {
		t.Errorf("expected no warning when not checked and no error, got: %s", warning)
	}
}

func TestBuildTOTPURI_SecretMatchesCodeGeneration(t *testing.T) {
	// This test ensures that the secret in the generated QR code URI
	// matches the secret used for TOTP code generation.
	// This was a bug where totp.Generate() re-encoded the already-encoded secret.
	secret := "JBSWY3DPEHPK3PXP"
	cred := &Credential{
		Service:    "TestService",
		Username:   "testuser",
		TOTPSecret: secret,
	}

	// Build the URI
	uri, err := cred.BuildTOTPURI()
	if err != nil {
		t.Fatalf("BuildTOTPURI failed: %v", err)
	}

	// Verify the URI contains the exact same secret (not re-encoded)
	if !strings.Contains(uri, "secret="+secret) {
		t.Errorf("URI should contain original secret=%s, got URI: %s", secret, uri)
	}

	// Parse the URI back and verify codes match
	parsedKey, err := otp.NewKeyFromURL(uri)
	if err != nil {
		t.Fatalf("Failed to parse generated URI: %v", err)
	}

	// The secret from the parsed URI should match the original
	if parsedKey.Secret() != secret {
		t.Errorf("Parsed secret=%s does not match original=%s", parsedKey.Secret(), secret)
	}

	// Most importantly: codes generated from both should match
	now := time.Now()
	codeFromCred, _, err := cred.GetTOTPCode()
	if err != nil {
		t.Fatalf("GetTOTPCode failed: %v", err)
	}

	codeFromURI, err := totp.GenerateCode(parsedKey.Secret(), now)
	if err != nil {
		t.Fatalf("GenerateCode from URI failed: %v", err)
	}

	if codeFromCred != codeFromURI {
		t.Errorf("Code mismatch! Credential generated %s, URI would generate %s", codeFromCred, codeFromURI)
	}
}

func TestBuildTOTPURI_NoTOTPConfigured(t *testing.T) {
	cred := &Credential{
		Service: "test",
	}

	_, err := cred.BuildTOTPURI()
	if err == nil {
		t.Error("expected error for no TOTP configured, got nil")
	}
}

func TestBuildTOTPURI_DefaultValues(t *testing.T) {
	cred := &Credential{
		Service:    "github",
		Username:   "user@example.com",
		TOTPSecret: "JBSWY3DPEHPK3PXP",
	}

	uri, err := cred.BuildTOTPURI()
	if err != nil {
		t.Fatalf("BuildTOTPURI failed: %v", err)
	}

	// Should have the secret
	if !strings.Contains(uri, "secret=JBSWY3DPEHPK3PXP") {
		t.Errorf("URI missing secret, got: %s", uri)
	}

	// Should use service as issuer when not specified
	if !strings.Contains(uri, "issuer=github") {
		t.Errorf("URI should use service as issuer, got: %s", uri)
	}

	// Should NOT have algorithm param (default SHA1 is omitted)
	if strings.Contains(uri, "algorithm=") {
		t.Errorf("URI should not include default algorithm=SHA1, got: %s", uri)
	}

	// Should NOT have digits param (default 6 is omitted)
	if strings.Contains(uri, "digits=") {
		t.Errorf("URI should not include default digits=6, got: %s", uri)
	}

	// Should NOT have period param (default 30 is omitted)
	if strings.Contains(uri, "period=") {
		t.Errorf("URI should not include default period=30, got: %s", uri)
	}
}

func TestBuildTOTPURI_NonDefaultValues(t *testing.T) {
	cred := &Credential{
		Service:       "github",
		Username:      "user@example.com",
		TOTPSecret:    "JBSWY3DPEHPK3PXP",
		TOTPAlgorithm: "SHA256",
		TOTPDigits:    8,
		TOTPPeriod:    60,
		TOTPIssuer:    "GitHub Inc",
	}

	uri, err := cred.BuildTOTPURI()
	if err != nil {
		t.Fatalf("BuildTOTPURI failed: %v", err)
	}

	// Parse the URI to check values
	parsedURL, err := url.Parse(uri)
	if err != nil {
		t.Fatalf("Failed to parse URI: %v", err)
	}

	params := parsedURL.Query()

	if params.Get("algorithm") != "SHA256" {
		t.Errorf("expected algorithm=SHA256, got: %s", params.Get("algorithm"))
	}
	if params.Get("digits") != "8" {
		t.Errorf("expected digits=8, got: %s", params.Get("digits"))
	}
	if params.Get("period") != "60" {
		t.Errorf("expected period=60, got: %s", params.Get("period"))
	}
	if params.Get("issuer") != "GitHub Inc" {
		t.Errorf("expected issuer='GitHub Inc', got: %s", params.Get("issuer"))
	}
}

func TestBuildTOTPURI_SpecialCharactersInLabel(t *testing.T) {
	cred := &Credential{
		Service:    "My Service",
		Username:   "user@example.com",
		TOTPSecret: "JBSWY3DPEHPK3PXP",
		TOTPIssuer: "Company Name",
	}

	uri, err := cred.BuildTOTPURI()
	if err != nil {
		t.Fatalf("BuildTOTPURI failed: %v", err)
	}

	// Should be a valid URI
	_, err = url.Parse(uri)
	if err != nil {
		t.Errorf("Generated URI is not valid: %v", err)
	}

	// Should start with otpauth://totp/
	if !strings.HasPrefix(uri, "otpauth://totp/") {
		t.Errorf("URI should start with otpauth://totp/, got: %s", uri)
	}
}
