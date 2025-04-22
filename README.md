# File Storage Server and CLI

A simple file storage system with a server and CLI client that allows uploading files in chunks, deleting files, and listing uploaded files.

## Overview

The project consists of two main components:

1. **File Server**: A RESTful API server that handles file uploads, deletions, and listings.
2. **CLI Client**: A command-line interface that interacts with the file server.

The system is designed using Clean Architecture principles and follows dependency injection patterns with Google's Wire.

## Features

- Upload files in chunks with parallel processing
- Delete files from the server
- List uploaded files
- Progress bar display during uploads
- Retry mechanism for failed chunk uploads
- Designed for large file uploads up to configurable size limits

## File Server API

The server provides the following endpoints:

- `POST /api/v1/upload/init/<file_name>` - Initialize a file upload
- `POST /api/v1/upload/<upload_id>/<chunk_id>` - Upload a file chunk
- `DELETE /api/v1/<file_name>` - Delete a file
- `GET /api/v1/files` - List uploaded files

## CLI Commands

The CLI provides the following commands:

- `fs-store upload-file <file>` - Upload a file to the server
- `fs-store delete-file <file>` - Delete a file from the server
- `fs-store list-files` - List files on the server

## Environment Variables

### Server

- `PORT` - Server port (default: 8080)
- `BASE_STORAGE_DIR` - Base directory for file storage (default: current directory)
- `UPLOAD_SIZE_LIMIT` - Maximum file size in bytes (default: 104857600, 100 MiB)
- `MAX_BODY_SIZE` - Maximum request body size in bytes (default: 10485760, 10 MiB)
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 3306)
- `DB_USER` - Database user (default: root)
- `DB_PASSWORD` - Database password (default: empty)
- `DB_NAME` - Database name (default: fs_store)
- `DB_CONN_TIMEOUT` - Database connection timeout in seconds (default: 10)

### CLI

- `SERVER_URL` - Server URL (default: http://localhost:38080)
- `CHUNK_SIZE` - Chunk size in bytes (default: 1048576, 1 MiB)
- `RETRIES` - Number of retries for failed chunk uploads (default: 3)

## Installation

### Prerequisites

- Go 1.18 or later
- MySQL 5.7 or later

### Building the Server

```bash
git clone https://github.com/yourusername/fs-store.git
cd fs-store
go build -o fs-server ./cmd/server
```

### Building the CLI

```bash
go build -o fs-store ./cmd/cli
```

## Usage

### Starting the Server

```bash
# Create the database
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS fs_store"

# Start the server
./fs-server
```

### Using the CLI

```bash
# Upload a file
./fs-store upload-file path/to/file.txt

# Delete a file
./fs-store delete-file file.txt

# List files
./fs-store list-files

# Specifying a different server
./fs-store upload-file path/to/file.txt --server http://different-server:8080
```

## Architecture

The project follows Clean Architecture principles with the following layers:

1. **Core Layer** - Contains entities and repository interfaces
2. **Application Layer** - Contains use cases that implement business logic
3. **Interface Layer** - Contains controllers/handlers that adapt external interfaces
4. **Infrastructure Layer** - Contains implementation details for repositories, database, etc.

Dependency Injection is managed using Google's Wire package.

## Database Schema

### `files` Table

| Column           | Type                                       | Description                             |
|------------------|--------------------------------------------|-----------------------------------------|
| id               | UNSIGNED BIGINT (PK, AUTO_INCREMENT)       | Primary key                             |
| name             | VARCHAR(255) (UNIQUE)                      | File name                               |
| size             | UNSIGNED BIGINT                            | Total file size                         |
| status           | ENUM('INITIALIZED','PROCESSING','FAILED','COMPLETED') | File upload status            |
| total_chunks     | UNSIGNED INT                               | Total number of chunks                  |
| completed_chunks | UNSIGNED INT                               | Number of completed chunks              |
| created_at       | DATETIME                                   | Creation timestamp                      |
| updated_at       | DATETIME                                   | Last update timestamp                   |

### `file_chunks` Table

| Column           | Type                                       | Description                             |
|------------------|--------------------------------------------|-----------------------------------------|
| id               | UNSIGNED BIGINT (PK, AUTO_INCREMENT)       | Primary key                             |
| parent_id        | UNSIGNED BIGINT (FK)                       | Foreign key to files.id                 |
| status           | ENUM('INITIALIZED','PROCESSING','FAILED','COMPLETED') | Chunk upload status          |
| file_path        | VARCHAR(1024)                              | Path to the chunk file                  |
| created_at       | DATETIME                                   | Creation timestamp                      |
| updated_at       | DATETIME                                   | Last update timestamp                   |

## License

MIT 

# Database Migration Tool

A tool for creating and managing database migrations for the fs-store project.

## Features

- Create a SQL logical database (fs_store)
- Create a superuser with password "superpass"
- Create database tables for files and file_chunks

## Requirements

- Go 1.18+
- MySQL 5.7+ or MariaDB 10.2+

## Usage

Build the migration tool:

```bash
go build -o migrate ./cmd/migrate
```

### Apply migrations

Apply all migrations:

```bash
./migrate -command up
```

Apply specific number of migrations:

```bash
./migrate -command up -steps 1
```

### Rollback migrations

Rollback all migrations:

```bash
./migrate -command down
```

Rollback specific number of migrations:

```bash
./migrate -command down -steps 1
```

### Check current migration version

```bash
./migrate -command version
```

### Force migration version

```bash
./migrate -command force -version 3
```

### Custom database connection

```bash
./migrate -dsn "user:password@tcp(localhost:3306)/mysql?multiStatements=true" -command up
```

## Migration Files

Migrations are stored in the `db/migrations` directory and follow the naming convention:

```
{version}_{description}.{up|down}.sql
```

For example:
- `000001_create_database.up.sql`
- `000001_create_database.down.sql`

## Database Schema

### Files Table

| Column | Data Type | Nullable | Key / Constraint | Default | Explanation |
| --- | --- | --- | --- | --- | --- |
| id | UNSIGNED BIGINT | × | PRIMARY | AUTO_INCREMENT | |
| name | VARCHAR(255) | × | UNIQUE | | File name |
| size | UNSIGNED BIGINT | × | | 0 | Total file size |
| status | ENUM | × | INDEX | INITIALIZED | File upload status |
| total_chunks | UNSIGNED INT | × | CHECK | 0 | Denormalized for better performance |
| completed_chunks | UNSIGNED INT | ○ | | 0 | Denormalized for better performance |
| created_at | DATETIME | × | | CURRENT_TIMESTAMP | |
| updated_at | DATETIME | × | | CURRENT_TIMESTAMP | |

### File_Chunks Table

| Column | Data Type | Nullable | Key / Constraint | Default | Explanation |
| --- | --- | --- | --- | --- | --- |
| id | UNSIGNED BIGINT | × | PRIMARY | AUTO_INCREMENT | |
| parent_id | UNSIGNED BIGINT | × | INDEX, FOREIGN KEY | | Reference to files table |
| status | ENUM | × | | | Chunk upload status |
| file_path | VARCHAR(1024) | × | | | Path to chunk file |
| created_at | DATETIME | × | | CURRENT_TIMESTAMP | |
| updated_at | DATETIME | × | | CURRENT_TIMESTAMP | | 