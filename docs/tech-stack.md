# Frontend
Cross-platform mobile app

- Flutter for pages/UI

# API Backend
Call APIs (not at query-time if possible?) for dynamic data on user-reported symptoms. Algorithm for querying data lives here and is returned to the frontend app. 

Data sources:
- Static symptoms database
    - symptoms to diseases mapping
- User info collections
    - Location, air quality, season, time, previous history
- External health APIs with symptoms information
    - live APIs that provide potential diagnoses and treatments based on symptom information

Options:
- Go
    - Pros: Concurrency and microservices for fast data querying
    - Cons: Probably need to learn on the fly
- JavaScript
    - Pros: Simplistic, infrastructure is 
    - Cons: Performance issues and lots of bugs
- Java
    - Pros: Robust and easy debugging, lots of infrastructure
    - Cons: Sometimes outdated and very strictly typed

# Database
Database will store user information and collected, indexed for fast query-time lookup

Options: 
- MongoDB
- PostgreSQL

# Data Ingestion
Ingesting data from static databases. Simple one-time ingest into our database, then need to form a logical view for quick access

Options: 
- Python (most likely because of in-class demo)

# Hosting
If necessary, it'd be easier to work on backend together if we could use a free hosting service so everything is stored on cloud

Options: 
- Render
- Railway
- Mongo Atlas (database)
