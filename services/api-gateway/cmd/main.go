package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"platform/gateway/internal/auth"
	"platform/gateway/internal/limiter"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var (
	authServiceURL     string
	functionServiceURL string
)

func main() {
	// Read service URLs from environment variables
	authServiceURL = os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://auth:8083" // Default Auth Service URL
	}

	functionServiceURL = os.Getenv("FUNCTION_SERVICE_URL")
	if functionServiceURL == "" {
		functionServiceURL = "http://functionservice:8080" // Default Search Service URL
	}

	// Initialize Gin router
	r := gin.Default()

	// Initialize Redis client for rate-limiting and concurrency control
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	// Apply rate limiting middleware, handle error
	limitMW, err := limiter.NewRateLimitMiddleware(redisClient, "token-bucket", 10, "1m")
	if err != nil {
		log.Fatalf("Error initializing rate limit middleware: %v", err)
	}
	r.Use(limitMW) // Apply the middleware if no error occurred

	r.Use(concurrencyMiddleware(redisClient))

	// Public routes (authentication)
	public := r.Group("/auth")
	{
		public.POST("/register", forwardToAuthService)
		public.POST("/login", forwardToAuthService)   // Handles OAuth2 + JWT generation
		public.POST("/refresh", forwardToAuthService) // Refresh JWT
	}

	// Protected routes (requires JWT validation)
	protected := r.Group("/")
	protected.Use(auth.AuthMiddleware()) // Apply JWT validation middleware
	{
		protected.GET("/health", func(c *gin.Context) {
			c.String(http.StatusOK, "API Gateway is healthy\n")
		})

		// Forward protected routes to the Search Service
		protected.Any("/functions", forwardToFunctionService)
		protected.Any("/functions/*rest", forwardToFunctionService)

		protected.Any("/jobs", forwardToFunctionService)
		protected.Any("/jobs/*rest", forwardToFunctionService)

		// Admin routes with additional role-based access control
		admin := protected.Group("/admin")
		admin.Use(auth.RequireRoles("admin"))
		{
			admin.GET("/dashboard", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"dashboard": "admin metrics"})
			})
		}
	}

	// Set the port for the API Gateway to run on
	addr := ":8082"
	if port := os.Getenv("GATEWAY_PORT"); port != "" {
		addr = ":" + port
	}

	log.Printf("Starting gateway on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start gateway:", err)
	}
}

// Middleware to handle concurrency control using Redis
func concurrencyMiddleware(rdb *redis.Client) gin.HandlerFunc {
	// Initialize rate limiting middleware and check for error
	limitMW, err := limiter.NewRateLimitMiddleware(rdb, "token-bucket", 10, "1m")
	if err != nil {
		log.Fatalf("Error initializing rate limit middleware: %v", err) // Log and terminate if there is an error
	}
	return limitMW // Return the middleware if no error
}

// Forward incoming requests to the Authentication Service
func forwardToAuthService(c *gin.Context) {
	path := strings.TrimPrefix(c.Request.URL.Path, "/auth")
	targetURL := fmt.Sprintf("%s%s", authServiceURL, path)
	log.Printf("Forwarding to Auth Service => %s", targetURL)

	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	for k, v := range c.Request.Header {
		req.Header[k] = v
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "auth service unreachable"})
		return
	}
	defer resp.Body.Close()

	c.Status(resp.StatusCode)
	for k, v := range resp.Header {
		c.Writer.Header()[k] = v
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	c.Writer.Write(bodyBytes)
}

// Forward incoming requests to the Search Service (functions, jobs)
func forwardToFunctionService(c *gin.Context) {
	finalURL := functionServiceURL + c.Request.URL.Path

	if c.Request.URL.RawQuery != "" {
		finalURL += "?" + c.Request.URL.RawQuery
	}

	log.Printf("Forwarding to Function Service => %s", finalURL)

	req, err := http.NewRequest(c.Request.Method, finalURL, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	for k, v := range c.Request.Header {
		req.Header[k] = v
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "function service unreachable"})
		return
	}
	defer resp.Body.Close()

	c.Status(resp.StatusCode)
	for k, v := range resp.Header {
		c.Writer.Header()[k] = v
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	c.Writer.Write(bodyBytes)
}
