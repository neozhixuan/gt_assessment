package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/neozhixuan/gt_assessment/database"
	"github.com/neozhixuan/gt_assessment/models"
	"github.com/neozhixuan/gt_assessment/utils"

	"github.com/google/uuid"
)

// GET Request on Applicants table
// - Each request requires 2 objects from net/http: ResponseWriter and a Request
func GetApplicants(w http.ResponseWriter, r *http.Request) {
	// Initialise an empty list variable to store our list of applicants
	var applicants []models.Applicant

	// Query the database for all rows in applicants table
	rows, err := database.DB.Query(`SELECT id, name, employment_status, sex, date_of_birth FROM applicants`)
	// Error control by writing into the ResponseWriter + closing the DB
	if err != nil {
		utils.SendJSONResponse(w, http.StatusInternalServerError, fmt.Sprintf("Unable to fetch applicants: %v", err))
		return
	}
	defer rows.Close()

	// For each row
	for rows.Next() {
		// Initialise an empty variable to store each applicant
		var applicant models.Applicant
		// Fetch each individual data into the empty applicants variable
		if err := rows.Scan(&applicant.ID, &applicant.Name, &applicant.EmploymentStatus, &applicant.Sex, &applicant.DateOfBirth); err != nil {
			utils.SendJSONResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error scanning applicants: %v", err))
			return
		}
		// Append it to our list of applicants initialised at the top
		applicants = append(applicants, applicant)
	}
	// Write our response using ResponseWriter
	utils.SendJSONResponse(w, http.StatusOK, applicants)
}

// POST Request into Applicants table
func CreateApplicant(w http.ResponseWriter, r *http.Request) {
	// Initialise an empty variable to store one applicant
	var applicant models.Applicant

	// Generate a Decoder that reads from our request
	// Then, decode the request body into our empty variable
	err := json.NewDecoder(r.Body).Decode(&applicant)
	// Error control by writing into the ResponseWriter
	if err != nil {
		utils.SendJSONResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
		return
	}

	// Generate a new UUID for the applicant
	applicant.ID = uuid.New().String()

	// Insert applicant into the database
	query := `INSERT INTO applicants (id, name, employment_status, sex, date_of_birth) VALUES ($1, $2, $3, $4, $5)`
	_, err = database.DB.Exec(query, applicant.ID, applicant.Name, applicant.EmploymentStatus, applicant.Sex, applicant.DateOfBirth)
	if err != nil {
		utils.SendJSONResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error creating applicant: %v", err))
		return
	}
	// Write our response using ResponseWriter
	utils.SendJSONResponse(w, http.StatusCreated, applicant)
}

// Extrafct parameters using Mux library.
func UpdateApplicant(w http.ResponseWriter, r *http.Request) {
	// Extract applicant ID from URL
	applicantID := r.URL.Query().Get("applicant")
	if applicantID == "" {
		http.Error(w, "applicant ID is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var applicant models.Applicant
	if err := json.NewDecoder(r.Body).Decode(&applicant); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
		return
	}

	// Build update query dynamically based on non-empty fields
	// Stores data into a slice "[]" (dynamic array) of interface{} types
	query := "UPDATE applicants SET "
	// We initialise an array of empty slices, of any type
	// - interface{} means any type
	// - {} initializes an empty slice
	values := []interface{}{}
	counter := 1

	if applicant.Name != "" {
		query += "name = $" + fmt.Sprint(counter) + ", "
		values = append(values, applicant.Name)
		counter++
	}

	if applicant.EmploymentStatus != "" {
		query += "employment_status = $" + fmt.Sprint(counter) + ", "
		values = append(values, applicant.EmploymentStatus)
		counter++
	}

	if applicant.Sex != "" {
		query += "employment_status = $" + fmt.Sprint(counter) + ", "
		values = append(values, applicant.Sex)
		counter++
	}

	if applicant.DateOfBirth != "" {
		query += "employment_status = $" + fmt.Sprint(counter) + ", "
		values = append(values, applicant.DateOfBirth)
		counter++
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]

	// Add WHERE clause
	query += " WHERE id = $" + fmt.Sprint(counter)
	values = append(values, applicantID)

	// Execute the query
	_, err := database.DB.Exec(query, values...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating applicants: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Applicant updated successfully"))
}

func DeleteApplicant(w http.ResponseWriter, r *http.Request) {
	// Extract applicant ID from URL
	applicantID := r.URL.Query().Get("applicant")
	if applicantID == "" {
		http.Error(w, "applicant ID is required", http.StatusBadRequest)
		return
	}

	// Delete applicant from DB
	query := `DELETE FROM applicants WHERE id = $1`
	_, err := database.DB.Exec(query, applicantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deletings applicants: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Applicant deleted successfully"))
}
