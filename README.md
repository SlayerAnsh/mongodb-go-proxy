# MongoDB Go Proxy

A high-performance REST API proxy for MongoDB operations built with Go and Echo framework. This proxy provides a secure, HTTP-based interface to interact with MongoDB databases, supporting both RESTful endpoints and **MongoDB's deprecated REST API-compatible endpoints** (Data API format).

## Features

- 🔐 **API Key Authentication** - Secure access with separate read-only and write API keys
- 🚀 **Dynamic Connection Management** - Efficient MongoDB connection pooling with automatic cleanup
- 📚 **Dual API Support**:
  - RESTful MongoDB operations (`/api/v1/databases`)
  - **MongoDB Deprecated REST API Compatible** - Full compatibility with MongoDB's deprecated REST API format (`/api/v1/data-api/action`), perfect for migrating from the deprecated MongoDB REST API
- 📖 **Swagger Documentation** - Interactive API documentation
- 🐳 **Docker Support** - Easy deployment with Docker Compose
- ⚡ **High Performance** - Optimized for concurrent requests
- 🔍 **Health Checks** - Built-in health monitoring

## Quick Start

### Prerequisites

- Go 1.21 or higher
- MongoDB instance (local or remote)
- Docker and Docker Compose (optional)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd mongodb-go-proxy
```

2. Install dependencies:
```bash
make deps
```

3. Create a `.env` file:
```bash
MONGO_URI=mongodb://localhost:27017
API_SECRET=your-secret-key-here
READONLY_API_SECRET=your-readonly-secret-here
PORT=8080
```

4. Generate Swagger documentation:
```bash
make swagger
```

5. Run the server:
```bash
make run
# or
go run main.go
```

The server will start on `http://localhost:8080`

## Configuration

The application uses environment variables for configuration. You can set them in a `.env` file or as environment variables:

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `MONGO_URI` | MongoDB connection URI | Yes | - |
| `API_SECRET` | API key for full access (read/write) | Yes | - |
| `READONLY_API_SECRET` | API key for read-only access | No | - |
| `PORT` | Server port | No | `8080` |
| `SWAGGER_HOST` | Host for Swagger documentation | No | `localhost:8080` |

### MongoDB URI Examples

- Local MongoDB: `mongodb://localhost:27017`
- MongoDB with authentication: `mongodb://username:password@host:27017/database`
- MongoDB Atlas: `mongodb+srv://username:password@cluster.mongodb.net/database`

## API Endpoints

### Health Check

```http
GET /api/health
```

Returns the health status of the API.

### RESTful MongoDB API (`/api/v1/databases`)

#### List Databases
```http
GET /api/v1/databases
Header: api-key: <your-api-key>
```

#### List Collections
```http
GET /api/v1/databases/{database}/collections
Header: api-key: <your-api-key>
```

#### Find Documents
```http
GET /api/v1/databases/{database}/collections/{collection}/documents?limit=10&skip=0&filter={...}
Header: api-key: <your-api-key>
```

#### Get Document by ID
```http
GET /api/v1/databases/{database}/collections/{collection}/documents/{id}
Header: api-key: <your-api-key>
```

#### Find One Document
```http
GET /api/v1/databases/{database}/collections/{collection}/document?filter={...}
Header: api-key: <your-api-key>
```

#### Insert Document
```http
POST /api/v1/databases/{database}/collections/{collection}/documents
Header: api-key: <your-api-key>
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com"
}
```

#### Update Document
```http
PUT /api/v1/databases/{database}/collections/{collection}/documents/{id}
Header: api-key: <your-api-key>
Content-Type: application/json

{
  "name": "Jane Doe"
}
```

#### Delete Document
```http
DELETE /api/v1/databases/{database}/collections/{collection}/documents/{id}
Header: api-key: <your-api-key>
```

### MongoDB Data API (`/api/v1/data-api/action`)

> **⚠️ Important: MongoDB Deprecated REST API Compatibility**
> 
> This endpoint provides **full compatibility with MongoDB's deprecated REST API**. If you're currently using MongoDB's deprecated REST API or the `mongo-rest-client` npm package, you can seamlessly migrate to this proxy without changing your client code. The API format, request/response structures, and action names are identical to the deprecated MongoDB REST API.

The Data API endpoint is compatible with:
- MongoDB's deprecated REST API format
- `mongo-rest-client` npm package
- Any existing code that was built for MongoDB's deprecated REST API

This makes it an ideal drop-in replacement for applications that were using MongoDB's deprecated REST API.

#### Insert One
```http
POST /api/v1/data-api/action/insertOne
Header: api-key: <your-api-key>
Content-Type: application/json

{
  "database": "mydb",
  "collection": "users",
  "document": {
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

#### Insert Many
```http
POST /api/v1/data-api/action/insertMany
Header: api-key: <your-api-key>
Content-Type: application/json

{
  "database": "mydb",
  "collection": "users",
  "documents": [
    {"name": "John"},
    {"name": "Jane"}
  ]
}
```

#### Find One
```http
POST /api/v1/data-api/action/findOne
Header: api-key: <your-api-key>
Content-Type: application/json

{
  "database": "mydb",
  "collection": "users",
  "filter": {"name": "John"},
  "sort": {"name": 1},
  "projection": {"name": 1, "email": 1}
}
```

#### Find
```http
POST /api/v1/data-api/action/find
Header: api-key: <your-api-key>
Content-Type: application/json

{
  "database": "mydb",
  "collection": "users",
  "filter": {"status": "active"},
  "sort": {"name": 1},
  "limit": 100,
  "skip": 0,
  "projection": {"name": 1, "email": 1}
}
```

#### Update One
```http
POST /api/v1/data-api/action/updateOne
Header: api-key: <your-api-key>
Content-Type: application/json

{
  "database": "mydb",
  "collection": "users",
  "filter": {"_id": "507f1f77bcf86cd799439011"},
  "update": {"$set": {"name": "Jane Doe"}}
}
```

#### Update Many
```http
POST /api/v1/data-api/action/updateMany
Header: api-key: <your-api-key>
Content-Type: application/json

{
  "database": "mydb",
  "collection": "users",
  "filter": {"status": "active"},
  "update": {"$set": {"status": "inactive"}}
}
```

#### Delete One
```http
POST /api/v1/data-api/action/deleteOne
Header: api-key: <your-api-key>
Content-Type: application/json

{
  "database": "mydb",
  "collection": "users",
  "filter": {"_id": "507f1f77bcf86cd799439011"}
}
```

#### Delete Many
```http
POST /api/v1/data-api/action/deleteMany
Header: api-key: <your-api-key>
Content-Type: application/json

{
  "database": "mydb",
  "collection": "users",
  "filter": {"status": "deleted"}
}
```

## Migration from MongoDB Deprecated REST API

If you're currently using MongoDB's deprecated REST API, this proxy provides a seamless migration path:

1. **No Code Changes Required**: The `/api/v1/data-api/action` endpoint uses the exact same format as MongoDB's deprecated REST API
2. **Drop-in Replacement**: Simply point your existing client code to this proxy's endpoint
3. **Same Action Names**: All actions (`insertOne`, `findOne`, `find`, `updateOne`, `updateMany`, `deleteOne`, `deleteMany`) work identically
4. **Compatible with `mongo-rest-client`**: The npm package `mongo-rest-client` works out of the box with this proxy

### Example Migration

**Before (MongoDB Deprecated REST API):**
```javascript
const client = new MongoClient({
  apiKey: 'your-api-key',
  baseUrl: 'https://data.mongodb-api.com/app/your-app/endpoint/data/v1'
});
```

**After (MongoDB Go Proxy):**
```javascript
const client = new MongoClient({
  apiKey: 'your-api-key',
  baseUrl: 'http://your-proxy-server:8080/api/v1/data-api'
});
```

That's it! No other changes needed.

## Authentication

The API uses API key authentication via the `api-key` header:

- **Read Operations**: Accept both `API_SECRET` and `READONLY_API_SECRET`
- **Write Operations**: Only accept `API_SECRET` (read-only keys are rejected)

### Example

```bash
curl -H "api-key: your-secret-key" \
     http://localhost:8080/api/v1/databases
```

## Swagger Documentation

Once the server is running, access the interactive Swagger documentation at:

```
http://localhost:8080/swagger/index.html
```

## Docker Deployment

### Build and Run with Docker Compose

1. Create a `.env` file with your configuration:
```bash
MONGO_URI=mongodb://mongodb:27017
API_SECRET=your-secret-key
READONLY_API_SECRET=your-readonly-key
PORT=8080
```

2. Build and start the containers:
```bash
make docker-run
```

3. View logs:
```bash
make docker-logs
```

4. Stop containers:
```bash
make docker-down
```

### Available Docker Commands

- `make docker-build` - Build Docker image
- `make docker-up` - Start containers
- `make docker-down` - Stop containers
- `make docker-logs` - View container logs
- `make docker-restart` - Restart containers
- `make docker-clean` - Stop and remove containers and volumes
- `make docker-shell` - Access container shell

## Development

### Project Structure

```
mongodb-go-proxy/
├── config/          # Configuration management
├── database/         # MongoDB client and connection management
├── handlers/         # HTTP request handlers
│   ├── data_api.go  # MongoDB Data API handlers
│   └── mongo.go     # RESTful MongoDB handlers
├── middleware/       # Authentication middleware
├── docs/            # Swagger documentation (generated)
├── tools/           # Stress testing tools
├── main.go          # Application entry point
├── Dockerfile       # Docker image definition
└── docker-compose.yml
```

### Building

```bash
# Build the application
make build

# The binary will be in bin/server
./bin/server
```

### Running Tests

```bash
# Run stress tests (see tools/README.md for details)
make stress-test
```

### Code Generation

```bash
# Generate Swagger documentation
make swagger
```

## Connection Management

The proxy implements intelligent connection management:

- **Lazy Connection**: MongoDB connection is established on first use
- **Automatic Cleanup**: Idle connections are automatically closed after 5 minutes
- **Thread-Safe**: Safe for concurrent use
- **Connection Pooling**: Efficient connection reuse

## Performance

The proxy is optimized for high-concurrency scenarios:

- Efficient connection pooling
- Automatic connection cleanup
- Support for concurrent requests
- Built-in stress testing tools

See `tools/README.md` for stress testing instructions.

## Security Considerations

1. **API Keys**: Use strong, randomly generated API keys
2. **HTTPS**: In production, use HTTPS/TLS to encrypt traffic
3. **CORS**: Configure CORS origins appropriately (currently allows all origins)
4. **MongoDB Authentication**: Always use authenticated MongoDB connections
5. **Network Security**: Restrict network access to the proxy and MongoDB

## Troubleshooting

### Connection Issues

If you're having trouble connecting to MongoDB:

1. Verify `MONGO_URI` is correct
2. Check MongoDB is running and accessible
3. Verify network connectivity
4. Check MongoDB authentication credentials

### Authentication Errors

- Ensure the `api-key` header is included in requests
- Verify the API key matches your configured `API_SECRET` or `READONLY_API_SECRET`
- For write operations, ensure you're using `API_SECRET` (not read-only key)

### Port Conflicts

If port 8080 is already in use:

1. Change the `PORT` environment variable
2. Update `docker-compose.yml` port mapping if using Docker

## License

[Add your license here]

## Contributing

[Add contribution guidelines here]

## Support

For issues and questions, please open an issue on the repository.

