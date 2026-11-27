
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	nonceTTL       = 30 * time.Second
	maxClockSkew   = 60
)

type ChallengeRequest struct {
	DID string `json:"did" binding:"required"`
}

type ChallengeResponse struct {
	Nonce     string `json:"nonce"`
	ExpiresAt int64  `json:"expires_at"`
}

type AccessRequest struct {
	DID       string `json:"did" binding:"required"`
	Nonce     string `json:"nonce" binding:"required"`
	Signature string `json:"signature" binding:"required"`
	Timestamp int64  `json:"timestamp" binding:"required"`

	Resource string `json:"resource" binding:"required"`
	Action   string `json:"action" binding:"required"`

	Method string `json:"method" binding:"required"`
	Path   string `json:"path" binding:"required"`
}

type NonceEntry struct {
	DID       string
	ExpiresAt time.Time
}

type NonceStore struct {
	mu     sync.Mutex
	values map[string]NonceEntry
}

var nonceStore = NonceStore{
	values: make(map[string]NonceEntry),
}

// Create generates and stores a one-time nonce.
func (s *NonceStore) Create(did string) (string, time.Time, error) {
	randomBytes := make([]byte, 32)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", time.Time{}, err
	}

	nonce := hex.EncodeToString(randomBytes)
	expiresAt := time.Now().Add(nonceTTL)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove expired nonces to prevent unlimited memory growth.
	now := time.Now()
	for key, entry := range s.values {
		if !now.Before(entry.ExpiresAt) {
			delete(s.values, key)
		}
	}

	s.values[nonce] = NonceEntry{
		DID:       did,
		ExpiresAt: expiresAt,
	}

	return nonce, expiresAt, nil
}

// Consume validates and deletes a nonce.
// A nonce can only be consumed once.
func (s *NonceStore) Consume(nonce, did string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.values[nonce]
	if !exists {
		return false
	}

	// Delete even if DID does not match.
	delete(s.values, nonce)

	return entry.DID == did && time.Now().Before(entry.ExpiresAt)
}

// signedPayload creates the message that the client must sign.
func signedPayload(req AccessRequest) []byte {
	payload := req.Nonce + "|" +
		req.DID + "|" +
		req.Resource + "|" +
		req.Action + "|" +
		stringInt(req.Timestamp) + "|" +
		req.Method + "|" +
		req.Path

	hash := sha256.Sum256([]byte(payload))
	return hash[:]
}

func stringInt(value int64) string {
	return strconv.FormatInt(value, 10)
}

// verifyAccess verifies identity, timestamp, signature and nonce.
func verifyAccess(req AccessRequest) error {
	identity, exists := getIdentity(req.DID)
	if !exists || identity.Status != "ACTIVE" {
		return errors.New("identity not found or inactive")
	}

	// Check request timestamp.
	now := time.Now().Unix()

	if req.Timestamp < now-maxClockSkew ||
		req.Timestamp > now+maxClockSkew {
		return errors.New("request timestamp expired or invalid")
	}

	// Decode signature.
	signature, err := base64.StdEncoding.DecodeString(req.Signature)
	if err != nil {
		return errors.New("invalid signature encoding")
	}

	if len(signature) != ed25519.SignatureSize {
		return errors.New("invalid signature length")
	}

	// Verify Ed25519 signature.
	if !ed25519.Verify(
		identity.PublicKey,
		signedPayload(req),
		signature,
	) {
		return errors.New("signature verification failed")
	}

	// Consume nonce only after signature verification.
	// The mutex ensures that concurrent replay attempts cannot both succeed.
	if !nonceStore.Consume(req.Nonce, req.DID) {
		return errors.New("invalid, expired or reused nonce")
	}

	return nil
}

// POST /auth/challenge
func handleChallenge(c *gin.Context) {
	var req ChallengeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": "invalid request",
		})
		return
	}

	identity, exists := getIdentity(req.DID)
	if !exists || identity.Status != "ACTIVE" {
		c.JSON(401, gin.H{
			"error": "identity not found or inactive",
		})
		return
	}

	nonce, expiresAt, err := nonceStore.Create(req.DID)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to create challenge",
		})
		return
	}

	c.JSON(200, ChallengeResponse{
		Nonce:     nonce,
		ExpiresAt: expiresAt.Unix(),
	})
}
