# Financial Assistance Schemes Backend

## Description

This is a backend system to manage financial assistance schemes and applications for needy individuals and families. The backend is written in Go and uses PostgreSQL as the database.

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

3. Set up a PostgreSQL database in your local computer and configure the .env file:

   ```bash
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=yourusername
   DB_PASSWORD=yourpassword
   DB_NAME=financial_assistance
   ```

4. Run the application

   ```bash
   go run main.go
   ```

5. The server will start at http://localhost:8080/.

### API Endpoints

- GET /api/applicants - Get all applicants
- POST /api/applicants - Create a new applicant
- PUT /api/applicants?applicant={id} - Update an applicant
- DELETE /api/applicants?applicant={id} - Delete an applicant
- GET /api/schemes - Get all schemes
- GET /api/schemes/eligible?applicant={id} - Get eligible schemes for an applicant
- PUT /api/schemes?scheme={id} - Update a schemes
- DELETE /api/schemes?scheme={id} - Get all schemes
- GET /api/applications - Get all applications
- POST /api/applications - Create a new application
- PUT /api/applications - Update an application
- DELETE /api/applications?application={id} - Delete an application
