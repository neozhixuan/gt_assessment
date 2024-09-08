package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lib/pq" // Import pq for handling arrays

	"github.com/neozhixuan/gt_assessment/database"
	"github.com/neozhixuan/gt_assessment/models"
	"github.com/neozhixuan/gt_assessment/utils"
)

func GetSchemes(w http.ResponseWriter, r *http.Request) {
	var schemes []models.Scheme

	// Query to fetch all schemes with criteria_ids and benefit_ids
	query := `SELECT id, name, criteria_ids, benefit_ids FROM schemes`
	rows, err := database.DB.Query(query)
	if err != nil {
		http.Error(w, "Error fetching schemes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var scheme models.Scheme
		var criteriaIDs, benefitIDs pq.StringArray // arrays for criteria and benefit IDs

		// Scan the scheme row, retrieving criteria_ids and benefit_ids as arrays
		err := rows.Scan(&scheme.ID, &scheme.Name, &criteriaIDs, &benefitIDs)
		if err != nil {
			http.Error(w, "Error scanning scheme", http.StatusInternalServerError)
			return
		}

		// Store criteria and benefits IDs as strings in the Scheme model
		scheme.CriteriaIDs = criteriaIDs
		scheme.BenefitIDs = benefitIDs

		// Append the scheme to the result list
		schemes = append(schemes, scheme)
	}

	// Check if there was an error during rows iteration
	if err = rows.Err(); err != nil {
		http.Error(w, "Error iterating over schemes", http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, schemes)
}

func GetEligibleSchemes(w http.ResponseWriter, r *http.Request) {
	// Get the ID from params
	applicantID := r.URL.Query().Get("applicant")
	if applicantID == "" {
		http.Error(w, "applicant ID is required", http.StatusBadRequest)
		return
	}

	// Populate the data from the applicant that we need
	var applicant models.ApplicantResponse
	query := `
		SELECT id, employment_status, 
		CASE 
			WHEN id IN (SELECT id1 FROM relations WHERE relation = 'spouse') THEN 'married' 
			ELSE 'single'
		END AS marital_status
		FROM applicants 
		WHERE id = $1`
	err := database.DB.QueryRow(query, applicantID).Scan(&applicant.ID, &applicant.EmploymentStatus, &applicant.MaritalStatus)
	if err != nil {
		http.Error(w, "Error fetching applicant", http.StatusInternalServerError)
		return
	}

	// Fetch children related to the applicant and calculate education levels
	childrenQuery := `
		SELECT id2, date_of_birth 
		FROM relations 
		JOIN applicants ON relations.id2 = applicants.id
		WHERE relations.id1 = $1 AND relations.relation = 'child'`
	rows, err := database.DB.Query(childrenQuery, applicantID)
	if err != nil {
		http.Error(w, "Failed to query children", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Create a set to track all the unique education levels of the applicant's children
	childrenEducationLevels := make(map[string]bool)
	for rows.Next() {
		var childID string
		var dob string
		if err := rows.Scan(&childID, &dob); err != nil {
			http.Error(w, "Failed to scan child", http.StatusInternalServerError)
			return
		}

		// Calculate the child's age
		childAge := utils.CalculateAge(dob)
		level := utils.CalculateEducationLevel(childAge)
		childrenEducationLevels[level] = true
	}

	// Convert map to a slice of unique education levels
	for level := range childrenEducationLevels {
		applicant.ChildrenLevel = append(applicant.ChildrenLevel, level)
	}

	// Fetch eligible schemes based on the applicant's marital status, employment status, and children education levels
	schemes := []models.Scheme{}

	// The unnest() function in SQL is used to expand an array into a set of rows.
	// We check that the scheme has criterias
	// Then, we left join each criteria to a criteria in the criteria table
	// Then, we check that our conditions are fulfilled
	// In the last row, we check that the count of actual criteria matched with the total number of criteria specified for the scheme
	schemeQuery := `
	SELECT schemes.id, schemes.name
	FROM schemes
	LEFT JOIN unnest(schemes.criteria_ids) AS criteria_id ON true
	LEFT JOIN criteria ON criteria.id = criteria_id::uuid
	WHERE (criteria.marital_status IS NULL OR criteria.marital_status = $1)
	AND (criteria.employment_status IS NULL OR criteria.employment_status = $2)
	AND (criteria.education_levels IS NULL OR criteria.education_levels && $3::text[])
	GROUP BY schemes.id, schemes.name
	HAVING COUNT(criteria.id) = (
		SELECT COUNT(*) FROM unnest(schemes.criteria_ids)
	);
	`

	// Make the applicant's children education array into a Postgres Array
	// Use it in our SQL query
	educationLevelArray := pq.Array(applicant.ChildrenLevel)
	rows, err = database.DB.Query(schemeQuery, applicant.MaritalStatus, applicant.EmploymentStatus, educationLevelArray)
	if err != nil {
		http.Error(w, "Failed to query eligible schemes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Update our final array for return
	for rows.Next() {
		var scheme models.Scheme
		if err := rows.Scan(&scheme.ID, &scheme.Name); err != nil {
			http.Error(w, "Failed to scan scheme", http.StatusInternalServerError)
			return
		}
		schemes = append(schemes, scheme)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, "Row iteration error", http.StatusInternalServerError)
		return
	}

	// Send eligible schemes as response
	utils.SendJSONResponse(w, http.StatusOK, schemes)
}

func UpdateScheme(w http.ResponseWriter, r *http.Request) {
	// Extract scheme ID from URL
	schemeID := r.URL.Query().Get("scheme")
	if schemeID == "" {
		http.Error(w, "scheme ID is required", http.StatusBadRequest)
		return
	}
	// Parse request body
	var scheme models.Scheme
	if err := json.NewDecoder(r.Body).Decode(&scheme); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Build update query dynamically based on non-empty fields
	query := "UPDATE applicants SET "
	values := []interface{}{}
	counter := 1

	if scheme.Name != "" {
		query += "name = $" + fmt.Sprint(counter) + ", "
		values = append(values, scheme.Name)
		counter++
	}

	if len(scheme.CriteriaIDs) > 0 {
		query += "name = $" + fmt.Sprint(counter) + ", "
		values = append(values, scheme.CriteriaIDs)
		counter++
	}

	if len(scheme.CriteriaIDs) > 0 {
		query += "name = $" + fmt.Sprint(counter) + ", "
		values = append(values, scheme.CriteriaIDs)
		counter++
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]

	// Add WHERE clause
	query += " WHERE id = $" + fmt.Sprint(counter)
	values = append(values, schemeID)

	// Execute the query
	_, err := database.DB.Exec(query, values...)
	if err != nil {
		http.Error(w, "Failed to update applicant", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Applicant updated successfully"))
}

func DeleteScheme(w http.ResponseWriter, r *http.Request) {
	// Extract scheme ID from URL
	schemeID := r.URL.Query().Get("scheme")
	if schemeID == "" {
		http.Error(w, "scheme ID is required", http.StatusBadRequest)
		return
	}

	// Delete scheme from DB
	query := `DELETE FROM schemes WHERE id = $1`
	_, err := database.DB.Exec(query, schemeID)
	if err != nil {
		http.Error(w, "Failed to delete scheme", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Scheme deleted successfully"))
}
