package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

type AuditRecord struct {
	Subject     string    `json:"subject"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Decision    string    `json:"decision"`
	Reason      string    `json:"reason"`
	PayloadHash string    `json:"payload_hash"`
	Timestamp   time.Time `json:"timestamp"`
}

var (
	auditRecords []AuditRecord
	auditMutex   sync.Mutex
)

func recordAudit(
	subject string,
	resource string,
	action string,
	decision string,
	reason string,
	payload any,
) {
	rawPayload, _ := json.Marshal(payload)

	hash := sha256.Sum256(rawPayload)

	record := AuditRecord{
		Subject:     subject,
		Resource:    resource,
		Action:      action,
		Decision:    decision,
		Reason:      reason,
		PayloadHash: hex.EncodeToString(hash[:]),
		Timestamp:   time.Now().UTC(),
	}

	auditMutex.Lock()
	defer auditMutex.Unlock()

	auditRecords = append(auditRecords, record)
}

func getAuditRecords() []AuditRecord {
	auditMutex.Lock()
	defer auditMutex.Unlock()

	result := make([]AuditRecord, len(auditRecords))
	copy(result, auditRecords)

	return result
}
