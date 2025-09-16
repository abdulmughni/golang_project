package models

const (
	UserId   = "requestUserIdClaim"
	TenantId = "requestTenantIdClaim"

	// System Roles
	ChatMessageRoleSystem    = "system"
	ChatMessageRoleUser      = "user"
	ChatMessageRoleAssistant = "assistant"
	ChatMessageRoleFunction  = "function"
	ChatMessageRoleTool      = "tool"

	// Available Models
	GPT4 = "gpt-4"
)

// Response errors constants
const (
	// Parameter errors
	UserIdError       = "User ID is required..."
	TenantIdError     = "Tenant ID is required..."
	ProjectIdError    = "Project ID is required..."
	ParameterRequired = "%v is required..."

	// Database errors
	DatabaseError    = "psql error: %v"
	TransactionError = "transaction error: %v"
	PrepareStatement = "prepare statement error: %v"
	ResponseSuccess  = "success"

	InternalServerError = "Internal Server Error"
)

// Response constants

const (
	StatusCreated = "Resource created successfully."
	StatusUpdated = "Resource updated successfully."
	StatusDeleted = "Resource deleted successfully."
	StatusSuccess = "Success"
)

// Test Color Codes
const (
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Red    = "\033[31m"
	Reset  = "\033[0m"
)
