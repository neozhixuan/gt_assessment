package utils

import "time"

// Helper function to calculate education level based on age
func CalculateEducationLevel(age int) string {
	if age <= 6 {
		return "kindergarten"
	} else if age <= 12 {
		return "primary"
	} else if age <= 16 {
		return "secondary"
	} else if age <= 18 {
		return "tertiary"
	} else {
		return "higher"
	}
}

// Calculate age of child
func CalculateAge(dob string) int {
	parsedDOB, _ := time.Parse(time.RFC3339, dob)
	now := time.Now()
	years := now.Year() - parsedDOB.Year()

	if now.YearDay() < parsedDOB.YearDay() {
		years--
	}
	return years
}
