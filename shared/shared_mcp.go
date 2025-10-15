package shared

import (
	"context"
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
	server       *mcp.Server
	validator    *validator.Validate
	handlers     []MCPHandler               // Store handlers for stateless access
	tools        []*mcp.Tool                // Store registered tools for dynamic listing
	toolHandlers map[string]ToolHandlerFunc // Map tool names to their handler functions
}

// ToolRegistrar interface for registering MCP tools
type ToolRegistrar interface {
	AddTool(tool *mcp.Tool, handler any)
}

// ToolHandlerFunc represents a function that handles a specific MCP tool
type ToolHandlerFunc func(ctx context.Context, req *mcp.CallToolRequest, arguments map[string]interface{}) (*mcp.CallToolResult, interface{}, error)

// MCPHandler interface for MCP tool handlers
type MCPHandler interface {
	RegisterMCPTools(registrar ToolRegistrar)
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer() *MCPServer {
	impl := &mcp.Implementation{
		Name:    "baralga-time-tracker",
		Version: "1.0.0",
	}
	server := mcp.NewServer(impl, nil)

	return &MCPServer{
		server:       server,
		validator:    validator.New(),
		toolHandlers: make(map[string]ToolHandlerFunc),
	}
}

// registerToolsFromHandlers registers tools from all MCP handlers and stores them
func (m *MCPServer) registerToolsFromHandlers(mcpHandlers []MCPHandler) {
	// Clear existing tools
	m.tools = nil

	// Register tools from each handler using our intercepting server
	for _, handler := range mcpHandlers {
		handler.RegisterMCPTools(m)
	}
}

// RegisterMCPRoutes registers MCP endpoints with the Chi router
func (m *MCPServer) RegisterMCPRoutes(router chi.Router, authMiddleware func(http.Handler) http.Handler, mcpHandlers []MCPHandler) {
	// Store handlers for stateless access
	m.handlers = mcpHandlers

	// Register tools from all handlers and store them
	m.registerToolsFromHandlers(mcpHandlers)

	log.Printf("[MCP] Using STATELESS mode with %d registered tools", len(m.tools))

	// Mount MCP endpoints under /mcp path
	router.Route("/mcp", func(r chi.Router) {
		// Add CORS headers for web compatibility
		r.Use(m.corsMiddleware)

		// Add API key authentication middleware
		r.Use(authMiddleware)

		// Handle MCP protocol requests (stateless only)
		r.HandleFunc("/", m.handleMCPRequest)
		r.HandleFunc("/*", m.handleMCPRequest)
		r.Options("/", m.handleOptions)
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

// handleMCPRequest handles incoming MCP protocol requests in stateless mode
func (m *MCPServer) handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	// Handle GET requests for server availability check
	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"name":    "baralga-time-tracker",
			"version": "1.0.0",
			"status":  "ready",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Only handle POST requests for JSON-RPC
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON-RPC request
	var jsonRPCReq map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&jsonRPCReq); err != nil {
		m.renderMCPError(w, -32700, "Parse error", "Invalid JSON")
		return
	}

	// Handle the request statelessly
	m.handleStatelessJSONRPC(r.Context(), w, jsonRPCReq)
}

// handleStatelessJSONRPC handles JSON-RPC requests without maintaining session state
func (m *MCPServer) handleStatelessJSONRPC(ctx context.Context, w http.ResponseWriter, req map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")

	method, _ := req["method"].(string)
	id := req["id"]
	params, _ := req["params"].(map[string]interface{})

	switch method {
	case "initialize":
		m.handleStatelessInitialize(w, id)
	case "tools/list":
		m.handleStatelessToolsList(w, id)
	case "tools/call":
		m.handleStatelessToolsCall(ctx, w, id, params)
	default:
		m.renderJSONRPCError(w, id, -32601, "Method not found", fmt.Sprintf("Unknown method: %s", method))
	}
}

// handleStatelessInitialize handles initialize requests without session state
func (m *MCPServer) handleStatelessInitialize(w http.ResponseWriter, id interface{}) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": false,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "baralga-time-tracker",
				"version": "1.0.0",
			},
		},
	}
	json.NewEncoder(w).Encode(response)
}

// handleStatelessToolsList handles tools/list requests without session state
func (m *MCPServer) handleStatelessToolsList(w http.ResponseWriter, id interface{}) {
	// Generate tools list dynamically from registered tools
	tools := make([]map[string]interface{}, len(m.tools))
	for i, tool := range m.tools {
		toolMap := map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
		}

		// Add input schema if available
		if tool.InputSchema != nil {
			toolMap["inputSchema"] = tool.InputSchema
		} else {
			// Provide a basic schema if none is specified
			toolMap["inputSchema"] = map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			}
		}

		tools[i] = toolMap
	}

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"tools": tools,
		},
	}
	json.NewEncoder(w).Encode(response)
}

// handleStatelessToolsCall handles tools/call requests without session state
func (m *MCPServer) handleStatelessToolsCall(ctx context.Context, w http.ResponseWriter, id interface{}, params map[string]interface{}) {
	toolName, _ := params["name"].(string)
	arguments, _ := params["arguments"].(map[string]interface{})

	if toolName == "" {
		m.renderJSONRPCError(w, id, -32602, "Invalid params", "Tool name is required")
		return
	}

	// Find the handler for this tool
	handler, exists := m.toolHandlers[toolName]
	if !exists {
		m.renderJSONRPCError(w, id, -32601, "Method not found", fmt.Sprintf("Tool '%s' not found", toolName))
		return
	}

	// Create a mock MCP request for the tool call
	argumentsJSON, err := json.Marshal(arguments)
	if err != nil {
		m.renderJSONRPCError(w, id, -32602, "Invalid params", "Failed to marshal arguments")
		return
	}

	mockRequest := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      toolName,
			Arguments: argumentsJSON,
		},
	}

	// Call the tool handler using reflection or type assertion
	result, callErr := m.callToolHandler(ctx, mockRequest, toolName, arguments, handler)

	if callErr != nil {
		log.Printf("[MCP] Tool call failed: %v", callErr)
		m.renderJSONRPCError(w, id, -32603, "Internal error", callErr.Error())
		return
	}

	if result == nil {
		m.renderJSONRPCError(w, id, -32603, "Internal error", "Tool handler returned nil result")
		return
	}

	// Convert the MCP result to JSON-RPC response
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}

	json.NewEncoder(w).Encode(response)
}

// renderJSONRPCError renders a JSON-RPC error response
func (m *MCPServer) renderJSONRPCError(w http.ResponseWriter, id interface{}, code int, message, details string) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
			"data":    details,
		},
	}
	json.NewEncoder(w).Encode(response)
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

// AddTool adds a tool to the server and stores it for dynamic listing
func (m *MCPServer) AddTool(tool *mcp.Tool, handler any) {
	// Store the tool for dynamic listing
	m.tools = append(m.tools, tool)

	// Store the handler function for this tool
	if handlerFunc, ok := handler.(ToolHandlerFunc); ok {
		m.toolHandlers[tool.Name] = handlerFunc
	} else {
		log.Printf("[MCP] Warning: Handler for tool '%s' is not a ToolHandlerFunc", tool.Name)
	}

	// Note: We don't register with the underlying server in stateless mode
	// The tools are handled directly in handleStatelessToolsCall
}

// GetTools returns the registered tools for testing
func (m *MCPServer) GetTools() []*mcp.Tool {
	return m.tools
}

// RegisterToolsFromHandlers is a public method for testing
func (m *MCPServer) RegisterToolsFromHandlers(mcpHandlers []MCPHandler) {
	m.registerToolsFromHandlers(mcpHandlers)
}

// callToolHandler calls the appropriate tool handler function
func (m *MCPServer) callToolHandler(ctx context.Context, req *mcp.CallToolRequest, toolName string, arguments map[string]interface{}, handler ToolHandlerFunc) (*mcp.CallToolResult, error) {
	// Call the tool handler function directly
	result, _, err := handler(ctx, req, arguments)
	return result, err
}

// parseArguments converts map[string]interface{} to typed parameters
func (m *MCPServer) parseArguments(arguments map[string]interface{}, target interface{}) error {
	// Convert arguments map to JSON and then unmarshal to target struct
	jsonBytes, err := json.Marshal(arguments)
	if err != nil {
		return errors.Wrap(err, "failed to marshal arguments")
	}

	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return errors.Wrap(err, "failed to unmarshal arguments to target type")
	}

	// Validate the parsed parameters
	if err := m.validator.Struct(target); err != nil {
		return errors.Wrap(err, "parameter validation failed")
	}

	return nil
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
