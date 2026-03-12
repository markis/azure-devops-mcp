package client

import "time"

// DateFields holds all date/time related fields.
type DateFields struct {
	StartDate          *time.Time // Microsoft.VSTS.Scheduling.StartDate
	FinishDate         *time.Time // Microsoft.VSTS.Scheduling.FinishDate
	TargetDate         *time.Time // Microsoft.VSTS.Scheduling.TargetDate
	DueDate            *time.Time // Microsoft.VSTS.Scheduling.DueDate
	MarketDate         *time.Time // Custom.MarketDate
	DevCompleteDate    *time.Time // Custom.DevCompleteDate
	QCStartDate        *time.Time // Custom.QCStartDate
	QCCompleteDate     *time.Time // Custom.QCCompleteDate
	OriginalTargetDate *time.Time // Custom.OriginalTargetDate
}

// StatusFields holds status tracking fields.
type StatusFields struct {
	ActivatedBy     string     // Microsoft.VSTS.Common.ActivatedBy
	ActivatedDate   *time.Time // Microsoft.VSTS.Common.ActivatedDate
	ResolvedBy      string     // Microsoft.VSTS.Common.ResolvedBy
	ResolvedDate    *time.Time // Microsoft.VSTS.Common.ResolvedDate
	ResolvedReason  string     // Microsoft.VSTS.Common.ResolvedReason
	ClosedBy        string     // Microsoft.VSTS.Common.ClosedBy
	ClosedDate      *time.Time // Microsoft.VSTS.Common.ClosedDate
	StateChangeDate *time.Time // Microsoft.VSTS.Common.StateChangeDate
}

// PlanningFields holds planning/prioritization fields.
type PlanningFields struct {
	BusinessValue   *int     // Microsoft.VSTS.Common.BusinessValue
	StackRank       *float64 // Microsoft.VSTS.Common.StackRank (CRITICAL for backlog ordering!)
	Risk            string   // Microsoft.VSTS.Common.Risk
	TimeCriticality *float64 // Microsoft.VSTS.Common.TimeCriticality
	Rating          string   // Microsoft.VSTS.Common.Rating
	Triage          string   // Microsoft.VSTS.Common.Triage
}

// BuildFields holds build integration fields.
type BuildFields struct {
	FoundIn          string // Microsoft.VSTS.Build.FoundIn
	IntegrationBuild string // Microsoft.VSTS.Build.IntegrationBuild
	ClosedInBuild    string // Custom.ClosedinBuild
}

// SalesforceFields holds Salesforce integration fields.
type SalesforceFields struct {
	CaseNumber         string // Custom.SalesforceCaseNumber
	CaseStatus         string // Custom.SalesforceCaseStatus
	CaseClosed         *bool  // Custom.SalesforceCaseClosed
	URL                string // Custom.SalesforceURL
	EscalatingEngineer string // Custom.SalesforceEscalatingEngineer
	EscalationReason   string // Custom.SalesforceEscalationReason
	AttachmentsFolder  string // Custom.EscalationAttachmentsFolder
}

// RequirementFields holds requirement-related fields.
type RequirementFields struct {
	FunctionalRequirements    string // Custom.FunctionalRequirements (HTML)
	NonfunctionalRequirements string // Custom.NonfunctionalRequirements (HTML)
	BusinessCase              string // Custom.BusinessCase (HTML)
	SuggestedTests            string // Custom.SuggestedTests (HTML)
	RejectedIdeas             string // Custom.RejectedIdeas (HTML)
	Resources                 string // Custom.Resources (HTML)
}

// QualityFields holds quality/review fields.
type QualityFields struct {
	ApprovedBy                  string // Custom.ApprovedBy
	ReviewedBy                  string // Microsoft.VSTS.Common.ReviewedBy
	InitialDetailQuality        string // Custom.InitialDetailQuality (picklist)
	InitialDetailQualityComment string // Custom.InitialDetailQualityComment
	Documentation               string // Custom.Documentation (picklist)
	RCAReason                   string // Custom.RCAReason (picklist)
}

// MetricsFields holds metrics/tracking fields.
type MetricsFields struct {
	DaysAwaitingDev          *int       // Custom.DaysAwaitingDev
	DaysAwaitingSupport      *int       // Custom.DaysAwaitingSupport
	DaysSinceLastDevUpdate   *int       // Custom.DaysSinceLastDevUpdate
	TimeSpent                *float64   // Custom.TimeSpent
	PrioritizationScore      *int       // Custom.PrioritizationScore
	Confidence               *int       // Custom.Confidence
	RemainingWorkChangedDate *time.Time // Custom.RemainingWorkChangedDate
}

// SecurityFields holds security-specific fields.
type SecurityFields struct {
	CVENumber           string // Custom.CVENumber
	VulnerabilitySource string // Custom.VulnerabilitySource (picklist)
}

// FeatureSpecificFields holds fields specific to Feature work items.
type FeatureSpecificFields struct {
	AtRisk         *bool  // Custom.AtRisk [REQUIRED for Feature]
	DeliveryRisk   string // Custom.DeliveryRisk (picklist)
	RiskReason     string // Custom.RiskReason
	MitigationPlan string // Microsoft.VSTS.CMMI.MitigationPlan (HTML)
}

// BugSpecificFields holds fields specific to Bug work items.
type BugSpecificFields struct {
	ReproSteps  string // Microsoft.VSTS.TCM.ReproSteps (HTML) - already supported
	SystemInfo  string // Microsoft.VSTS.TCM.SystemInfo (HTML)
	Blocked     string // Microsoft.VSTS.CMMI.Blocked
	ProposedFix string // Microsoft.VSTS.CMMI.ProposedFix (HTML)
}

// UserStorySpecificFields holds fields specific to User Story work items.
type UserStorySpecificFields struct {
	DevOwner string // Custom.DevOwner
	Poker    string // Custom.Poker (picklist - planning poker)
}

// TestCaseSpecificFields holds fields specific to Test Case work items.
type TestCaseSpecificFields struct {
	Steps                string // Microsoft.VSTS.TCM.Steps (HTML)
	AutomatedTestName    string // Microsoft.VSTS.TCM.AutomatedTestName
	AutomatedTestStorage string // Microsoft.VSTS.TCM.AutomatedTestStorage
	AutomatedTestType    string // Microsoft.VSTS.TCM.AutomatedTestType
	AutomatedTestID      string // Microsoft.VSTS.TCM.AutomatedTestId
	AutomationStatus     string // Microsoft.VSTS.TCM.AutomationStatus
	Parameters           string // Microsoft.VSTS.TCM.Parameters (HTML)
	LocalDataSource      string // Microsoft.VSTS.TCM.LocalDataSource (HTML)
}

// CodeReviewFields holds fields for Code Review Request/Response.
type CodeReviewFields struct {
	Context        string     // Microsoft.VSTS.CodeReview.Context
	ContextCode    *int       // Microsoft.VSTS.CodeReview.ContextCode
	ContextOwner   string     // Microsoft.VSTS.CodeReview.ContextOwner
	ContextType    string     // Microsoft.VSTS.CodeReview.ContextType
	AcceptedBy     string     // Microsoft.VSTS.CodeReview.AcceptedBy
	AcceptedDate   *time.Time // Microsoft.VSTS.CodeReview.AcceptedDate
	ClosedStatus   string     // Microsoft.VSTS.CodeReview.ClosedStatus
	ClosingComment string     // Microsoft.VSTS.CodeReview.ClosingComment
}
