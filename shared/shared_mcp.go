package shared

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	errorResponse := map[string]any{
		"jsonrpc": "2.0",
		"error": map[string]any{
			"code":    code,
			"message": message,
			"data": map[string]any{
				"type":    "mcp_error",
				"details": details,
			},
		},
		"id": nil,
	}

	json.NewEncoder(w).Encode(errorResponse)
}

// ConvertDomainErrorToMCP converts domain errors to MCP error responses
func (m *MCPServer) ConvertDomainErrorToMCP(err error) (int, string, string) {
	if err == nil {
		return -32603, "Internal error", "Unknown error occurred"
	}

	// Handle validation errors
	if validationErr, ok := err.(validator.ValidationErrors); ok {
		return -32602, "Invalid params", fmt.Sprintf("Validation failed: %s", validationErr.Error())
	}

	// Handle not found errors (check for common patterns)
	errMsg := err.Error()
	if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist") {
		return -32602, "Invalid params", "Requested resource not found"
	}

	// Handle authorization errors
	if strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "access denied") {
		return -32603, "Internal error", "Access denied"
	}

	// Handle constraint violations
	if strings.Contains(errMsg, "constraint") || strings.Contains(errMsg, "duplicate") {
		return -32602, "Invalid params", "Data constraint violation"
	}

	// Default to internal error
	return -32603, "Internal error", "An internal error occurred"
}

// ValidateToolParams validates MCP tool parameters using struct tags
func (m *MCPServer) ValidateToolParams(params interface{}) error {
	return m.validator.Struct(params)
}

// GetServer returns the underlying MCP server for tool registration
func (m *MCPServer) GetServer() *mcp.Server {
	return m.server
}
