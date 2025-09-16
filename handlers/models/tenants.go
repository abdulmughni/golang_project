package models

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices"
)

type CreateRequest struct {
	SubscriptionID     string `json:"subscription_id" binding:"required"`
	ResourceGroupName  string `json:"resource_group_name" binding:"required"`
	Location           string `json:"location" binding:"required"`
	OpenAIResourceName string `json:"openai_resource_name" binding:"required"`
}

type TenantOpenAiResource struct {
	SubscriptionID     string         `json:"subscription_id" binding:"required"`
	ResourceGroupName  string         `json:"resource_group_name" binding:"required"`
	Location           string         `json:"location" binding:"required"`
	OpenAIResourceName string         `json:"openai_resource_name" binding:"required"`
	Account            Account        `json:"account" binding:"required"`
	Deployment         Deployment     `json:"deployment" binding:"required"`
	StorageAccount     StorageAccount `json:"storage_account" binding:"required"`
}

type Account struct {
	Identity *armcognitiveservices.Identity `json:"identity"`
	Kind     *string                        `json:"kind"`
	SKU      *armcognitiveservices.SKU      `json:"sku"`
	Tags     map[string]*string             `json:"tags"`
}

type Deployment struct {
	// Value: "base_model"
	Name string `json:"deployment_name" binding:"required"`

	// Sub object for all the deployment properties
	Properties Properties `json:"properties"`
}

type Properties struct {
	Name                  string          `json:"name" binding:"required"`                    // Value: "base_model"
	Model                 DeploymentModel `json:"model" binding:"required"`                   // Sub object
	CurrentCapacity       int32           `json:"current_capacity" binding:"required"`        // 80
	VersionUpgrageOptions string          `json:"version_upgrade_options" binding:"required"` // Value: "OnceCurrentVersionExpired"
	Sku                   Sku             `json:"sku" binding:"required"`                     // Sub object
}

type Sku struct {
	Name     string `json:"name" binding:"required"`     // Standard
	Capacity int32  `json:"capacity" binding:"required"` // 80
}

type DeploymentModel struct {
	Name    string `json:"name" binding:"required"`    // gpt-4
	Format  string `json:"format" binding:"required"`  // OpenAI
	Version string `json:"version" binding:"required"` // 1106-Preview
}

type StorageAccount struct {
	Name          string `json:"name" binding:"required"`
	ContainerName string `json:"container_name" binding:"required"`
	SKU           string `json:"sku" binding:"required"`
	Kind          string `json:"kind" binding:"required"`
	AccessTier    string `json:"access_tier" binding:"required"`
}

/* ****************************************** */
/* *********** Tenant Models ************** */
/* ****************************************** */

type User struct {
	ID         string `json:"id"`
	Email      string `json:"email" binding:"required"`
	FirstName  string `json:"first_name" binding:"required"`
	LastName   string `json:"last_name" binding:"required"`
	Occupation string `json:"occupation"`
	Status     string `json:"status"`
	TenantID   string `json:"tenant_id"`
	TenantRole string `json:"tenant_role"`
}

type Tenant struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// UserRole represents the possible roles a user can have in a tenant
type UserRole string

const (
	UserRoleAdmin  UserRole = "admin"
	UserRoleMember UserRole = "member"
)

type TenantMember struct {
	ID        string   `json:"id"`
	UserID    string   `json:"user_id"`
	TenantID  string   `json:"tenant_id"`
	Role      UserRole `json:"role"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type TenantInvitation struct {
	ID              string   `json:"id"`
	TenantID        string   `json:"tenant_id"`
	InviterID       string   `json:"inviter_id"`
	InviteeEmail    string   `json:"invitee_email"`
	Role            UserRole `json:"role"`
	Status          string   `json:"status"`
	InvitationToken string   `json:"invitation_token"`
	ExpiresAt       string   `json:"expires_at"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

type ResourceGroupType string
type ResourceType string

const (
	ResourceGroupProject   ResourceGroupType = "project"
	ResourceGroupTemplate  ResourceGroupType = "template"
	ResourceGroupCommunity ResourceGroupType = "community"

	ResourceTypeDocument ResourceType = "document"
	ResourceTypeDiagram  ResourceType = "diagram"
	ResourceTypeTChart   ResourceType = "tchart"
	ResourceTypePnC      ResourceType = "pros&cons"
	ResourceTypeSwot     ResourceType = "swot"
	ResourceTypeMatrix   ResourceType = "decision_matrix"
)

func (rgt ResourceGroupType) IsValid() bool {
	switch rgt {
	case ResourceGroupProject, ResourceGroupTemplate, ResourceGroupCommunity:
		return true
	default:
		return false
	}
}

func (rt ResourceType) IsValid() bool {
	switch rt {
	case ResourceTypeDocument,
		ResourceTypeDiagram,
		ResourceTypeTChart,
		ResourceTypePnC,
		ResourceTypeSwot,
		ResourceTypeMatrix:
		return true
	default:
		return false
	}
}

type ResourceIdentifier struct {
	ResourceGroupType ResourceGroupType
	ResourceGroupID   *string
	ResourceType      *ResourceType
	ResourceID        *string
}
