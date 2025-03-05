package utils

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestComparePassword(t *testing.T) {
	// Hash a sample password for testing
	hash, err := HashPassword("testPassword")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	tests := []struct {
		name     string
		hash     string
		password string
		expected bool
	}{
		{"Correct password", string(hash), "testPassword", true},
		{"Incorrect password", string(hash), "wrongPassword", false},
		{"Empty password", string(hash), "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComparePassword(tt.hash, tt.password)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHashPassword(t *testing.T) {
	password := "testPassword"

	// Test hashing password
	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Check if hash is a valid bcrypt hash
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	assert.NoError(t, err)
}

func TestGenerateID(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		length int
	}{
		{"Generate short ID", "input", 10},
		{"Generate long ID", "input", 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := GenerateID(tt.input, tt.length)
			assert.Len(t, id, tt.length)
		})
	}
}

func TestValidateStruct(t *testing.T) {
	// Define a simple struct to validate
	type User struct {
		Username string `validate:"required"`
		Email    string `validate:"required,email"`
	}

	validUser := User{Username: "validUser", Email: "user@example.com"}
	invalidUser := User{Username: "", Email: "invalidEmail"}

	tests := []struct {
		name     string
		input    interface{}
		expected error
	}{
		{"Valid struct", validUser, nil},
		{"Invalid struct", invalidUser, validator.New().Struct(&invalidUser)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)
			if tt.expected == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestApplyLuhnAlgorithm(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"Valid number", 12345, "123455"},
		{"Another valid number", 987654321, "9876543217"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyLuhnAlgorithm(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateLuhnAlgorithm(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid Luhn number", "2352352351", true},
		{"Invalid Luhn number", "2352352355", false},
		{"Another valid Luhn number", "6546546547", true},
		{"Invalid Luhn number", "9876543220", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateLuhnAlgorithm(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
