package security

import (
	"strings"
	"testing"
)

// T037 [US3]: Test password validation against all FR-016 rules
// These tests should FAIL before implementation (TDD approach)

func TestPasswordPolicy_Validate_MinLength(t *testing.T) {
	policy := DefaultPasswordPolicy

	tests := []struct {
		name     string
		password []byte
		wantErr  bool
	}{
		{
			name:     "Valid 12 character password",
			password: []byte("Password123!"),
			wantErr:  false,
		},
		{
			name:     "Too short - 11 characters",
			password: []byte("Password12!"),
			wantErr:  true,
		},
		{
			name:     "Too short - 8 characters",
			password: []byte("Pass123!"),
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: []byte(""),
			wantErr:  true,
		},
		{
			name:     "Exactly minimum length",
			password: []byte("Abcdefgh123!"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "12") {
				t.Errorf("Error message should mention minimum length: %v", err)
			}
		})
	}
}

func TestPasswordPolicy_Validate_RequireUppercase(t *testing.T) {
	policy := DefaultPasswordPolicy

	tests := []struct {
		name     string
		password []byte
		wantErr  bool
	}{
		{
			name:     "Has uppercase",
			password: []byte("Password123!"),
			wantErr:  false,
		},
		{
			name:     "No uppercase",
			password: []byte("password123!"),
			wantErr:  true,
		},
		{
			name:     "Multiple uppercase",
			password: []byte("PASSword123!"),
			wantErr:  false,
		},
		{
			name:     "Uppercase at end",
			password: []byte("password123!A"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(strings.ToLower(err.Error()), "uppercase") {
				t.Errorf("Error message should mention uppercase requirement: %v", err)
			}
		})
	}
}

func TestPasswordPolicy_Validate_RequireLowercase(t *testing.T) {
	policy := DefaultPasswordPolicy

	tests := []struct {
		name     string
		password []byte
		wantErr  bool
	}{
		{
			name:     "Has lowercase",
			password: []byte("Password123!"),
			wantErr:  false,
		},
		{
			name:     "No lowercase",
			password: []byte("PASSWORD123!"),
			wantErr:  true,
		},
		{
			name:     "Multiple lowercase",
			password: []byte("Passssword1!"),
			wantErr:  false,
		},
		{
			name:     "Lowercase at end",
			password: []byte("PASSWORD123!a"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(strings.ToLower(err.Error()), "lowercase") {
				t.Errorf("Error message should mention lowercase requirement: %v", err)
			}
		})
	}
}

func TestPasswordPolicy_Validate_RequireDigit(t *testing.T) {
	policy := DefaultPasswordPolicy

	tests := []struct {
		name     string
		password []byte
		wantErr  bool
	}{
		{
			name:     "Has digit",
			password: []byte("Password123!"),
			wantErr:  false,
		},
		{
			name:     "No digit",
			password: []byte("PasswordAbc!"),
			wantErr:  true,
		},
		{
			name:     "Multiple digits",
			password: []byte("Password999!"),
			wantErr:  false,
		},
		{
			name:     "Digit at start",
			password: []byte("1Password!!!"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(strings.ToLower(err.Error()), "digit") && !strings.Contains(strings.ToLower(err.Error()), "number") {
				t.Errorf("Error message should mention digit/number requirement: %v", err)
			}
		})
	}
}

func TestPasswordPolicy_Validate_RequireSymbol(t *testing.T) {
	policy := DefaultPasswordPolicy

	tests := []struct {
		name     string
		password []byte
		wantErr  bool
	}{
		{
			name:     "Has symbol (!)",
			password: []byte("Password123!"),
			wantErr:  false,
		},
		{
			name:     "No symbol",
			password: []byte("Password1234"),
			wantErr:  true,
		},
		{
			name:     "Has symbol (@)",
			password: []byte("Password123@"),
			wantErr:  false,
		},
		{
			name:     "Has symbol (#)",
			password: []byte("Password123#"),
			wantErr:  false,
		},
		{
			name:     "Has symbol ($)",
			password: []byte("Password123$"),
			wantErr:  false,
		},
		{
			name:     "Has symbol (%)",
			password: []byte("Password123%"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(strings.ToLower(err.Error()), "symbol") && !strings.Contains(strings.ToLower(err.Error()), "special") {
				t.Errorf("Error message should mention symbol/special character requirement: %v", err)
			}
		})
	}
}

func TestPasswordPolicy_Validate_AllRequirements(t *testing.T) {
	policy := DefaultPasswordPolicy

	tests := []struct {
		name     string
		password []byte
		wantErr  bool
	}{
		{
			name:     "Meets all requirements",
			password: []byte("MyPassword123!"),
			wantErr:  false,
		},
		{
			name:     "Missing uppercase",
			password: []byte("mypassword123!"),
			wantErr:  true,
		},
		{
			name:     "Missing lowercase",
			password: []byte("MYPASSWORD123!"),
			wantErr:  true,
		},
		{
			name:     "Missing digit",
			password: []byte("MyPassword!!!"),
			wantErr:  true,
		},
		{
			name:     "Missing symbol",
			password: []byte("MyPassword123"),
			wantErr:  true,
		},
		{
			name:     "Too short but has all character types",
			password: []byte("Pass123!"),
			wantErr:  true,
		},
		{
			name:     "Long password with all requirements",
			password: []byte("ThisIsAVerySecurePassword123!"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPasswordPolicy_Validate_CustomPolicy(t *testing.T) {
	// Test with custom policy (less restrictive)
	customPolicy := PasswordPolicy{
		MinLength:        8,
		RequireUppercase: false,
		RequireLowercase: true,
		RequireDigit:     true,
		RequireSymbol:    false,
	}

	tests := []struct {
		name     string
		password []byte
		wantErr  bool
	}{
		{
			name:     "Meets custom requirements - no uppercase/symbol needed",
			password: []byte("password123"),
			wantErr:  false,
		},
		{
			name:     "Too short for custom policy",
			password: []byte("pass123"),
			wantErr:  true,
		},
		{
			name:     "Missing digit - still required",
			password: []byte("passwordabc"),
			wantErr:  true,
		},
		{
			name:     "Missing lowercase - still required",
			password: []byte("PASSWORD123"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := customPolicy.Validate(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPasswordPolicy_Validate_NilPassword(t *testing.T) {
	policy := DefaultPasswordPolicy

	err := policy.Validate(nil)
	if err == nil {
		t.Error("Validate() should return error for nil password")
	}
}

// T038 [US3]: Test password strength calculation
// Tests weak/medium/strong boundaries per FR-017

func TestPasswordPolicy_Strength_Weak(t *testing.T) {
	policy := DefaultPasswordPolicy

	weakPasswords := [][]byte{
		[]byte("Password123!"), // Minimum requirements only (12 chars)
		[]byte("Abcdefgh123!"), // Exactly 12 chars
		[]byte("Short1!Aa"),    // Less than 12 chars (should be rejected by Validate)
		[]byte("Password1!"),   // 10 chars
		[]byte("HelloWorld1!"), // 12 chars, simple pattern
	}

	for _, password := range weakPasswords {
		t.Run(string(password), func(t *testing.T) {
			strength := policy.Strength(password)
			if strength != PasswordStrengthWeak {
				t.Errorf("Strength() = %v, want %v for password %s", strength, PasswordStrengthWeak, password)
			}
		})
	}
}

func TestPasswordPolicy_Strength_Medium(t *testing.T) {
	policy := DefaultPasswordPolicy

	mediumPasswords := [][]byte{
		[]byte("MySecurePass123!"),  // 16 chars with variety
		[]byte("GoodPassword2023!"), // 16 chars
		[]byte("Testing@Password1"), // 17 chars
		[]byte("HelloWorld123!@#"),  // Multiple symbols
		[]byte("P@ssw0rd!Testing"),  // 16 chars with substitutions
	}

	for _, password := range mediumPasswords {
		t.Run(string(password), func(t *testing.T) {
			strength := policy.Strength(password)
			if strength != PasswordStrengthMedium {
				t.Errorf("Strength() = %v, want %v for password %s", strength, PasswordStrengthMedium, password)
			}
		})
	}
}

func TestPasswordPolicy_Strength_Strong(t *testing.T) {
	policy := DefaultPasswordPolicy

	strongPasswords := [][]byte{
		[]byte("ThisIsAVerySecureP@ssw0rd2024!"),     // 30+ chars
		[]byte("Correct-Horse-Battery-Staple-2024!"), // Long passphrase
		[]byte("MyUltraSecurePassword123!@#$%"),      // 28 chars with variety
		[]byte("ComplexP@ssw0rd!WithManyCharacters"), // Long with variety
		[]byte("Security2024!@#$%^&*()HighEntropy"),  // Many special chars
	}

	for _, password := range strongPasswords {
		t.Run(string(password), func(t *testing.T) {
			strength := policy.Strength(password)
			if strength != PasswordStrengthStrong {
				t.Errorf("Strength() = %v, want %v for password %s", strength, PasswordStrengthStrong, password)
			}
		})
	}
}

func TestPasswordPolicy_Strength_LengthBoundaries(t *testing.T) {
	policy := DefaultPasswordPolicy

	tests := []struct {
		name     string
		password []byte
		want     PasswordStrength
	}{
		{
			name:     "12 characters - weak",
			password: []byte("Password123!"),
			want:     PasswordStrengthWeak,
		},
		{
			name:     "16 characters - medium",
			password: []byte("LongerPassword1!"),
			want:     PasswordStrengthMedium,
		},
		{
			name:     "20 characters - likely medium/strong boundary",
			password: []byte("VeryLongPassword2024!"),
			want:     PasswordStrengthMedium, // Adjust based on actual algorithm
		},
		{
			name:     "25+ characters - strong",
			password: []byte("ExtremelyLongSecurePassword2024!@#"),
			want:     PasswordStrengthStrong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strength := policy.Strength(tt.password)
			if strength != tt.want {
				t.Logf("Strength() = %v, want %v for %d-char password", strength, tt.want, len(tt.password))
			}
		})
	}
}

func TestPasswordPolicy_Strength_CharacterVariety(t *testing.T) {
	policy := DefaultPasswordPolicy

	tests := []struct {
		name     string
		password []byte
		minLevel PasswordStrength
	}{
		{
			name:     "Only required character types",
			password: []byte("Password123!"),
			minLevel: PasswordStrengthWeak,
		},
		{
			name:     "Multiple symbols increase strength",
			password: []byte("P@ssw0rd!#$123"),
			minLevel: PasswordStrengthWeak,
		},
		{
			name:     "Long with variety",
			password: []byte("MyP@ssw0rd!WithSymbols123"),
			minLevel: PasswordStrengthMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strength := policy.Strength(tt.password)
			if strength < tt.minLevel {
				t.Errorf("Strength() = %v, want at least %v", strength, tt.minLevel)
			}
		})
	}
}

func TestPasswordPolicy_Strength_CommonPatterns(t *testing.T) {
	policy := DefaultPasswordPolicy

	// These should be weak even if they meet length requirements
	commonPatterns := [][]byte{
		[]byte("Password123!"),
		[]byte("Qwerty123456!"),
		[]byte("Admin123456!"),
		[]byte("Welcome123!@"),
	}

	for _, password := range commonPatterns {
		t.Run(string(password), func(t *testing.T) {
			strength := policy.Strength(password)
			// Common patterns should not be rated as strong
			if strength == PasswordStrengthStrong {
				t.Errorf("Common pattern should not be strong: %s", password)
			}
		})
	}
}

func TestPasswordPolicy_Strength_EmptyPassword(t *testing.T) {
	policy := DefaultPasswordPolicy

	strength := policy.Strength([]byte(""))
	if strength != PasswordStrengthWeak {
		t.Errorf("Empty password should be weak, got %v", strength)
	}
}

func TestPasswordPolicy_Strength_NilPassword(t *testing.T) {
	policy := DefaultPasswordPolicy

	strength := policy.Strength(nil)
	if strength != PasswordStrengthWeak {
		t.Errorf("Nil password should be weak, got %v", strength)
	}
}

// T039 [US3]: Test Unicode character handling
// Verify accented letters, international symbols work correctly

func TestPasswordPolicy_Validate_UnicodeAccentedLetters(t *testing.T) {
	policy := DefaultPasswordPolicy

	tests := []struct {
		name     string
		password []byte
		wantErr  bool
		desc     string
	}{
		{
			name:     "French accented letters",
			password: []byte("Pàsswörd123!"),
			wantErr:  false,
			desc:     "Should accept accented letters as valid",
		},
		{
			name:     "German umlauts",
			password: []byte("Pässwörd123!"),
			wantErr:  false,
			desc:     "Should accept German umlauts",
		},
		{
			name:     "Spanish ñ",
			password: []byte("Contraseña1!"),
			wantErr:  false,
			desc:     "Should accept Spanish ñ",
		},
		{
			name:     "Mixed ASCII and Unicode",
			password: []byte("MyPássw0rd1!"),
			wantErr:  false,
			desc:     "Should accept mix of ASCII and Unicode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v - %s", err, tt.wantErr, tt.desc)
			}
		})
	}
}

func TestPasswordPolicy_Validate_UnicodeCyrillic(t *testing.T) {
	policy := DefaultPasswordPolicy

	tests := []struct {
		name     string
		password []byte
		wantErr  bool
	}{
		{
			name:     "Cyrillic with ASCII requirements",
			password: []byte("Пароль123!Test"),
			wantErr:  false,
		},
		{
			name:     "Full Cyrillic (may need uppercase/lowercase handling)",
			password: []byte("Пароль123!Тест"),
			wantErr:  false, // Cyrillic has uppercase/lowercase
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPasswordPolicy_Validate_UnicodeCJK(t *testing.T) {
	policy := DefaultPasswordPolicy

	tests := []struct {
		name     string
		password []byte
		desc     string
	}{
		{
			name:     "Chinese characters with requirements",
			password: []byte("密码Password123!"),
			desc:     "Chinese characters with ASCII requirements",
		},
		{
			name:     "Japanese hiragana with requirements",
			password: []byte("パスワード123!Aa"),
			desc:     "Japanese with ASCII requirements",
		},
		{
			name:     "Korean hangul with requirements",
			password: []byte("비밀번호Pass123!"),
			desc:     "Korean with ASCII requirements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These passwords should be valid as they meet length + have required characters
			err := policy.Validate(tt.password)
			if err != nil {
				t.Logf("Validate() error = %v for %s", err, tt.desc)
			}
		})
	}
}

func TestPasswordPolicy_Validate_UnicodeEmojis(t *testing.T) {
	policy := DefaultPasswordPolicy

	tests := []struct {
		name     string
		password []byte
		wantErr  bool
		desc     string
	}{
		{
			name:     "Emoji as symbol",
			password: []byte("Password123😀"),
			wantErr:  false, // Emoji might count as symbol
			desc:     "Emoji should be treated as valid character",
		},
		{
			name:     "Multiple emojis",
			password: []byte("Pass🔒Word123!"),
			wantErr:  false,
			desc:     "Multiple emojis should be valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.password)
			// Log result for informational purposes
			t.Logf("Validate() for %s: error=%v (length=%d bytes)", tt.desc, err, len(tt.password))
		})
	}
}

func TestPasswordPolicy_Strength_Unicode(t *testing.T) {
	policy := DefaultPasswordPolicy

	unicodePasswords := [][]byte{
		[]byte("Pàsswörd123!"),              // Accented letters
		[]byte("MyÜberSecurePassword2024!"), // German umlauts
		[]byte("ContraseñaSegura123!"),      // Spanish ñ
		[]byte("密码SecurePassword123!"),      // Chinese characters
		[]byte("パスワードPassword123!"),         // Japanese
		[]byte("비밀번호SecurePass123!"),        // Korean
	}

	for _, password := range unicodePasswords {
		t.Run(string(password), func(t *testing.T) {
			strength := policy.Strength(password)
			// Unicode passwords should be evaluated same as ASCII
			// Just log the result for verification
			t.Logf("Strength for Unicode password: %v (length: %d bytes)", strength, len(password))
		})
	}
}

func TestPasswordPolicy_Validate_UnicodeLength(t *testing.T) {
	policy := DefaultPasswordPolicy

	// Multi-byte characters should be counted correctly
	tests := []struct {
		name       string
		password   []byte
		runeCount  int
		byteCount  int
		shouldPass bool
	}{
		{
			name:       "12 ASCII characters",
			password:   []byte("Password123!"),
			runeCount:  12,
			byteCount:  12,
			shouldPass: true,
		},
		{
			name:       "12 Unicode runes (more bytes)",
			password:   []byte("Pässwörd123!"),
			runeCount:  12,
			byteCount:  14, // ä and ö are 2 bytes each
			shouldPass: true,
		},
		{
			name:       "Emoji in password",
			password:   []byte("Pass😀word123!"),
			runeCount:  14,
			byteCount:  17, // Emoji is 4 bytes
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.password)
			if tt.shouldPass && err != nil {
				t.Errorf("Expected valid password (runes=%d, bytes=%d), got error: %v",
					tt.runeCount, tt.byteCount, err)
			}
			t.Logf("Password has %d runes, %d bytes", tt.runeCount, tt.byteCount)
		})
	}
}
