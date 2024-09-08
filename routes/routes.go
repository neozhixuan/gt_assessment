package routes

import (
	"log"

	"github.com/gorilla/mux"
	"github.com/neozhixuan/gt_assessment/controllers"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/api/applicants", controllers.GetApplicants).Methods("GET")
	r.HandleFunc("/api/applicants", controllers.CreateApplicant).Methods("POST")
	r.HandleFunc("/api/applicants", controllers.UpdateApplicant).Methods("PUT")
	r.HandleFunc("/api/applicants", controllers.DeleteApplicant).Methods("DELETE")
	r.HandleFunc("/api/schemes", controllers.GetSchemes).Methods("GET")
	r.HandleFunc("/api/schemes", controllers.DeleteScheme).Methods("DELETE")
	// r.HandleFunc("/api/schemes", controllers.CreateScheme).Methods("POST")
	r.HandleFunc("/api/schemes", controllers.UpdateScheme).Methods("PUT")
	r.HandleFunc("/api/schemes/eligible", controllers.GetEligibleSchemes).Methods("GET")
	r.HandleFunc("/api/applications", controllers.GetApplications).Methods("GET")
	r.HandleFunc("/api/applications", controllers.CreateApplication).Methods("POST")
	r.HandleFunc("/api/applications", controllers.UpdateApplication).Methods("PUT")
	r.HandleFunc("/api/applications", controllers.DeleteApplication).Methods("DELETE")
	log.Println("Set up routes.")
	return r
}
