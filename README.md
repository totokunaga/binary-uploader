# README

# File Storage Server & CLI

This project provides a simple file storage server accessible via a command-line interface (CLI).

## Prerequisites

- [**Docker**](https://www.docker.com/) & [**Docker Compose**](https://docs.docker.com/compose/install/)
- [**Go**](https://go.dev/doc/install) (v1.24+)
- `make` (Recommended for easy command execution)

## How to Use

These instructions cover running the Go file server and MySQL in Docker, and building/running the CLI locally.

*(Docker is used for reproducible builds and consistent environments; see Tech Stack section for details).*

1. **Build the CLI**

```bash
make cli
# Or, if make is unavailable:
# cd cli/cmd && go build -o ../../fs-store . && cd ../..
```

1. **Build and Run Server & Database (Docker)**

```bash
make up # Use 'make up-d' to run detached
# Or, if make is unavailable:
# docker compose -f docker-compose.yml --env-file .env up
```

- Defaults: Host port `38080` -> File Server, Host port `33306` -> MySQL.
- To change ports, edit `HOST_FILE_SERVER_PORT` / `HOST_MYSQL_PORT` in `.env`, then run `make build` and `make up`.
1. **Use the CLI**

```bash
# Upload
./fs-store upload-file <path_to_file>

# List
./fs-store list-files

# Delete
./fs-store delete-file <file_name>
```

- For detailed usage: `./fs-store help` or `./fs-store [command] --help`.
1. **Stop Server & Database**

```bash
make down
# Or, if make is unavailable:
# docker compose -f docker-compose.yml down
```

1. **Clean Up Resources (Volumes, etc.)**Bash

```bash
make clean
# Or, if make is unavailable:
# docker compose -f docker-compose.yml down -v --remove-orphans
# docker compose -f docker-compose.tool.yml run --rm clean
```

## Project Tooling

Use `make` for development tasks:

- `make wire`: Generate dependency injection code using [Google Wire](https://github.com/google/wire).
- `make test`: Run Go tests (`go test ./...`).
- `make lint`: Run Go linter (`golangci-lint run`).
- `make mockgen`: Run mockgen to generate mocks

## Software Distribution

- `make build-cli-all`: Create a CLI executable binary for Windows, MacOS, and Linux with AMD64 and ARM64 respectively.
- Created executables can be distributed to user through a direct download or package managers

## Software Distribution

## Technical Specifications

Details on implementation, technology choices, performance considerations, and potential extensions.

## Overview

Here’s the overview of how each feature gets completed in the CLI and the file server

### Upload

- A file to be uploaded is divided into small chunks, and the chunks are sent to the server in parallel.
- If any of the chukns fails (e.g. network error, database down, etc), the CLI retries to send failed chunks up to certain times.
- When all retries also fail and the user tries to upload the same file (determined by a file name and checksum), the CLI sends only failed chunks and not all the chunks to avoid pressuring the file server
- When the user tries to upload a file whose name conflicts with another file which is already uploaded to the file server, the CLI deletes the conflicting file in the file server and upload the new one after user’s confirmation

### Delete

- A file deletion request is handled in a synchronous manner between the CLI and the file server (meaning the file server doesn’t send a response to the CLI until the deletion of all the file chunks completes)
- File chunks are deleted in parallel

### List

- The CLI lists files whose chunks are all uploaded to the file server

## Tech Stack

- **Go**
    - **Gin**: Web framework for simple yet high performant REST API development
    - **GORM**: ORM to interact with database for database abstraction, type safety, etc
    - **slog**: Structured logs for consistent, concise loggings
- **Clean Architecture**: To decouple each functions per responsibility fully utilizing Go interfaces. This design keeps the code maintainable, testable, and flexible to external dependencies like database engine changes
- **Dependency Injection**: To achieve Clean Architecture by feeding concrete types to each layer (such as interface and use case layers)
- **Docker**: To achieve reproducible builds and consistent runtime environment, a container technology is crucial if the application may run more than one environment . Docker is the de-facto standard for it. (Especially, Go build must be done for the OS and the CPU architecture of the host machine which executes the binary. So the use of Docker helps removing the concern of the inconsistency between the end binary and a runtime environment)
- **MySQL with InnoDB**: MVCC keeps the system performant and concurrency-safer

## Technical Decisions for Performance

- **Dividing file into chunks**: For a gigantic file, uploading the entire binaries all at once can cause several propblems such as data can’t fit in a memory, slower network transmissions, overwhelming the file server, etc. By splitting the original data into pieces, it can achieve higher throughput and speed, efficient retries when failed, and also the upload progress is trackable (better UX)
- **Data compression before sending**: To increases the network bandwidth by reducing the size of the file chunks before sending it to the file server. This is a trade-off between CPU cycles and network bandwidth, so the compressing should be turned off if network bandwidth is well enough (default off, and configurable through `-z` flag)
- **Stream reading & writing**: For a gigantic file, loading all the file content on a memory overwhelms the host (both the CLI and the file server). Stream reading and writing solves this issue by reading/writing a certain amount of bytes at a time.
- **SQL table design and queries**:
    - **Composite Index Key**: When retrieving file chunk data or updating chunk status, the target records are looked up by using `parent_id` and `status`. Thus, the use of a composite key for those columns increases the query efficiency
    - **Avoiding N+1 Problem**: To avoid multiple database connections, query planning, etc, when multiple records need to be interacted, one big query is used rather than multiple small queries

## Discarded Ideas / Extensions

- **Asynchrounous uploading and deleting**
    - By employing a message queue (like Apache Kafka) and several consumer jobs, the file server can send the response back to the CLI as soon as a file chunk data is queued to the message queue, and the operation can potentially get quicker. Also, asynchronouly deleting files using a separate job process (like a cronjob) and the file server itself just mark a certain file as “deleted” would greatly reduce the workload of the server at a time
    - Discarded this idea mainly because I didn’t think it can be implemented within the deadline caring about all the potential edge cases in each component. Also I felt it would introduce overcomplication and it’s not a *simple* file storage server
- **Content-addressable storage (CAS)**
    - By managing the file contents through cryptographic hash values of binary data, it achieves
        - Reducing duplicated binary data because chunks from different files but same contents are managed as a same data (this works best with file versioning).
        - Fast data retrieval owing to overall less data in the system and simpler file path parsing by calculating the hash and retrieve the content
    - Discarded this idea because implementing CAS requires a lot more works (like counting the reference to a certain chunk and garbage collect them),so I didn’t think it’s achievable within the deadline. Also the challenge requirement doesn’t mention about file versioning (CAS is best for it), so it feels too much for a simple file storage server.
- **Byte-level re-uploading**
    - Although the approach I took this time handles failed chunks properly by re-uploading the whole chunk, the total amount of re-uploaded bytes can be reduced more by checking exactly until what bytes the file server was able to write. For example, with the current implemtation, if a server fails to read and write 95% of the file chunk, the CLI is required to upload the whole chunk rather than the rest of 5%.
    - This approach would make the performance of the file server by reducing the number of data being processed. However, it’s a little questionable about the data integrity of the failed chunk data, and also the atomicity of the trasaction is discussable.