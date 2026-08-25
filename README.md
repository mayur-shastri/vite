# FileVault Backend

Backend service for **FileVault**, a secure file-management system built with Go, GraphQL, PostgreSQL, and local content-addressed storage.

The backend provides authentication, file and folder management, deduplicated storage, permissions, sharing, tagging, search, and file downloads through a GraphQL API.

## Features

* **JWT authentication**

  * User registration and login
  * 24-hour JWT tokens
  * Bearer-token authentication middleware

* **File management**

  * Upload files
  * Rename, move, and delete files
  * Download files through an authenticated HTTP endpoint
  * MIME type and file-size metadata

* **Folder management**

  * Create nested folders
  * Rename, move, and delete folders
  * Navigate resources using a unified resource hierarchy

* **Content deduplication**

  * Files are identified using SHA-256 hashes
  * Identical files share a single physical copy on disk
  * Reference counting manages shared physical files
  * Storage statistics expose the space saved through deduplication

* **Permissions**

  * Resource-level access control
  * `VIEWER` and `EDITOR` roles
  * Grant and revoke access using user email addresses

* **Public sharing**

  * Generate shareable resource tokens
  * Make files/folders publicly accessible
  * Resolve resources through share tokens

* **Tags**

  * Add tags to files and folders
  * Remove tags
  * Search resources using tags

* **Search**

  * Search by name
  * Filter by resource type
  * Filter by MIME type
  * Filter by file size
  * Filter by creation date
  * Filter by tags
  * Paginated results

* **Resource hierarchy**

  * Files and folders are represented using a unified `Resource` model
  * PostgreSQL closure-table ancestry tracking enables efficient hierarchy operations

* **API protection**

  * CORS configuration
  * Request rate limiting
  * Authorization checks at the service layer

## Tech Stack

| Component              | Technology                     |
| ---------------------- | ------------------------------ |
| Language               | Go 1.25.1                      |
| API                    | GraphQL                        |
| GraphQL implementation | gqlgen                         |
| HTTP router            | chi                            |
| Database               | PostgreSQL                     |
| ORM                    | GORM                           |
| Authentication         | JWT                            |
| Password hashing       | `golang.org/x/crypto`          |
| File storage           | Local filesystem               |
| Rate limiting          | Tollbooth                      |
| IDs / tokens           | UUID                           |
| Configuration          | Environment variables / `.env` |

## Architecture

The backend follows a layered architecture:

```text
                    ┌──────────────────────┐
                    │      GraphQL API      │
                    │      gqlgen/chi       │
                    └──────────┬───────────┘
                               │
                    ┌──────────▼───────────┐
                    │      Resolvers        │
                    └──────────┬───────────┘
                               │
          ┌────────────────────▼────────────────────┐
          │                 Services                 │
          │                                          │
          │ User │ File │ Folder │ Permission │ ... │
          └────────────────────┬────────────────────┘
                               │
                    ┌──────────▼───────────┐
                    │     Repositories      │
                    └──────────┬───────────┘
                               │
                 ┌─────────────┴─────────────┐
                 │                           │
        ┌────────▼────────┐        ┌─────────▼────────┐
        │   PostgreSQL    │        │  File Storage    │
        │     / GORM      │        │   Filesystem     │
        └─────────────────┘        └──────────────────┘
```

### Project Structure

```text
server/
├── cmd/
│   └── server/
│       └── main.go
│
├── graph/
│   ├── generated/
│   ├── model/
│   ├── resolver.go
│   ├── schema.graphqls
│   └── schema.resolvers.go
│
├── internal/
│   ├── database/
│   │   ├── database.go
│   │   └── models.go
│   ├── file/
│   ├── folders/
│   ├── middleware/
│   ├── permission/
│   ├── search/
│   ├── share/
│   ├── tag/
│   └── user/
│
└── pkg/
    ├── auth/
    └── utils/
```

Each major domain is separated into a repository and service layer. Services contain business logic and authorization rules, while repositories handle persistence.

## Storage Model

FileVault separates a user's **logical file resource** from the **physical file stored on disk**.

```text
Resource
   │
   └── PhysicalFile
          │
          └── actual file on disk
```

When a file is uploaded, the backend calculates its SHA-256 hash.

The hash determines the physical storage location:

```text
<storage-root>/
├── ab/
│   └── cd/
│       └── abcdef...
```

If another user uploads the exact same file:

1. The SHA-256 hash is calculated.
2. The existing `PhysicalFile` is found.
3. No second copy is written to disk.
4. Its reference count is incremented.
5. A new logical `Resource` is created for the user.

This allows multiple resources to point to the same physical file.

When a resource is deleted, the physical file's reference count is decremented. Once the count reaches zero, the physical file and its database record are removed.

## Resource Hierarchy

Files and folders use a common `resources` table.

Each resource contains:

```text
Resource
├── id
├── owner
├── parent
├── name
├── type
├── isPublic
├── shareToken
└── physicalFile (files only)
```

Folders can contain both files and other folders.

The backend also maintains a `resource_ancestors` closure table:

```text
Ancestor → Descendant → Depth
```

Database triggers maintain this table when resources are inserted or moved, allowing hierarchy queries without recursively traversing the entire tree.

## Authentication

Authentication uses JWTs.

Register or login through GraphQL:

```graphql
mutation {
  login(
    email: "user@example.com"
    password: "password"
  ) {
    token
    user {
      id
      username
      email
    }
  }
}
```

Use the returned token in subsequent requests:

```http
Authorization: Bearer <JWT_TOKEN>
```

The authentication middleware validates the token and places the authenticated user's ID into the request context.

Some GraphQL operations, such as registration and login, can be accessed without authentication. Protected operations require a valid authenticated context.

## GraphQL API

The GraphQL endpoint is:

```text
POST /query
```

The GraphQL Playground is available at:

```text
GET /
```

when the server is running.

### Queries

Available queries include:

```graphql
me
file(id: ID!)
folder(id: ID!)
resolveShareLink(token: String!, expectedType: String!)
resources(folderId: ID)
searchResources(filters: SearchFilters!, offset: Int, limit: Int)
allResources
```

### Mutations

Authentication:

```graphql
register(username: String!, email: String!, password: String!)
login(email: String!, password: String!)
```

Files:

```graphql
uploadFile(file: Upload!, parentId: ID)
renameFile(id: ID!, newName: String!)
deleteFile(id: ID!)
moveFile(fileId: ID!, newParentId: ID)
```

Folders:

```graphql
createFolder(name: String!, parentId: ID)
renameFolder(id: ID!, newName: String!)
deleteFolder(id: ID!)
moveFolder(folderId: ID!, newParentId: ID)
```

Permissions:

```graphql
grantPermission(resourceId: ID!, email: String!, role: Role!)
revokePermission(resourceId: ID!, email: String!)
```

Tags:

```graphql
addTagToResource(resourceID: ID!, tagName: String!)
removeTagFromResource(resourceID: ID!, tagID: ID!)
```

Public sharing:

```graphql
makeResourcePublic(resourceId: ID!)
removeResourcePublicAccess(resourceId: ID!)
```

## File Downloads

Downloads use a dedicated HTTP endpoint:

```text
GET /download/{resourceID}
```

The endpoint is protected by the same authentication middleware and checks resource permissions before opening the physical file.

The original filename, MIME type, and file size are retained as metadata.

## Search

Search is exposed through the `searchResources` query.

Example:

```graphql
query {
  searchResources(
    filters: {
      name: "report"
      types: ["file"]
      mimeTypes: ["application/pdf"]
      minSizeBytes: 1000
      maxSizeBytes: 10000000
      tags: ["important"]
    }
    offset: 0
    limit: 25
  ) {
    id
    name
    type
    createdAt
  }
}
```

Search queries are scoped to the authenticated user's resources.

## Permissions

Resources support two explicit permission levels:

```text
VIEWER
EDITOR
```

The resource owner retains ownership regardless of these permissions.

Permissions are stored as resource-user ACL entries:

```text
Resource ────── Permission ────── User
                   │
                   └── VIEWER / EDITOR
```

Owners can grant and revoke permissions using the user's email address.

## Configuration

Create a `.env` file inside `server/`:

```env
DATABASE_URL=postgres://username:password@localhost:5432/filevault
JWT_SECRET_KEY=your-secret-key
FILEVAULT_STORAGE_PATH=./uploads
```

Optional:

```env
PORT=4007
```

If `PORT` is not specified, the server defaults to:

```text
4007
```

If `FILEVAULT_STORAGE_PATH` is not specified, the server defaults to:

```text
./testuploads
```

## Requirements

* Go 1.25.1+
* PostgreSQL
* A configured PostgreSQL database
* A sufficiently strong JWT secret
* Write access to the configured file storage directory

## Running Locally

### 1. Clone the repository

```bash
git clone <repository-url>
cd balkan-backend/server
```

### 2. Configure environment variables

```bash
cp .env.example .env
```

Fill in:

```env
DATABASE_URL=...
JWT_SECRET_KEY=...
FILEVAULT_STORAGE_PATH=...
```

### 3. Install dependencies

```bash
go mod download
```

### 4. Prepare the database

Create the PostgreSQL database specified by `DATABASE_URL`.

The project includes GORM models and an `Migrate` function for schema creation. Database migration is currently disabled in `main.go`, so an existing schema must be available before starting the application.

The `dbTriggers.txt` file contains the PostgreSQL triggers required to maintain the resource ancestry closure table.

### 5. Start the server

```bash
go run ./cmd/server
```

The server will start on:

```text
http://localhost:4007
```

Open the GraphQL Playground at:

```text
http://localhost:4007/
```

## Generating GraphQL Code

The project uses gqlgen.

After changing `graph/schema.graphqls`, regenerate the generated GraphQL code using:

```bash
go run github.com/99designs/gqlgen generate
```

The gqlgen configuration is located at:

```text
server/gqlgen.yml
```

## Database Model

The primary database entities are:

```text
User
 │
 ├── Resource
 │     ├── File metadata → PhysicalFile
 │     ├── Parent/Child hierarchy
 │     ├── Permissions
 │     └── Tags
 │
 └── Permissions

PhysicalFile
 └── ReferenceCount

ResourceAncestor
 └── Closure table for hierarchy traversal

Tag
 └── Many-to-many relationship with Resource
```

## Rate Limiting

The HTTP server uses Tollbooth to limit incoming requests.

The current limiter allows approximately **2 requests per second per client**, with an expirable limiter configuration.

Requests exceeding the configured limit receive:

```http
429 Too Many Requests
```

## Security Considerations

The backend currently implements:

* Password hashing
* JWT-based authentication
* Bearer-token validation
* Resource-level authorization
* Permission checks
* Owner-only destructive operations
* Authenticated downloads
* Request rate limiting
* Unique share tokens
* SHA-256 content hashing

For production deployment, the following should additionally be considered:

* Store secrets in a dedicated secret manager.
* Restrict CORS origins instead of allowing `*`.
* Use HTTPS.
* Configure appropriate upload-size limits.
* Add production database migrations.
* Add stronger validation around resource moves and hierarchy cycles.
* Add comprehensive automated tests.
* Add structured logging and monitoring.
* Consider object storage such as S3-compatible storage for horizontally scaled deployments.

## Development Notes

The application is initialized using dependency injection in `cmd/server/main.go`.

The dependency flow is approximately:

```text
Repository
    ↓
Service
    ↓
Resolver
    ↓
GraphQL
```

For example, file handling is wired as:

```text
file.Repository
       ↓
file.Service
       ↓
GraphQL Resolver
```

This keeps persistence logic separate from business rules and API concerns.

## License

This project is currently intended as a FileVault backend implementation and does not specify a public license.
