package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/lib/pq" // Import pq for handling arrays

	"github.com/neozhixuan/gt_assessment/database"
	"github.com/neozhixuan/gt_assessment/models"
	"github.com/neozhixuan/gt_assessment/utils"
)

func GetSchemes(w http.ResponseWriter, r *http.Request) {
	var schemes []models.Scheme

	// Query to fetch all schemes with criteria_ids and benefit_ids
	query := `
        SELECT schemes.id, schemes.name, 
        ARRAY(SELECT criteria_id FROM scheme_criteria WHERE scheme_id = schemes.id) AS criteria_ids, 
        ARRAY(SELECT benefit_id FROM scheme_benefits WHERE scheme_id = schemes.id) AS benefit_ids
        FROM schemes
    `
	rows, err := database.DB.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching schemes: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var scheme models.Scheme
		var criteriaIDs, benefitIDs pq.StringArray // arrays for criteria and benefit IDs

		// Scan the scheme row, retrieving criteria_ids and benefit_ids as arrays
		err := rows.Scan(&scheme.ID, &scheme.Name, &criteriaIDs, &benefitIDs)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching scheme: %v", err), http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("Error iterating scheme: %v", err), http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("Error fetching applicant: %v", err), http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("Error fetching children: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Create a set to track all the unique education levels of the applicant's children
	childrenEducationLevels := make(map[string]bool)
	for rows.Next() {
		var childID string
		var dob string
		if err := rows.Scan(&childID, &dob); err != nil {
			http.Error(w, fmt.Sprintf("Error scanning child: %v", err), http.StatusInternalServerError)
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

	// We check that the scheme has criterias
	// Then, we left join each criteria to a criteria in the criteria table
	// Then, we check that our conditions are fulfilled
	// In the last row, we check that the count of actual criteria matched with the total number of criteria specified for the scheme
	schemeQuery := `
    SELECT schemes.id, schemes.name
    FROM schemes
    LEFT JOIN scheme_criteria ON schemes.id = scheme_criteria.scheme_id
    LEFT JOIN criteria ON criteria.id = scheme_criteria.criteria_id
    WHERE (criteria.marital_status IS NULL OR criteria.marital_status = $1)
      AND (criteria.employment_status IS NULL OR criteria.employment_status = $2)
      AND (criteria.education_levels IS NULL OR criteria.education_levels && $3::text[])
    GROUP BY schemes.id, schemes.name
    HAVING COUNT(criteria.id) = (
        SELECT COUNT(*) FROM scheme_criteria WHERE scheme_criteria.scheme_id = schemes.id
    );`

	// Make the applicant's children education array into a Postgres Array
	// Use it in our SQL query
	educationLevelArray := pq.Array(applicant.ChildrenLevel)
	rows, err = database.DB.Query(schemeQuery, applicant.MaritalStatus, applicant.EmploymentStatus, educationLevelArray)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching eligible scheme: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Update our final array for return
	for rows.Next() {
		var scheme models.Scheme
		if err := rows.Scan(&scheme.ID, &scheme.Name); err != nil {
			http.Error(w, fmt.Sprintf("Error scanning scheme: %v", err), http.StatusInternalServerError)
			return
		}
		schemes = append(schemes, scheme)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Error iterating scheme: %v", err), http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("Error input payload: %v", err), http.StatusBadRequest)
		return
	}

	// Build update query dynamically based on non-empty fields
	query := "UPDATE schemes SET "
	values := []interface{}{}
	counter := 1

	if scheme.Name != "" {
		query += "name = $" + fmt.Sprint(counter) + ", "
		values = append(values, scheme.Name)
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
		http.Error(w, fmt.Sprintf("Error updating scheme: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Scheme updated successfully"))
}

func DeleteScheme(w http.ResponseWriter, r *http.Request) {
	// Extract scheme ID from URL
	schemeID := r.URL.Query().Get("scheme")
	if schemeID == "" {
		http.Error(w, "scheme ID is required", http.StatusBadRequest)
		return
	}

	// Start a transaction to ensure atomicity of the delete operations
	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error starting tx: %v", err), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// First delete the associated entries in scheme_criteria table
	_, err = tx.Exec(`DELETE FROM scheme_criteria WHERE scheme_id = $1`, schemeID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deletring scheme-c: %v", err), http.StatusInternalServerError)
		return
	}

	// Then delete the associated entries in scheme_benefits table
	_, err = tx.Exec(`DELETE FROM scheme_benefits WHERE scheme_id = $1`, schemeID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting scheme-b: %v", err), http.StatusInternalServerError)
		return
	}

	// Finally, delete the scheme itself from the schemes table
	_, err = tx.Exec(`DELETE FROM schemes WHERE id = $1`, schemeID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting scheme: %v", err), http.StatusInternalServerError)
		return
	}

	// Commit the transaction to save all the changes
	err = tx.Commit()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error committing tx: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Scheme and associated data deleted successfully"))
}

func CreateScheme(w http.ResponseWriter, r *http.Request) {
	var requestBody models.SchemesRequest
	// Decode the request body
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, fmt.Sprintf("Error payload: %v", err), http.StatusBadRequest)
		return
	}

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error starting tx: %v", err), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for _, scheme := range requestBody.Schemes {
		log.Println(scheme)
		// 1. Insert criteria into the `criteria` table
		criteriaID := uuid.New().String() // Use a function to generate a new UUID
		_, err = tx.Exec(
			`INSERT INTO criteria (id, employment_status, marital_status, education_levels) VALUES ($1, $2, $3, $4)`,
			criteriaID, utils.NilIfEmpty(scheme.Criteria.EmploymentStatus), utils.NilIfEmpty(scheme.Criteria.MaritalStatus), pq.Array(scheme.Criteria.EducationLevels),
		)
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to insert criteria", http.StatusInternalServerError)
			return
		}

		// 2. Insert the scheme into the `schemes` table
		_, err = tx.Exec(
			`INSERT INTO schemes (id, name) VALUES ($1, $2)`,
			scheme.ID, scheme.Name,
		)
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to insert scheme", http.StatusInternalServerError)
			return
		}

		// 3. Insert the relationship into `scheme_criteria`
		_, err = tx.Exec(
			`INSERT INTO scheme_criteria (scheme_id, criteria_id) VALUES ($1, $2)`,
			scheme.ID, criteriaID,
		)
		if err != nil {
			http.Error(w, "Failed to insert scheme_criteria relationship", http.StatusInternalServerError)
			return
		}

		// 4. Insert the benefits into the `benefits` table and link them to the scheme
		for _, benefit := range scheme.Benefits {
			// Insert the benefit
			_, err = tx.Exec(
				`INSERT INTO benefits (id, name, amount) VALUES ($1, $2, $3)`,
				benefit.ID, benefit.Name, benefit.Amount,
			)
			if err != nil {
				http.Error(w, "Failed to insert benefit", http.StatusInternalServerError)
				return
			}

			// Link the benefit to the scheme
			_, err = tx.Exec(
				`INSERT INTO scheme_benefits (scheme_id, benefit_id) VALUES ($1, $2)`,
				scheme.ID, benefit.ID,
			)
			if err != nil {
				http.Error(w, "Failed to insert scheme_benefit relationship", http.StatusInternalServerError)
				return
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Scheme(s) created successfully"))
}
