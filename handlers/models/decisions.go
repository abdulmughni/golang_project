package models

// TBar
// This struct is used to create new TBar analysis
type TBarAnalysisWithOptions struct {
	TBarAnalysis
	OptionA  string  `json:"option_a" binding:"required"`
	OptionB  string  `json:"option_b" binding:"required"`
	TenantId *string `json:"tenant_id"`
}

type TBarAnalysis struct {
	ID               string  `db:"id" json:"id"`
	UserID           string  `db:"user_id" json:"user_id"`
	TenantID         string  `db:"tenant_id" json:"tenant_id"`
	TBarTitle        *string `db:"tbar_title" json:"tbar_title"`
	TBarDescription  *string `db:"tbar_description" json:"tbar_description"`
	TBarStatus       *string `db:"tbar_status" json:"tbar_status"`
	TBarCategory     *string `db:"tbar_category" json:"tbar_category"`
	TBarBetterOption *string `db:"tbar_better_option" json:"tbar_better_option"`
	Assumptions      *string `db:"assumptions" json:"assumptions"`
	FinalDecision    *string `db:"final_decision" json:"final_decision"`
	ADecisionId      *string `db:"architectural_decision_id" json:"architectural_decision_id"`
	Implications     *string `db:"implications" json:"implications"`
	CreatedAt        *string `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt        *string `db:"updated_at" json:"updated_at,omitempty"`
	ProjectID        *string `db:"project_id" json:"project_id"`
}

type TBarOptions struct {
	ID             string  `db:"id" json:"id"`
	UserID         string  `db:"user_id" json:"user_id"`
	TBarAnalysisID string  `db:"tbar_analysis_id" json:"tbar_analysis_id"`
	OptionTitle    string  `db:"option_title" json:"option_title"`
	TenantId       *string `json:"tenant_id"`
}

type TBarArgument struct {
	ID             string  `db:"id" json:"id"`
	UserID         string  `db:"user_id" json:"user_id"`
	TenantID       string  `db:"tenant_id" json:"tenant_id"`
	OptionID       string  `db:"option_id" json:"option_id"`
	ArgumentName   string  `db:"argument_name" json:"argument_name"`
	ArgumentWeight int     `db:"argument_weight" json:"argument_weight"`
	Description    *string `json:"description"`
}

// Pros and Cons Analysis
type PncAnalysis struct {
	ID             string  `json:"id" db:"id"`
	UserID         string  `json:"user_id" db:"user_id"`
	TenantID       string  `json:"tenant_id" db:"tenant_id"`
	Title          *string `json:"title" db:"title"`
	PNCDescription *string `json:"pnc_description" db:"pnc_description"`
	PNCStatus      *string `json:"pnc_status" db:"pnc_status"`
	Category       *string `json:"category" db:"category"`
	BetterOption   *string `json:"better_option" db:"better_option"`
	Assumptions    *string `json:"assumptions" db:"assumptions"`
	FinalDecision  *string `json:"final_decision" db:"final_decision"`
	ADecisionId    *string `json:"architectural_decision_id" db:"architectural_decision_id"`
	Implications   *string `json:"implications" db:"implications"`
	ProjectID      string  `json:"project_id" db:"project_id"`
}

// Pros and Cons Argument
type PncArgument struct {
	ID             string  `json:"id,omitempty"`
	UserID         string  `json:"user_id,omitempty"`
	TenantID       string  `json:"tenant_id,omitempty"`
	PncID          string  `json:"pnc_id,omitempty"`
	Argument       *string `json:"argument,omitempty"`
	ArgumentWeight *int    `json:"argument_weight,omitempty"`
	Side           string  `json:"side,omitempty"`
	Description    *string `json:"description"`
}

// SWOT

type Swot struct {
	ID              *string `json:"id" db:"id"`
	UserId          *string `json:"user_id"`
	TenantID        *string `json:"tenant_id"`
	Title           *string `json:"title"`
	SwotDescription *string `json:"swot_description"`
	SwotStatus      *string `json:"swot_status"`
	Category        *string `json:"category"`
	Assumptions     *string `json:"assumptions"`
	FinalDecision   *string `json:"final_decision"`
	ADecisionId     *string `json:"architectural_decision_id"`
	Implications    *string `json:"implications"`
	ProjectID       *string `json:"project_id" db:"project_id"` // New field for project_id
}

// SWOT Argument struct with project_id
type SwotArgument struct {
	ID             *string `json:"id" db:"id"`
	SwotID         *string `json:"swot_id" db:"swot_id"`
	UserID         *string `json:"user_id" db:"user_id"`
	TenantID       *string `json:"tenant_id" db:"tenant_id"`
	Argument       *string `json:"argument" db:"argument"`
	ArgumentWeight *int    `json:"argument_weight" db:"argument_weight"`
	Side           *string `json:"side" db:"side"`
	Description    *string `json:"description"`
}

// Decision Matrix
// Decision Matrix
type Matrix struct {
	Id                *string `json:"id" db:"id"`
	UserID            *string `json:"user_id"`
	TenantID          *string `json:"tenant_id"`
	Title             *string `json:"title"`
	MatrixDescription *string `json:"matrix_description"`
	MatrixStatus      *string `json:"matrix_status"`
	Category          *string `json:"category"`
	Assumptions       *string `json:"assumptions"`
	FinalDecision     *string `json:"final_decision"`
	ADecisionId       *string `json:"architectural_decision_id"`
	Implications      *string `json:"implications"`
	ProjectID         *string `json:"project_id" db:"project_id"` // New field for project association
}

type MatrixCriteria struct {
	Id                      string `json:"id" db:"id"`
	MatrixID                string `json:"matrix_id" db:"matrix_id"`
	UserID                  string `json:"user_id" db:"user_id"`
	TenantID                string `json:"tenant_id" db:"tenant_id"`
	Title                   string `json:"title" db:"title"`
	CriteriaMultiplier      int    `json:"criteria_multiplier" db:"criteria_multiplier"`
	CriteriaMultiplierTitle string `json:"criteria_multiplier_title" db:"criteria_multiplier_title"`
}

type MatrixConcept struct {
	Id         string `json:"id" db:"id"`
	MatrixID   string `json:"matrix_id" db:"matrix_id"`
	UserID     string `json:"user_id" db:"user_id"`
	TenantID   string `json:"tenant_id" db:"tenant_id"`
	Title      string `json:"title" db:"title"`
	UserRating int    `json:"user_rating" db:"user_rating"`
}

type MatrixUserRating struct {
	Id         string  `json:"id" db:"id"`
	CriteriaID string  `json:"criteria_id" db:"criteria_id"`
	ConceptID  string  `json:"concept_id" db:"concept_id"`
	UserID     string  `json:"user_id" db:"user_id"`
	UserRating int     `json:"user_rating" db:"user_rating"`
	TenantId   *string `json:"tenant_id"`
}

type MatrixCriteriaWithConcepts struct {
	MatrixCriteria
	Concepts []MatrixConceptRating `json:"concepts"`
	TenantId *string               `json:"tenant_id"`
}

type MatrixConceptRating struct {
	Id         string  `json:"id" db:"id"`
	UserRating int     `json:"user_rating" db:"user_rating"`
	TenantId   *string `json:"tenant_id"`
}

// Decision Matrix
type MatrixAnalysis struct {
	Id                *string `json:"id" db:"id"`
	TenantId          *string `json:"tenant_id"`
	Title             *string `json:"title"`
	MatrixDescription *string `json:"matrix_description"`
	MatrixStatus      *string `json:"matrix_status"`
	Category          *string `json:"category"`
	Assumptions       *string `json:"assumptions"`
	FinalDecision     *string `json:"final_decision"`
	ADecisionId       *string `json:"architectural_decision_id"`
	Implications      *string `json:"implications"`
}
