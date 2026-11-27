
package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	keyFile = "private.key"
)

func main() {
	gateway := flag.String("gateway", "http://localhost:8080", "Gateway URL")
	did := flag.String("did", "did:example:user1", "Client DID")
	resource := flag.String("resource", "doc-001", "Resource ID")
	action := flag.String("action", "read", "Access action")
	method := flag.String("method", "GET", "HTTP method")
	path := flag.String("path", "/documents/doc-001", "Resource path")
	flag.Parse()

	// 1. Load or generate key pair.
	privateKey, publicKey, err := loadOrGenerateKey()
	if err != nil {
		fmt.Println("Key error:", err)
		return
	}

	fmt.Println("Client DID:", *did)
	fmt.Println("Public Key:", base64.StdEncoding.EncodeToString(publicKey))

	// 2. Request challenge.
	fmt.Println("\n[1] Requesting challenge...")

	challenge, err := requestChallenge(*gateway, *did)
	if err != nil {
		fmt.Println("Challenge error:", err)
		return
	}

	fmt.Println("Nonce received:", challenge.Nonce)

	// 3. Create timestamp and signed payload.
	timestamp := time.Now().Unix()

	payload := fmt.Sprintf(
		"%s|%s|%s|%s|%d|%s|%s",
		challenge.Nonce,
		*did,
		*resource,
		*action,
		timestamp,
		*method,
		*path,
	)

	// 4. Sign SHA-256 hash using Ed25519.
	hash := sha256.Sum256([]byte(payload))
	signature := ed25519.Sign(privateKey, hash[:])

	fmt.Println("[2] Request signed successfully.")

	// 5. Create access request.
	accessReq := AccessRequest{
		DID:       *did,
		Nonce:     challenge.Nonce,
		Signature: base64.StdEncoding.EncodeToString(signature),
		Timestamp: timestamp,
		Resource:  *resource,
		Action:    *action,
		Method:    *method,
		Path:      *path,
	}

	// 6. Send POST /access.
	fmt.Println("[3] Sending access request...")

	body, err := json.Marshal(accessReq)
	if err != nil {
		fmt.Println("JSON error:", err)
		return
	}

	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest(
		http.MethodPost,
		strings.TrimRight(*gateway, "/")+"/access",
		bytes.NewReader(body),
	)
	if err != nil {
		fmt.Println("Request error:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("Gateway connection error:", err)
		return
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Response error:", err)
		return
	}

	// 7. Display result.
	fmt.Println("\n===== ACCESS RESULT =====")
	fmt.Println("HTTP Status:", resp.Status)
	fmt.Println("Response:", string(responseBody))
}

// Load existing private key or generate a new key pair.

func loadOrGenerateKey() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	var privateKey ed25519.PrivateKey
	var publicKey ed25519.PublicKey

	data, err := os.ReadFile(keyFile)

	if err == nil {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
		if err != nil {
			return nil, nil, fmt.Errorf("invalid private key file: %w", err)
		}

		if len(decoded) != ed25519.PrivateKeySize {
			return nil, nil, fmt.Errorf("invalid private key length")
		}

		privateKey = ed25519.PrivateKey(decoded)
		publicKey = privateKey.Public().(ed25519.PublicKey)

	} else {
		if !os.IsNotExist(err) {
			return nil, nil, err
		}

		publicKey, privateKey, err = ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, err
		}

		encoded := base64.StdEncoding.EncodeToString(privateKey)

		if err := os.WriteFile(keyFile, []byte(encoded), 0600); err != nil {
			return nil, nil, err
		}

		fmt.Println("New Ed25519 key pair generated.")
		fmt.Println("Private key saved to:", keyFile)
	}

	// Save Public Key in both cases.
	encodedPublicKey := base64.StdEncoding.EncodeToString(publicKey)

	if err := os.WriteFile("public.key", []byte(encodedPublicKey), 0644); err != nil {
		return nil, nil, fmt.Errorf("public key save error: %w", err)
	}

	fmt.Println("Public Key:", encodedPublicKey)
	fmt.Println("Public Key saved to: public.key")

	return privateKey, publicKey, nil
}


// Request a nonce from Gateway.
func requestChallenge(gateway, did string) (*ChallengeResponse, error) {
	body, err := json.Marshal(ChallengeRequest{
		DID: did,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		strings.TrimRight(gateway, "/")+"/auth/challenge",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"HTTP %d: %s",
			resp.StatusCode,
			string(responseBody),
		)
	}

	var result ChallengeResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, err
	}

	if result.Nonce == "" {
		return nil, fmt.Errorf("empty nonce received")
	}

	return &result, nil
}
