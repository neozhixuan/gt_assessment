package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Initialise a global SQL DB object that can be accessed from anywhere
var DB *sql.DB

// Function to initialise DB
func InitDB() {
	// Initialise connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PASSWORD"))

	// Open a PostgreSQL connection and check for errors
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Ping the database and check for errors
	err = DB.Ping()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	// Initialize tables
	initTables()

	// Seed the database with initial data
	seedData()

	// Console log success
	log.Println("Set up database successfully.")
}

// Function to initialize tables if they don't exist
func initTables() {
	applicantsTable := `
	CREATE TABLE IF NOT EXISTS applicants (
		id UUID PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		employment_status VARCHAR(50) NOT NULL,
		sex VARCHAR(10) NOT NULL,
		date_of_birth DATE NOT NULL
	);`

	applicationsTable := `
	CREATE TABLE IF NOT EXISTS applications (
		id UUID PRIMARY KEY,
		applicant_id UUID REFERENCES applicants(id),
		scheme_id UUID REFERENCES schemes(id),
		status VARCHAR(50) NOT NULL
	);`

	relationsTable := `
	CREATE TABLE IF NOT EXISTS relations (
		id1 UUID,
		id2 UUID,
		relation text,
		PRIMARY KEY (id1, id2)
	);`

	schemesTable := `
	CREATE TABLE IF NOT EXISTS schemes (
		id UUID PRIMARY KEY,
		name VARCHAR(255) NOT NULL
	);`

	criteriaTable := `CREATE TABLE IF NOT EXISTS criteria (
		id UUID PRIMARY KEY,
		marital_status VARCHAR(50) CHECK (marital_status IN ('single', 'married', 'widowed', 'divorced') OR marital_status IS NULL),
		employment_status VARCHAR(50) CHECK (employment_status IN ('employed', 'unemployed') OR employment_status IS NULL),
		education_levels TEXT[]
	);`

	benefitsTable := `CREATE TABLE IF NOT EXISTS benefits (
		id UUID PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		amount NUMERIC(10, 2) NOT NULL  -- Benefit amount	
	);`

	schemeCriteriaTable := `CREATE TABLE IF NOT EXISTS scheme_criteria (
		scheme_id UUID REFERENCES schemes(id) ON DELETE CASCADE,
		criteria_id UUID REFERENCES criteria(id) ON DELETE CASCADE,
		PRIMARY KEY (scheme_id, criteria_id)
	);`

	schemeBenefitsTable := `CREATE TABLE IF NOT EXISTS scheme_benefits (
		scheme_id UUID REFERENCES schemes(id) ON DELETE CASCADE,
		benefit_id UUID REFERENCES benefits(id) ON DELETE CASCADE,
		PRIMARY KEY (scheme_id, benefit_id)
	);`

	// Execute table creation
	// We only use := at the start to instantiate the "err" variable
	_, err := DB.Exec(applicantsTable)
	if err != nil {
		log.Fatalf("Error creating applicants table: %v", err)
	}

	_, err = DB.Exec(relationsTable)
	if err != nil {
		log.Fatalf("Error creating relations table: %v", err)
	}

	_, err = DB.Exec(schemesTable)
	if err != nil {
		log.Fatalf("Error creating schemes table: %v", err)
	}

	_, err = DB.Exec(applicationsTable)
	if err != nil {
		log.Fatalf("Error creating applications table: %v", err)
	}

	_, err = DB.Exec(criteriaTable)
	if err != nil {
		log.Fatalf("Error creating criteria table: %v", err)
	}

	_, err = DB.Exec(benefitsTable)
	if err != nil {
		log.Fatalf("Error creating benefits table: %v", err)
	}

	_, err = DB.Exec(schemeCriteriaTable)
	if err != nil {
		log.Fatalf("Error creating scheme-criteria table: %v", err)
	}

	_, err = DB.Exec(schemeBenefitsTable)
	if err != nil {
		log.Fatalf("Error creating scheme-benefits table: %v", err)
	}

}

// Function to seed initial data
func seedData() {
	// Check if applicants already exist
	row := DB.QueryRow(`SELECT COUNT(*) FROM applicants`)
	var count int
	err := row.Scan(&count)
	if err != nil {
		log.Fatalf("Error checking applicant count: %v", err)
	}

	// If there are no applicants, seed the database
	if count == 0 {
		log.Println("No entries detected. Seeding database...")

		insertApplicants := `
		INSERT INTO applicants (id, name, employment_status, sex, date_of_birth)
		VALUES
		('01913b7a-4493-74b2-93f8-e684c4ca935c'::uuid, 'James', 'unemployed', 'male', '1990-07-01'),
		('01913b80-2c04-7f9d-86a4-497ef68cb3a0'::uuid, 'Mary', 'unemployed', 'female', '1984-10-06'),
		('01913b88-1d4d-7152-a7ce-75796a2e8ecf'::uuid, 'Gwen', 'unemployed', 'female', '2016-02-01'),
		('01913b88-65c6-7255-820f-9c4dd1e5ce79'::uuid, 'Jayden', 'unemployed', 'male', '2018-03-15');
		`
		_, err := DB.Exec(insertApplicants)
		if err != nil {
			log.Fatalf("Error inserting seed applicants: %v", err)
		}

		// Insert seed data for relation
		// id2 is related to id1 via relation
		insertRelation := `
		INSERT INTO relations (id1, id2, relation)
		VALUES
		('01913b80-2c04-7f9d-86a4-497ef68cb3a0'::uuid, '01913b88-1d4d-7152-a7ce-75796a2e8ecf'::uuid, 'child'),
		('01913b80-2c04-7f9d-86a4-497ef68cb3a0'::uuid, '01913b88-65c6-7255-820f-9c4dd1e5ce79'::uuid, 'child');
		`
		_, err = DB.Exec(insertRelation)
		if err != nil {
			log.Fatalf("Error inserting seed relation: %v", err)
		}

		// Insert seed data for schemes
		insertSchemes := `
		-- Insert the first scheme
		INSERT INTO schemes (id, name) 
		VALUES 
		('01913b89-9a43-7163-8757-01cc254783f3'::uuid, 'Retrenchment Assistance Scheme'),
		('01913b89-befc-7ae3-bb37-3079aa7f1be0'::uuid, 'Retrenchment Assistance Scheme (families)');
		`

		_, err = DB.Exec(insertSchemes)
		if err != nil {
			log.Fatalf("Error inserting seed schemes: %v", err)
		}

		// Insert seed data for criteria
		insertCriteria := `
		-- Insert criteria for the first scheme
		INSERT INTO criteria (id, marital_status, employment_status, education_levels)
		VALUES 
		('01913b89-9a43-7163-8757-01cc254783f3'::uuid, NULL, 'unemployed', NULL),
		('01913b89-befc-7ae3-bb37-3079aa7f1be0'::uuid, NULL, 'unemployed', ARRAY['primary']);
		`
		_, err = DB.Exec(insertCriteria)
		if err != nil {
			log.Fatalf("Error inserting seed schemes: %v", err)
		}

		// Insert seed data for benefit
		insertBenefit := `
		INSERT INTO benefits (id, name, amount)
		VALUES ('01913b8b-9b12-7d2c-a1fa-ea613b802ebc'::uuid, 'SkillsFuture Credits', 500.00);
		`

		_, err = DB.Exec(insertBenefit)
		if err != nil {
			log.Fatalf("Error inserting seed schemes: %v", err)
		}

		// Insert seed data for schemes-benefit
		insertSchemesBenefit := `
		INSERT INTO scheme_benefits (scheme_id, benefit_id)
		VALUES ('01913b89-9a43-7163-8757-01cc254783f3'::uuid, '01913b8b-9b12-7d2c-a1fa-ea613b802ebc'::uuid);
		`

		_, err = DB.Exec(insertSchemesBenefit)
		if err != nil {
			log.Fatalf("Error inserting scheme-benefit: %v", err)
		}

		// Insert seed data for schemes-benefit
		insertSchemesCriteria := `
		INSERT INTO scheme_criteria (scheme_id, criteria_id)
		VALUES 
		('01913b89-9a43-7163-8757-01cc254783f3'::uuid, '01913b89-9a43-7163-8757-01cc254783f3'::uuid),
		('01913b89-befc-7ae3-bb37-3079aa7f1be0'::uuid, '01913b89-befc-7ae3-bb37-3079aa7f1be0'::uuid);
		`

		_, err = DB.Exec(insertSchemesCriteria)
		if err != nil {
			log.Fatalf("Error inserting scheme-criteria: %v", err)
		}

	} else {
		log.Println("Database has been populated, no seeding done.")
	}

	log.Println("Database successfully initialized and seeded.")
}
