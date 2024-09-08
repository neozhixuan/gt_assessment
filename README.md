# Financial Assistance Schemes Backend

## Goal

The goal of this assignment is to build a system that

-     Allows the management of a fictitious set of financial assistance schemes.
-     Save and update records of applicants who applied for schemes.
-     The system should advise users of schemes that each applicant can apply for.
-     The system should also save the outcome of granting of schemes to applicants.

## Description

This is a backend system to manage financial assistance schemes and applications for needy individuals and families. The backend is written in Go and uses PostgreSQL as the database. We use Mux as the router, and joho as the environment manager.

### Prerequisites

- Go 1.22.1
- PostgreSQL

### Setup Instructions

1. Clone the repository:

   ```bash
   git clone https://github.com/neozhixuan/gt_assessment.git
   cd gt_assessment
   ```

2. Install Go dependencies:

   ```bash
   go mod tidy
   ```

3. Set up a PostgreSQL database in your local computer

4. Configure the .env file:

   ```bash
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=yourusername
   DB_PASSWORD=yourpassword
   DB_NAME=financial_assistance
   ```

5. Run the application. The database will be populated and seeded automatically.

   ```bash
   go run main.go
   ```

6. The server will start at http://localhost:8080/.

### API Endpoints

- GET /api/applicants - Get all applicants
- POST /api/applicants - Create a new applicant
- PUT /api/applicants?applicant={id} - Update an applicant
- DELETE /api/applicants?applicant={id} - Delete an applicant
- GET /api/schemes - Get all schemes
- GET /api/schemes/eligible?applicant={id} - Get eligible schemes for an applicant
- PUT /api/schemes?scheme={id} - Update a schemes
- DELETE /api/schemes?scheme={id} - Delete a scheme
- GET /api/applications - Get all applications
- POST /api/applications - Create a new application
- PUT /api/applications - Update an application
- DELETE /api/applications?application={id} - Delete an application

### Database Design

The database has 8 tables.

1. applicants

2. applications

3. benefits (to track each benefit)

4. criteria (to track each set of criterias as one object)

5. relations

6. scheme_benefits

7. scheme_criteria

8. schemes

I separated criteria into its own object so that future changes to the criterion can be changed only in this object and will be decoupled from the main scheme changes.

I used scheme_benefits and scheme_criteria along with foreign key references to improve normalisation of the database, decoupling the schemes from its benefits and criteria and reducing the redundancy. I can also ensure consistency of references by using the keys as reference in these tables.

### Backend Logic / API Design

For the backend functions, I used the `err` design pattern in Golang to detect any errors during the PostgreSQL row retrieval functions like `QueryRow`, to ensure that every transaction's error was accounted for.

For functions with multiple changes, I used `tx.commit()` and `tx.rollback()` at the end to ensure that my transaction is committed in one go.

### Future Improvements

Due to the lack of time, I did not separate the criteria for a child's educational levels into a separate table. It is currently in the education_levels table as an array of strings, which is not the optimal database design.

At the same time, I could have made an applicants_children table to store the references from a applicant to their child instead of the current many-to-many relation within the `relations` table.
