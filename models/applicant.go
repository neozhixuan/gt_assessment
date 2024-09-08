package models

// DB Schema
type Applicant struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	EmploymentStatus string `json:"employment_status"`
	Sex              string `json:"sex"`
	DateOfBirth      string `json:"date_of_birth"`
}

// Response Schema
type ApplicantResponse struct {
	ID               string   `json:"id"`
	MaritalStatus    string   `json:"marital_status"`
	EmploymentStatus string   `json:"employment_status"`
	ChildrenLevel    []string `json:"children_level"`
}
