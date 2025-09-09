package unit

import (
	"testing"
	"time"

	"lms-manager/utils"
)

func TestHashPassword(t *testing.T) {
	password := "TestPassword123!"
	
	hash, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == password {
		t.Error("Hash should not be the same as password")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "TestPassword123!"
	
	hash, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Test correct password
	if !utils.CheckPasswordHash(password, hash) {
		t.Error("Password check should pass for correct password")
	}

	// Test incorrect password
	if utils.CheckPasswordHash("wrongpassword", hash) {
		t.Error("Password check should fail for incorrect password")
	}
}

func TestGenerateID(t *testing.T) {
	id1 := utils.GenerateID()
	id2 := utils.GenerateID()

	if id1 == "" {
		t.Error("Generated ID should not be empty")
	}

	if id2 == "" {
		t.Error("Generated ID should not be empty")
	}

	if id1 == id2 {
		t.Error("Generated IDs should be unique")
	}

	// Test ID format (should be hex string)
	if len(id1) != 32 {
		t.Errorf("Generated ID should be 32 characters long, got %d", len(id1))
	}
}

func TestGenerateToken(t *testing.T) {
	token1 := utils.GenerateToken()
	token2 := utils.GenerateToken()

	if token1 == "" {
		t.Error("Generated token should not be empty")
	}

	if token2 == "" {
		t.Error("Generated token should not be empty")
	}

	if token1 == token2 {
		t.Error("Generated tokens should be unique")
	}

	// Test token format (should be hex string)
	if len(token1) != 64 {
		t.Errorf("Generated token should be 64 characters long, got %d", len(token1))
	}
}

func TestIsValidEmail(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user.name@domain.co.uk",
		"admin@k2net.id",
		"support+test@company.org",
	}

	invalidEmails := []string{
		"invalid-email",
		"@domain.com",
		"user@",
		"user@domain",
		"",
		"user@domain.",
	}

	for _, email := range validEmails {
		if !utils.IsValidEmail(email) {
			t.Errorf("Email '%s' should be valid", email)
		}
	}

	for _, email := range invalidEmails {
		if utils.IsValidEmail(email) {
			t.Errorf("Email '%s' should be invalid", email)
		}
	}
}

func TestIsValidUsername(t *testing.T) {
	validUsernames := []string{
		"admin",
		"user123",
		"test_user",
		"operator",
		"viewer",
	}

	invalidUsernames := []string{
		"ab", // too short
		"a", // too short
		"", // empty
		"user@domain", // invalid character
		"user-name", // invalid character
		"user.name", // invalid character
		"user name", // space
		"user123456789012345678901234567890123456789012345678901", // too long
	}

	for _, username := range validUsernames {
		if !utils.IsValidUsername(username) {
			t.Errorf("Username '%s' should be valid", username)
		}
	}

	for _, username := range invalidUsernames {
		if utils.IsValidUsername(username) {
			t.Errorf("Username '%s' should be invalid", username)
		}
	}
}

func TestIsValidPassword(t *testing.T) {
	validPasswords := []string{
		"Password123!",
		"TestPass456@",
		"AdminPass789#",
		"SecurePass123$",
	}

	invalidPasswords := []string{
		"password", // no uppercase, digit, special
		"PASSWORD", // no lowercase, digit, special
		"Password", // no digit, special
		"Password123", // no special
		"Pass123!", // too short
		"", // empty
	}

	for _, password := range validPasswords {
		if !utils.IsValidPassword(password) {
			t.Errorf("Password '%s' should be valid", password)
		}
	}

	for _, password := range invalidPasswords {
		if utils.IsValidPassword(password) {
			t.Errorf("Password '%s' should be invalid", password)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TB"},
		{1536, "1.5 KB"},
		{2048, "2.0 KB"},
	}

	for _, test := range tests {
		result := utils.FormatBytes(test.bytes)
		if result != test.expected {
			t.Errorf("FormatBytes(%d) = %s, expected %s", test.bytes, result, test.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds  int64
		expected string
	}{
		{0, "0s"},
		{30, "30s"},
		{60, "1m 0s"},
		{90, "1m 30s"},
		{3600, "1h 0m"},
		{3661, "1h 1m 1s"},
		{86400, "1d 0h 0m"},
		{90061, "1d 1h 1m 1s"},
	}

	for _, test := range tests {
		result := utils.FormatDuration(test.seconds)
		if result != test.expected {
			t.Errorf("FormatDuration(%d) = %s, expected %s", test.seconds, result, test.expected)
		}
	}
}

func TestGetCurrentTimestamp(t *testing.T) {
	timestamp := utils.GetCurrentTimestamp()
	
	if timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}

	// Check if timestamp is recent (within last minute)
	now := time.Now()
	if timestamp.After(now) {
		t.Error("Timestamp should not be in the future")
	}

	if now.Sub(timestamp) > time.Minute {
		t.Error("Timestamp should be recent")
	}
}

func TestFormatTimestamp(t *testing.T) {
	timestamp := time.Date(2025, 1, 9, 15, 30, 45, 0, time.UTC)
	formatted := utils.FormatTimestamp(timestamp)
	
	expected := "2025-01-09 15:30:45"
	if formatted != expected {
		t.Errorf("FormatTimestamp() = %s, expected %s", formatted, expected)
	}
}

func TestParseTimestamp(t *testing.T) {
	timestampStr := "2025-01-09 15:30:45"
	timestamp, err := utils.ParseTimestamp(timestampStr)
	
	if err != nil {
		t.Fatalf("ParseTimestamp failed: %v", err)
	}

	expected := time.Date(2025, 1, 9, 15, 30, 45, 0, time.UTC)
	if !timestamp.Equal(expected) {
		t.Errorf("ParseTimestamp() = %v, expected %v", timestamp, expected)
	}
}

func TestIsExpired(t *testing.T) {
	now := time.Now()
	
	// Test expired timestamp
	expiredTime := now.Add(-2 * time.Hour)
	if !utils.IsExpired(expiredTime, time.Hour) {
		t.Error("Timestamp should be expired")
	}

	// Test not expired timestamp
	notExpiredTime := now.Add(-30 * time.Minute)
	if utils.IsExpired(notExpiredTime, time.Hour) {
		t.Error("Timestamp should not be expired")
	}
}

func TestGetExpirationTime(t *testing.T) {
	duration := time.Hour
	expirationTime := utils.GetExpirationTime(duration)
	
	now := time.Now()
	expected := now.Add(duration)
	
	// Allow 1 second difference for execution time
	if expirationTime.Sub(expected).Abs() > time.Second {
		t.Errorf("GetExpirationTime() = %v, expected %v", expirationTime, expected)
	}
}

func TestHashString(t *testing.T) {
	input := "test string"
	hash1 := utils.HashString(input)
	hash2 := utils.HashString(input)
	
	if hash1 == "" {
		t.Error("Hash should not be empty")
	}

	if hash1 != hash2 {
		t.Error("Same input should produce same hash")
	}

	// Test different input produces different hash
	hash3 := utils.HashString("different string")
	if hash1 == hash3 {
		t.Error("Different input should produce different hash")
	}
}
