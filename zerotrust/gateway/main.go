package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Initialize Cosmos Chain connection.
	initData()

	// 2. Initialize Gin.
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// 3. Health check.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "zero-trust-gateway",
			"status":  "healthy",
		})
	})

	// 4. Authentication - Challenge.
	router.POST("/auth/challenge", handleChallenge)

	// 5. Access control - Authentication + Authorization + PEP.
	router.POST("/access", handleAccess)

	// 6. Audit records.
	router.GET("/audit", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"records": getAuditRecords(),
		})
	})

	// 7. Start HTTP server.
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Gateway running on port %s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start Gateway: ", err)
	}
}
