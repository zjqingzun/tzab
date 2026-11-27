package main

type ChallengeRequest struct {
	DID string `json:"did"`
}

type ChallengeResponse struct {
	Nonce     string `json:"nonce"`
	ExpiresIn int    `json:"expires_in"`
}

type AccessRequest struct {
	DID       string `json:"did"`
	Nonce     string `json:"nonce"`
	Signature string `json:"signature"`
	Timestamp int64  `json:"timestamp"`
	Resource  string `json:"resource"`
	Action    string `json:"action"`
	Method    string `json:"method"`
	Path      string `json:"path"`
}
