package models

// Scheme represents a financial assistance scheme.
type Scheme struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	CriteriaIDs     []string `json:"criteria_ids"` // References to criteria table
	BenefitIDs      []string `json:"benefit_ids"`  // References to benefits table
	EducationLevels []string `json:"education_levels"`
}

// Criteria represents the conditions for eligibility.
type Criteria struct {
	ID               string `json:"id"`
	MaritalStatus    string `json:"marital_status"`    // Single, Married, Widowed, Divorced
	EmploymentStatus string `json:"employment_status"` // Employed, Unemployed
}

// Benefit represents the benefits that can be granted under a scheme.
type Benefit struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Amount float64 `json:"amount"` // Monetary value of the benefit
}
