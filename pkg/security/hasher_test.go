// ============================================
// FILE: pkg/security/hasher_test.go
// ============================================

package security

import (
	"crypto/sha256" // 确保导入 crypto/sha256
	"encoding/hex"  // 确保导入 encoding/hex
	"testing"

	"log/slog"
	"os"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestBcryptHasher_HashPassword(t *testing.T) {
	hasher := NewBcryptHasher(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	password := "mysecretpassword"

	hash, err := hasher.HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Verify the hash cost matches the default
	cost, err := bcrypt.Cost([]byte(hash))
	assert.NoError(t, err)
	assert.Equal(t, defaultPasswordCost, cost)

	// Hashing the same password again should produce a different hash (due to salt)
	hash2, err2 := hasher.HashPassword(password)
	assert.NoError(t, err2)
	assert.NotEqual(t, hash, hash2)
}

func TestBcryptHasher_CheckPasswordHash(t *testing.T) {
	hasher := NewBcryptHasher(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	password := "mysecretpassword"
	wrongPassword := "wrongpassword"

	hash, err := hasher.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Check correct password
	match := hasher.CheckPasswordHash(password, hash)
	assert.True(t, match)

	// Check incorrect password
	match = hasher.CheckPasswordHash(wrongPassword, hash)
	assert.False(t, match)

	// Check with invalid hash format (should return false and log warning)
	match = hasher.CheckPasswordHash(password, "invalid-hash-format")
	assert.False(t, match)
}

func TestSha256Hash(t *testing.T) {
	input := "this is a test string"
	// 使用标准库直接计算预期的哈希值，确保正确性
	h := sha256.New()
	h.Write([]byte(input))
	expectedHashBytes := h.Sum(nil)
	expectedHash := hex.EncodeToString(expectedHashBytes)
	// 正确的 SHA-256 哈希值应该是：c7be1ed902fb8dd4d48997c6452f5d7e509fbcdbe2808b16bcf4edce4c07d14e

	// 调用被测试的函数
	hash := Sha256Hash(input)
	// 断言生成的哈希值与预期值相等
	assert.Equal(t, expectedHash, hash, "Input: '%s'", input) // 添加消息以便调试

	// 测试空字符串
	emptyInput := ""
	hEmpty := sha256.New()
	hEmpty.Write([]byte(emptyInput))
	expectedEmptyHashBytes := hEmpty.Sum(nil)
	expectedEmptyHash := hex.EncodeToString(expectedEmptyHashBytes)

	emptyHash := Sha256Hash(emptyInput)
	assert.Equal(t, expectedEmptyHash, emptyHash, "Input: '%s'", emptyInput)

	// 测试再次哈希相同值得到相同结果 (确定性)
	hashAgain := Sha256Hash(input)
	assert.Equal(t, hash, hashAgain, "Hashing same input again should yield same hash")

	// 测试不同输入产生不同哈希
	differentInput := "this is another test string"
	differentHash := Sha256Hash(differentInput)
	assert.NotEqual(t, hash, differentHash, "Hashing different inputs should yield different hashes")
}
