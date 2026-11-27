
package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"log"

	"github.com/gin-gonic/gin"
)

const resourceURL = "http://localhost:8081"

// handleAccess handles authenticated and authorized requests.
func handleAccess(c *gin.Context) {
	var req AccessRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	// 1. Validate HTTP method and resource path.
	if req.Method != http.MethodGet ||
		req.Action != "read" ||
		!validResourcePath(req.Path, req.Resource) {

		c.JSON(http.StatusForbidden, gin.H{
			"decision": "DENY",
			"reason":   "unsupported method, action or path",
		})
		return
	}

	// 2. Authentication.
	if err := verifyAccess(req); err != nil {
		recordAudit(
			req.DID,
			req.Resource,
			req.Action,
			"DENY",
			err.Error(),
			req,
		)

		c.JSON(http.StatusUnauthorized, gin.H{
			"decision": "DENY",
			"reason":   err.Error(),
		})
		return
	}

	// 3. Authorization - PDP.
	allowed, reason := checkAccess(
		req.DID,
		req.Resource,
		req.Action,
	)

	decision := "DENY"
	if allowed {
		decision = "ALLOW"
	}

	// 4. Audit.
	recordAudit(
		req.DID,
		req.Resource,
		req.Action,
		decision,
		reason,
		req,
	)

	// 5. PEP enforcement.
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{
			"decision": "DENY",
			"reason":   reason,
		})
		return
	}

	// 6. Forward request to Protected Resource.
	proxyToResource(c, req)
}

// validResourcePath validates the requested resource path.
func validResourcePath(path, resource string) bool {
	if path == "/documents" {
		return resource == "documents"
	}

	if strings.HasPrefix(path, "/documents/") {
		documentID := strings.TrimPrefix(path, "/documents/")

		if documentID == "" || strings.Contains(documentID, "/") {
			return false
		}

		return documentID == resource
	}

	return false
}

// proxyToResource forwards an authorized request.
func proxyToResource(c *gin.Context, req AccessRequest) {
	target, err := url.Parse(resourceURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid resource URL",
		})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Director = func(r *http.Request) {
		r.Method = req.Method
		r.URL.Scheme = target.Scheme
		r.URL.Host = target.Host
		r.URL.Path = req.Path
		r.URL.RawQuery = ""

		r.Host = target.Host

		// Request body đã được Gateway đọc bằng ShouldBindJSON.
		// Resource chỉ cần nhận GET request.
		r.Body = http.NoBody
		r.ContentLength = 0
		r.Header.Del("Content-Length")
		r.Header.Del("Transfer-Encoding")

		// Không tin header DID do Client tự gửi.
		r.Header.Del("X-Authenticated-DID")

		// DID được Gateway xác minh.
		r.Header.Set("X-Authenticated-DID", req.DID)
	}

	proxy.ErrorHandler = func(
		w http.ResponseWriter,
		r *http.Request,
		err error,
	) {
		log.Printf("Resource proxy error: %v", err)

		http.Error(
			w,
			"Protected Resource unavailable",
			http.StatusBadGateway,
		)
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
