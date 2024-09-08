package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/neozhixuan/gt_assessment/database"
	"github.com/neozhixuan/gt_assessment/models"
	"github.com/neozhixuan/gt_assessment/utils"

	"github.com/google/uuid"
)

// GET request
func GetApplications(w http.ResponseWriter, r *http.Request) {
	// Get list of applications
	var applications []models.Application
	rows, err := database.DB.Query("SELECT id, applicant_id, scheme_id, status FROM applications")
	if err != nil {
		log.Fatalf("Error fetching applications: %v", err)
		http.Error(w, "Error fetching applications", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Save each application into an object and append it to our list
	for rows.Next() {
		var application models.Application
		err := rows.Scan(&application.ID, &application.ApplicantID, &application.SchemeID, &application.Status)
		if err != nil {
			log.Fatalf("Error scanning application: %v", err)
			http.Error(w, "Error processing applications", http.StatusInternalServerError)
			return
		}
		applications = append(applications, application)
	}

	// Write our response using ResponseWriter
	utils.SendJSONResponse(w, http.StatusOK, applications)
}

// POST request
func CreateApplication(w http.ResponseWriter, r *http.Request) {
	// Initialise empty variable to create application
	var application models.Application
	// Create a decoder using the request and decode the request into our variable
	err := json.NewDecoder(r.Body).Decode(&application)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Insert into DB using SQL, with a unique UUID
	application.ID = uuid.New().String()
	_, err = database.DB.Exec("INSERT INTO applications (id, applicant_id, scheme_id, status) VALUES ($1, $2, $3, $4)",
		application.ID, application.ApplicantID, application.SchemeID, application.Status)
	if err != nil {
		log.Fatalf("Error inserting application: %v", err)
		http.Error(w, "Error creating application", http.StatusInternalServerError)
		return
	}

	// Write our response using ResponseWriter
	utils.SendJSONResponse(w, http.StatusOK, application)
}

func DeleteApplication(w http.ResponseWriter, r *http.Request) {
	// Extract scheme ID from URL
	applicationID := r.URL.Query().Get("application")
	if applicationID == "" {
		http.Error(w, "application ID is required", http.StatusBadRequest)
		return
	}

	// Delete application from DB
	query := `DELETE FROM applications WHERE id = $1`
	_, err := database.DB.Exec(query, applicationID)
	if err != nil {
		http.Error(w, "Failed to delete application", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Application deleted successfully"))
}

// Extrafct parameters using Mux library.
func UpdateApplication(w http.ResponseWriter, r *http.Request) {
	// Extract applicant ID from URL
	applicationID := r.URL.Query().Get("application")
	if applicationID == "" {
		http.Error(w, "application ID is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var application models.Application
	if err := json.NewDecoder(r.Body).Decode(&application); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Build update query dynamically based on non-empty fields
	query := "UPDATE application SET "
	values := []interface{}{}
	counter := 1

	if application.ApplicantID != "" {
		query += "name = $" + fmt.Sprint(counter) + ", "
		values = append(values, application.ApplicantID)
		counter++
	}

	if application.SchemeID != "" {
		query += "name = $" + fmt.Sprint(counter) + ", "
		values = append(values, application.SchemeID)
		counter++
	}

	if application.Status != "" {
		query += "employment_status = $" + fmt.Sprint(counter) + ", "
		values = append(values, application.Status)
		counter++
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]

	// Add WHERE clause
	query += " WHERE id = $" + fmt.Sprint(counter)
	values = append(values, applicationID)

	// Execute the query
	_, err := database.DB.Exec(query, values...)
	if err != nil {
		http.Error(w, "Failed to update application", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Application updated successfully"))
}
