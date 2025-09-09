package services

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"lms-manager/config"
	"lms-manager/utils"

	"github.com/gin-gonic/gin"
)

// SecurityService handles security features
type SecurityService struct {
	config      config.SecurityConfig
	rateLimiter *RateLimiter
	blacklist   map[string]time.Time
	whitelist   map[string]bool
	mu          sync.RWMutex
}

// RateLimiter handles rate limiting
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
}

// NewSecurityService creates a new security service
func NewSecurityService(cfg config.SecurityConfig) *SecurityService {
	return &SecurityService{
		config:      cfg,
		rateLimiter: NewRateLimiter(cfg.RateLimit),
		blacklist:   make(map[string]time.Time),
		whitelist:   make(map[string]bool),
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxRequests int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
	}
}

// RateLimitMiddleware returns a rate limiting middleware
func (s *SecurityService) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := s.getClientIP(c)
		
		// Check if IP is blacklisted
		if s.isBlacklisted(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "IP address is blacklisted",
			})
			c.Abort()
			return
		}

		// Check rate limit
		if !s.rateLimiter.Allow(clientIP) {
			s.addToBlacklist(clientIP, 5*time.Minute)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORSMiddleware returns a CORS middleware
func (s *SecurityService) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// IPWhitelistMiddleware returns an IP whitelist middleware
func (s *SecurityService) IPWhitelistMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := s.getClientIP(c)
		
		if !s.isIPAllowed(clientIP) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied from this IP address",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware returns a security headers middleware
func (s *SecurityService) SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// InputValidationMiddleware returns an input validation middleware
func (s *SecurityService) InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate request method
		allowedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		methodAllowed := false
		for _, method := range allowedMethods {
			if c.Request.Method == method {
				methodAllowed = true
				break
			}
		}

		if !methodAllowed {
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"error": "Method not allowed",
			})
			c.Abort()
			return
		}

		// Validate content type for POST/PUT requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			contentType := c.GetHeader("Content-Type")
			if !strings.Contains(contentType, "application/json") && 
			   !strings.Contains(contentType, "application/x-www-form-urlencoded") &&
			   !strings.Contains(contentType, "multipart/form-data") {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid content type",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// LoggingMiddleware returns a security logging middleware
func (s *SecurityService) LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		clientIP := s.getClientIP(c)
		userAgent := c.GetHeader("User-Agent")

		// Process request
		c.Next()

		// Log request
		duration := time.Since(start)
		status := c.Writer.Status()

		// Log suspicious activity
		if status >= 400 {
			s.logSuspiciousActivity(clientIP, userAgent, c.Request.URL.Path, status)
		}

		// Log slow requests
		if duration > 5*time.Second {
			utils.Warn("Slow request: %s %s from %s took %v", 
				c.Request.Method, c.Request.URL.Path, clientIP, duration)
		}

		utils.Info("Request: %s %s from %s - %d - %v", 
			c.Request.Method, c.Request.URL.Path, clientIP, status, duration)
	}
}

// Helper methods

// getClientIP gets the client IP address
func (s *SecurityService) getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	xri := c.GetHeader("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fallback to remote address
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}

	return ip
}

// isIPAllowed checks if an IP is allowed
func (s *SecurityService) isIPAllowed(ip string) bool {
	// Check whitelist first
	if len(s.config.AllowedIPs) > 0 {
		for _, allowedIP := range s.config.AllowedIPs {
			if s.isIPInRange(ip, allowedIP) {
				return true
			}
		}
		return false
	}

	return true // Allow all if no whitelist configured
}

// isIPInRange checks if an IP is in a given range
func (s *SecurityService) isIPInRange(ip, cidr string) bool {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		// If not a CIDR, treat as single IP
		return ip == cidr
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	return network.Contains(parsedIP)
}

// isBlacklisted checks if an IP is blacklisted
func (s *SecurityService) isBlacklisted(ip string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	expiry, exists := s.blacklist[ip]
	if !exists {
		return false
	}

	// Check if blacklist entry has expired
	if time.Now().After(expiry) {
		s.mu.RUnlock()
		s.mu.Lock()
		delete(s.blacklist, ip)
		s.mu.Unlock()
		s.mu.RLock()
		return false
	}

	return true
}

// addToBlacklist adds an IP to the blacklist
func (s *SecurityService) addToBlacklist(ip string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.blacklist[ip] = time.Now().Add(duration)
	utils.Warn("IP blacklisted: %s for %v", ip, duration)
}

// removeFromBlacklist removes an IP from the blacklist
func (s *SecurityService) removeFromBlacklist(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.blacklist, ip)
	utils.Info("IP removed from blacklist: %s", ip)
}

// logSuspiciousActivity logs suspicious activity
func (s *SecurityService) logSuspiciousActivity(ip, userAgent, path string, status int) {
	utils.Warn("Suspicious activity: IP=%s, UserAgent=%s, Path=%s, Status=%d", 
		ip, userAgent, path, status)
}

// Rate limiter methods

// Allow checks if a request is allowed for the given IP
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	window := 1 * time.Minute // 1 minute window

	// Clean old requests
	if requests, exists := rl.requests[ip]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) < window {
				validRequests = append(validRequests, reqTime)
			}
		}
		rl.requests[ip] = validRequests
	}

	// Check if under limit
	if len(rl.requests[ip]) >= 100 { // Max 100 requests per minute
		return false
	}

	// Add current request
	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}

// GetBlacklist returns the current blacklist
func (s *SecurityService) GetBlacklist() map[string]time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy
	blacklist := make(map[string]time.Time)
	for ip, expiry := range s.blacklist {
		blacklist[ip] = expiry
	}

	return blacklist
}

// GetWhitelist returns the current whitelist
func (s *SecurityService) GetWhitelist() []string {
	return s.config.AllowedIPs
}

// UpdateWhitelist updates the IP whitelist
func (s *SecurityService) UpdateWhitelist(ips []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config.AllowedIPs = ips
	utils.Info("IP whitelist updated: %v", ips)
}

// ClearBlacklist clears the IP blacklist
func (s *SecurityService) ClearBlacklist() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.blacklist = make(map[string]time.Time)
	utils.Info("IP blacklist cleared")
}

// GetSecurityStats returns security statistics
func (s *SecurityService) GetSecurityStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]interface{}{
		"blacklisted_ips": len(s.blacklist),
		"whitelisted_ips": len(s.config.AllowedIPs),
		"rate_limit":      s.config.RateLimit,
		"session_timeout": s.config.SessionTimeout,
	}

	return stats
}
