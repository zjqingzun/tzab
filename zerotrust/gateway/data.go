package main

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	zttypes "zerotrust/x/zt/types"
)

const (
	defaultChainGRPCAddr = "localhost:9090"
	queryTimeout         = 5 * time.Second
)

type Identity struct {
	DID       string
	PublicKey ed25519.PublicKey
	Role      string
	Status    string
	CreatedAt int64
}

type Grant struct {
	Subject  string
	Resource string
	Action   string
	Status   string
}

var (
	queryClient zttypes.QueryClient
	grpcConn    *grpc.ClientConn
	initOnce    sync.Once
)

func initData() {
	initOnce.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Println("Warning: .env file not found, using system environment variables")
		}

		addr := os.Getenv("CHAIN_GRPC_ADDR")
		if addr == "" {
			addr = defaultChainGRPCAddr
		}

		conn, err := grpc.NewClient(
			addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Fatal("Failed to create Cosmos gRPC connection:", err)
		}

		grpcConn = conn
		queryClient = zttypes.NewQueryClient(conn)

		log.Println("Connected to Cosmos gRPC:", addr)
	})
}

// getIdentity truy vấn Identity trực tiếp từ Cosmos Chain.
func getIdentity(did string) (Identity, bool) {
	if queryClient == nil {
		log.Println("Identity query client is not initialized")
		return Identity{}, false
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		queryTimeout,
	)
	defer cancel()

	res, err := queryClient.GetIdentity(
		ctx,
		&zttypes.QueryGetIdentityRequest{
			Did: did,
		},
	)
	if err != nil {
		log.Printf("Failed to query Identity %s: %v", did, err)
		return Identity{}, false
	}

	chainIdentity := res.Identity

	if chainIdentity.Did == "" {
		return Identity{}, false
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(
		chainIdentity.Pubkey,
	)
	if err != nil {
		log.Printf("Invalid public key encoding for DID %s: %v", did, err)
		return Identity{}, false
	}

	if len(publicKeyBytes) != ed25519.PublicKeySize {
		log.Printf("Invalid Ed25519 public key length for DID %s", did)
		return Identity{}, false
	}

	identity := Identity{
		DID:       chainIdentity.Did,
		PublicKey: ed25519.PublicKey(publicKeyBytes),
		Role:      chainIdentity.Role,
		Status:    strings.ToUpper(chainIdentity.Status),
		CreatedAt: chainIdentity.CreatedAt,
	}

	return identity, true
}

// checkAccessOnChain queries access permission from Cosmos Blockchain.
func checkAccessOnChain(
	subject, resource, action string,
) (bool, string) {
	if queryClient == nil {
		log.Println("Blockchain query client is not initialized")
		return false, "blockchain query client unavailable"
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		queryTimeout,
	)
	defer cancel()

	res, err := queryClient.CheckAccess(
		ctx,
		&zttypes.QueryCheckAccessRequest{
			Subject:  subject,
			Resource: resource,
			Action:   action,
		},
	)

	if err != nil {
		log.Printf("Failed to query blockchain access: %v", err)
		return false, "blockchain access query failed"
	}

	if res == nil {
		return false, "empty blockchain response"
	}

	return res.Allowed, res.Reason
}
