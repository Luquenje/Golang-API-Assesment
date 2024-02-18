# Golang-API-Assesment

Develop a backend application that will be part of a system which teachers can use to perform administrative functions for their students. Teachers and students are identified by their email addresses.

# Use Instructions

### This program assumes that Go, PostgreSQL (and WSL depending if the system needs it to run PostgreSQL) have already been installed on the device. I am using Go ver 1.22.0 for Windows and PostgreSQL ver 16.2 for Windows.

1. Open and run PostgreSQL on your local machine
2. Go to the .env file and change the POSTGRESQL_CONNECTION_STRING to mach your database specifications. The most notable one is your database password. I will be using the default user and dbname which is postgres.
3. Open the command line in the project folder and run **make get** to get the packages used in this project.
4. Run **make run** in the command line to run the server.
5. Run **make test** in the command line to run the unit tests.

# Packages Used

### 1. mux - github.com/gorilla/mux

HTTP router and URL matcher

### 2. postgresql - github.com/lib/pq

Database of choice for this assesment

### 3. godotenv - github.com/joho/godotenv

Create and access environment variables in a .env file
