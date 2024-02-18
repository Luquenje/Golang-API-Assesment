# Golang-API-Assesment

Develop a backend application that will be part of a system which teachers can use to perform administrative functions for their students. Teachers and students are identified by their email addresses.

I have hosted the Golang API at https://golang-api-assesment-production.up.railway.app/ using Railway's free trail plan.
It uses a PostgreSQL which is hosted on ElephantSQL.

# Use Instructions

### This program assumes that Go, PostgreSQL (and WSL depending if the system needs it to run PostgreSQL) have already been installed on the device. I am using Go ver 1.22.0 for Windows and PostgreSQL ver 16.2 for Windows.

1. Open and run PostgreSQL on your local machine
2. Go to the .env file and change the POSTGRESQL_CONNECTION_STRING to mach your database specifications if you are using a locally ran PostgreSQL. The most notable one is your database password. I have a commented out a connection string using the default user and dbname which is postgres. I am using a connection string of the PostgreSQL that I have hosted on ElephantSQL (the connection string and .env file is not supposed to be exposed in github but for this assesment I will expose them). You could use that instead of running a local instance of PostgreSQL if you wish.
3. Open the command line in the project folder and run "./build.bat" to run the server. 
   If you prefer using makefile, run "make get" to get the packages used in this project, then run "make run" to run the server.
4. Run "./build_test.bat" in the command line to run the unit tests.
   If you prefer using makefile, "make test".

# Packages Used

### 1. mux - github.com/gorilla/mux

HTTP router and URL matcher

### 2. postgresql - github.com/lib/pq

Database of choice for this assesment

### 3. godotenv - github.com/joho/godotenv

Create and access environment variables in a .env file

### 4. pgxpool - github.com/jackc/pgx/v5/pgxpool

Pool database connections
