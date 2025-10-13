package shared

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pkg/errors"
)

// MCPServer wraps the MCP server functionality
type MCPServer struct {
	server    *mcp.Server
	validator *validator.Validate
}

// MCPHandler interface for MCP tool handlers
type MCPHandler interface {
	RegisterMCPTools(server *mcp.Server)
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer() *MCPServer {
	impl := &mcp.Implementation{
		Name:    "baralga-time-tracker",
		Version: "1.0.0",
	}
	server := mcp.NewServer(impl, nil)

	return &MCPServer{
		server:    server,
		validator: validator.New(),
	}
}

// RegisterMCPRoutes registers MCP endpoints with the Chi router
func (m *MCPServer) RegisterMCPRoutes(router chi.Router, authMiddleware func(http.Handler) http.Handler, mcpHandlers []MCPHandler) {
	// Register all MCP tools from handlers
	for _, handler := range mcpHandlers {
		handler.RegisterMCPTools(m.server)
	}

	// Mount MCP endpoints under /mcp path
	router.Route("/mcp", func(r chi.Router) {
		// Add CORS headers for web compatibility
		r.Use(m.corsMiddleware)

		// Add API key authentication middleware
		r.Use(authMiddleware)

		// Handle MCP protocol requests
		r.Post("/*", m.handleMCPRequest)
		r.Options("/*", m.handleOptions)
	})
}

// corsMiddleware adds CORS headers for web compatibility
func (m *MCPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		w.Header().Set("Access-Control-Max-Age", "86400")

		next.ServeHTTP(w, r)
	})
}

// handleMCPRequest handles incoming MCP protocol requests using StreamableHTTPHandler
func (m *MCPServer) handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	// Create a StreamableHTTPHandler that returns our server
	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return m.server
	}, nil)

	// Delegate to the MCP HTTP handler
	handler.ServeHTTP(w, r)
}

// handleOptions handles CORS preflight requests
func (m *MCPServer) handleOptions(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// renderMCPError renders an MCP-compliant error response
func (m *MCPServer) renderMCPError(w http.ResponseWriter, code int, message, details string) {
	RenderMCPError(w, code, message, details)
}

// RenderMCPError renders an MCP-compliant error response (public function)
func RenderMCPError(w http.ResponseWriter, code int, message, details string) {
	RenderMCPErrorWithType(w, code, message, details, MCPErrorTypeSystem)
}

// RenderMCPErrorWithType renders an MCP-compliant error response with structured error type
func RenderMCPErrorWithType(w http.ResponseWriter, code int, message, details string, errorType MCPErrorType) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	errorResponse := map[string]any{
		"jsonrpc": "2.0",
		"error": map[string]any{
			"code":    code,
			"message": message,
			"data": MCPErrorDetails{
				Type:    errorType,
				Details: details,
			},
		},
		"id": nil,
	}

	// Log the error response for debugging
	log.Printf("[MCP] Sending error response: code=%d, message=%s, type=%s, details=%s",
		code, message, errorType, details)

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Printf("[MCP] Failed to encode error response: %v", err)
	}
}

// RenderMCPErrorFromDomain renders an MCP error response from a domain error
func RenderMCPErrorFromDomain(w http.ResponseWriter, err error) {
	code, message, details := ConvertDomainErrorToMCP(err)
	errorType := determineErrorType(err)
	RenderMCPErrorWithType(w, code, message, details, errorType)
}

// determineErrorType determines the MCP error type from a domain error
func determineErrorType(err error) MCPErrorType {
	if err == nil {
		return MCPErrorTypeSystem
	}

	errMsg := err.Error()

	// Check validation errors first
	if _, ok := err.(validator.ValidationErrors); ok {
		return MCPErrorTypeValidation
	}

	// Check for authentication errors
	if strings.Contains(errMsg, "authentication") || strings.Contains(errMsg, "api key") {
		return MCPErrorTypeAuthentication
	}

	// Check for authorization errors
	if strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "access denied") {
		return MCPErrorTypeAuthorization
	}

	// Check for not found errors
	if isNotFoundError(err) {
		return MCPErrorTypeNotFound
	}

	// Check for business logic errors
	if isBusinessLogicError(errMsg) {
		return MCPErrorTypeBusinessLogic
	}

	// Default to system error
	return MCPErrorTypeSystem
}

// MCPErrorType represents different categories of MCP errors
type MCPErrorType string

const (
	MCPErrorTypeAuthentication MCPErrorType = "authentication_error"
	MCPErrorTypeAuthorization  MCPErrorType = "authorization_error"
	MCPErrorTypeValidation     MCPErrorType = "validation_error"
	MCPErrorTypeNotFound       MCPErrorType = "not_found_error"
	MCPErrorTypeBusinessLogic  MCPErrorType = "business_logic_error"
	MCPErrorTypeSystem         MCPErrorType = "system_error"
)

// MCPErrorDetails contains structured error information for MCP responses
type MCPErrorDetails struct {
	Type    MCPErrorType `json:"type"`
	Details string       `json:"details"`
	Field   string       `json:"field,omitempty"`
}

// ConvertDomainErrorToMCP converts domain errors to MCP error responses with logging
func (m *MCPServer) ConvertDomainErrorToMCP(err error) (int, string, string) {
	return ConvertDomainErrorToMCP(err)
}

// ConvertDomainErrorToMCP converts domain errors to MCP error responses with logging (public function)
func ConvertDomainErrorToMCP(err error) (int, string, string) {
	if err == nil {
		log.Printf("[MCP] Unexpected nil error")
		return -32603, "Internal error", "Unknown error occurred"
	}

	// Log the original error for debugging
	log.Printf("[MCP] Converting domain error: %v", err)

	// Handle validation errors from go-playground/validator
	if validationErr, ok := err.(validator.ValidationErrors); ok {
		details := formatValidationErrors(validationErr)
		log.Printf("[MCP] Validation error: %s", details)
		return -32602, "Invalid params", details
	}

	// Handle specific domain errors by checking error types
	errMsg := err.Error()

	// Handle authentication errors (missing or invalid API key)
	if strings.Contains(errMsg, "authentication") || strings.Contains(errMsg, "invalid api key") || strings.Contains(errMsg, "missing api key") {
		log.Printf("[MCP] Authentication error: %s", errMsg)
		return -32600, "Invalid Request", "Authentication required"
	}

	// Handle authorization errors (user not found or insufficient permissions)
	if strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "access denied") || strings.Contains(errMsg, "permission denied") {
		log.Printf("[MCP] Authorization error: %s", errMsg)
		return -32603, "Internal error", "Access denied"
	}

	// Handle not found errors (check for specific domain errors and patterns)
	if isNotFoundError(err) {
		log.Printf("[MCP] Not found error: %s", errMsg)
		return -32602, "Invalid params", "Requested resource not found"
	}

	// Handle business logic errors (domain rule violations)
	if isBusinessLogicError(errMsg) {
		log.Printf("[MCP] Business logic error: %s", errMsg)
		return -32602, "Invalid params", errMsg
	}

	// Handle database constraint violations
	if isDatabaseConstraintError(err) {
		log.Printf("[MCP] Database constraint error: %s", errMsg)
		return -32602, "Invalid params", "Data constraint violation"
	}

	// Handle system errors (database connection issues, etc.)
	if isSystemError(err) {
		log.Printf("[MCP] System error: %s", errMsg)
		return -32603, "Internal error", "A system error occurred"
	}

	// Default to internal error for unknown errors
	log.Printf("[MCP] Unknown error type: %s", errMsg)
	return -32603, "Internal error", "An internal error occurred"
}

// formatValidationErrors formats validator.ValidationErrors into a readable string
func formatValidationErrors(validationErr validator.ValidationErrors) string {
	var errorMessages []string
	for _, err := range validationErr {
		switch err.Tag() {
		case "required":
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is required", err.Field()))
		case "email":
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' must be a valid email address", err.Field()))
		case "min":
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' must be at least %s characters", err.Field(), err.Param()))
		case "max":
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' must be at most %s characters", err.Field(), err.Param()))
		case "gte":
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' must be greater than or equal to %s", err.Field(), err.Param()))
		case "lte":
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' must be less than or equal to %s", err.Field(), err.Param()))
		default:
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' validation failed: %s", err.Field(), err.Tag()))
		}
	}
	return strings.Join(errorMessages, "; ")
}

// isNotFoundError checks if the error is a domain "not found" error
func isNotFoundError(err error) bool {
	// Import tracking domain errors - we'll need to import these when implementing
	// For now, check by error message patterns
	errMsg := err.Error()
	return strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "does not exist") ||
		strings.Contains(errMsg, "activity not found") ||
		strings.Contains(errMsg, "user not found") ||
		strings.Contains(errMsg, "project not found")
}

// isBusinessLogicError checks if the error is a business logic violation
func isBusinessLogicError(errMsg string) bool {
	return strings.Contains(errMsg, "end time must be after start time") ||
		strings.Contains(errMsg, "invalid time range") ||
		strings.Contains(errMsg, "duration must be positive") ||
		strings.Contains(errMsg, "invalid date") ||
		strings.Contains(errMsg, "business rule")
}

// isDatabaseConstraintError checks if the error is a database constraint violation
func isDatabaseConstraintError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "constraint") ||
		strings.Contains(errMsg, "duplicate") ||
		strings.Contains(errMsg, "unique violation") ||
		strings.Contains(errMsg, "foreign key") ||
		strings.Contains(errMsg, "check constraint")
}

// isSystemError checks if the error is a system-level error
func isSystemError(err error) bool {
	// Check for database connection errors
	if errors.Is(err, pgx.ErrNoRows) {
		return false // This is handled as not found, not system error
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "network") ||
		strings.Contains(errMsg, "database") ||
		strings.Contains(errMsg, "pool") ||
		strings.Contains(errMsg, "context deadline exceeded")
}

// ValidateToolParams validates MCP tool parameters using struct tags
func (m *MCPServer) ValidateToolParams(params interface{}) error {
	return m.validator.Struct(params)
}

// GetServer returns the underlying MCP server for tool registration
func (m *MCPServer) GetServer() *mcp.Server {
	return m.server
}

// Utility functions for common MCP error scenarios

// RenderMCPAuthenticationError renders an authentication error response
func RenderMCPAuthenticationError(w http.ResponseWriter, details string) {
	log.Printf("[MCP] Authentication error: %s", details)
	RenderMCPErrorWithType(w, -32600, "Invalid Request", details, MCPErrorTypeAuthentication)
}

// RenderMCPAuthorizationError renders an authorization error response
func RenderMCPAuthorizationError(w http.ResponseWriter, details string) {
	log.Printf("[MCP] Authorization error: %s", details)
	RenderMCPErrorWithType(w, -32603, "Internal error", details, MCPErrorTypeAuthorization)
}

// RenderMCPValidationError renders a validation error response
func RenderMCPValidationError(w http.ResponseWriter, details string) {
	log.Printf("[MCP] Validation error: %s", details)
	RenderMCPErrorWithType(w, -32602, "Invalid params", details, MCPErrorTypeValidation)
}

// RenderMCPNotFoundError renders a not found error response
func RenderMCPNotFoundError(w http.ResponseWriter, resource string) {
	details := fmt.Sprintf("%s not found", resource)
	log.Printf("[MCP] Not found error: %s", details)
	RenderMCPErrorWithType(w, -32602, "Invalid params", details, MCPErrorTypeNotFound)
}

// RenderMCPBusinessLogicError renders a business logic error response
func RenderMCPBusinessLogicError(w http.ResponseWriter, details string) {
	log.Printf("[MCP] Business logic error: %s", details)
	RenderMCPErrorWithType(w, -32602, "Invalid params", details, MCPErrorTypeBusinessLogic)
}

// RenderMCPSystemError renders a system error response
func RenderMCPSystemError(w http.ResponseWriter, details string) {
	log.Printf("[MCP] System error: %s", details)
	RenderMCPErrorWithType(w, -32603, "Internal error", details, MCPErrorTypeSystem)
}

// LogMCPError logs MCP-specific errors with structured information
func LogMCPError(operation string, err error, context map[string]any) {
	contextStr := ""
	if context != nil {
		if contextBytes, marshalErr := json.Marshal(context); marshalErr == nil {
			contextStr = string(contextBytes)
		}
	}

	log.Printf("[MCP] Error in %s: %v | Context: %s", operation, err, contextStr)
}

// LogMCPToolCall logs MCP tool calls for debugging and monitoring
func LogMCPToolCall(toolName string, params map[string]any, success bool) {
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}

	paramsStr := ""
	if params != nil {
		if paramsBytes, err := json.Marshal(params); err == nil {
			paramsStr = string(paramsBytes)
		}
	}

	log.Printf("[MCP] Tool call %s: %s | Params: %s", toolName, status, paramsStr)
}
