package auth

import (
	"log"
	"net/http"
	"strings"

	"github.com/baralga/shared"
	"github.com/baralga/user"
)

// MCPAuthService handles MCP-specific authentication
type MCPAuthService struct {
	userRepository user.UserRepository
}

// NewMCPAuthService creates a new MCP authentication service
func NewMCPAuthService(userRepository user.UserRepository) *MCPAuthService {
	return &MCPAuthService{
		userRepository: userRepository,
	}
}

// AuthenticationMiddleware validates API key and creates principal context for MCP requests
func (m *MCPAuthService) AuthenticationMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for OPTIONS requests and initial MCP connection requests
			if r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			// For MCP streamable transport, we might need to allow some requests without API key initially
			// The MCP SDK handles session management internally
			apiKey := m.extractAPIKey(r)
			if apiKey == "" {
				// Log the request for debugging
				log.Printf("[MCP Auth] No API key provided for %s %s", r.Method, r.URL.Path)
				// For now, let's allow requests without API key and let the MCP handler deal with it
				// TODO: Implement proper MCP session-based authentication
				next.ServeHTTP(w, r)
				return
			}

			// Validate email format (basic validation)
			if !m.isValidEmail(apiKey) {
				m.renderMCPError(w, -32602, "Invalid API key format", "API key must be a valid email address")
				return
			}

			// Lookup user by email
			user, err := m.userRepository.FindUserByUsername(r.Context(), apiKey)
			if err != nil {
				log.Printf("User lookup failed for email %s: %v", apiKey, err)
				m.renderMCPError(w, -32603, "Authentication failed", "Invalid API key or user not found")
				return
			}

			// Fetch user roles
			roles, err := m.userRepository.FindRolesByUserID(r.Context(), user.OrganizationID, user.ID)
			if err != nil {
				log.Printf("Role lookup failed for user %s: %v", user.Username, err)
				m.renderMCPError(w, -32603, "Authentication failed", "Failed to retrieve user permissions")
				return
			}

			// Create principal context using existing mapping function
			principal := mapUserToPrincipal(user, roles)

			// Add principal to context
			ctx := shared.ToContextWithPrincipal(r.Context(), principal)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// extractAPIKey extracts API key from X-API-Key header or Authorization Bearer token
func (m *MCPAuthService) extractAPIKey(r *http.Request) string {
	// Try X-API-Key header first
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return apiKey
	}

	// Try Authorization Bearer token
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

// isValidEmail performs basic email validation
func (m *MCPAuthService) isValidEmail(email string) bool {
	if email == "" {
		return false
	}

	// Must contain exactly one @ symbol
	atCount := strings.Count(email, "@")
	if atCount != 1 {
		return false
	}

	// Split by @ to get local and domain parts
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	local, domain := parts[0], parts[1]

	// Local part cannot be empty
	if local == "" {
		return false
	}

	// Domain part must contain at least one dot and cannot be empty
	if domain == "" || !strings.Contains(domain, ".") {
		return false
	}

	// Domain cannot start or end with a dot
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}

	return true
}

// renderMCPError renders an MCP-compliant error response
func (m *MCPAuthService) renderMCPError(w http.ResponseWriter, code int, message, details string) {
	shared.RenderMCPError(w, code, message, details)
}
