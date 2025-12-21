package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost follows OWASP recommendations with minimum cost of 12 for 2025
	BcryptCost = 12

	// Magic link token expiry time
	MagicLinkExpiry = 15 * time.Minute // 15 minutes for magic link authentication
)

// HashPassword hashes a password using bcrypt with high cost factor
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// HashEmail creates a SHA256 hash of email + secret key for privacy
func HashEmail(email, secretKey string) string {
	hasher := sha256.New()
	hasher.Write([]byte(email + secretKey))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken() string {
	return uuid.New().String()
}

// GenerateMagicLinkToken creates an HMAC-signed token for magic link authentication
func GenerateMagicLinkToken(emailHash string, hmacKey []byte) (token string, expiry time.Time) {
	expiry = time.Now().Add(MagicLinkExpiry)
	expiryUnix := expiry.Unix()

	// Create message: emailHash|timestamp
	message := fmt.Sprintf("%s|%d", emailHash, expiryUnix)

	// Sign with HMAC
	h := hmac.New(sha256.New, hmacKey)
	h.Write([]byte(message))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	// Token format: base64(emailHash|timestamp|signature)
	token = base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s|%s", message, signature)))
	return token, expiry
}

// VerifyMagicLinkToken validates an HMAC-signed magic link token
func VerifyMagicLinkToken(token string, hmacKey []byte) (emailHash string, valid bool) {
	// Decode token
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return "", false
	}

	parts := strings.Split(string(decoded), "|")
	if len(parts) != 3 {
		return "", false
	}

	emailHash = parts[0]
	timestampStr := parts[1]
	receivedSig := parts[2]

	// Parse timestamp
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return "", false
	}

	// Check expiry
	if time.Now().After(time.Unix(timestamp, 0)) {
		return "", false
	}

	// Verify HMAC
	message := fmt.Sprintf("%s|%s", emailHash, timestampStr)
	h := hmac.New(sha256.New, hmacKey)
	h.Write([]byte(message))
	expectedSig := base64.URLEncoding.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(receivedSig), []byte(expectedSig)) {
		return "", false
	}

	return emailHash, true
}