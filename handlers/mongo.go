package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mongodb-go-proxy/database"
)

// MongoHandler handles MongoDB proxy operations
type MongoHandler struct {
	dbClient *database.Client
}

// NewMongoHandler creates a new MongoDB handler
func NewMongoHandler(dbClient *database.Client) *MongoHandler {
	return &MongoHandler{
		dbClient: dbClient,
	}
}

// Response structs for Swagger documentation

// ListDatabasesResponse represents the response for listing databases
type ListDatabasesResponse struct {
	Databases []string `json:"databases" example:"[\"mydb\",\"testdb\"]"` // List of database names
	Count     int      `json:"count" example:"2"`                         // Number of databases
}

// ListCollectionsResponse represents the response for listing collections
type ListCollectionsResponse struct {
	Database    string   `json:"database" example:"mydb"`                     // Database name
	Collections []string `json:"collections" example:"[\"users\",\"posts\"]"` // List of collection names
	Count       int      `json:"count" example:"2"`                           // Number of collections
}

// FindDocumentsResponse represents the response for finding documents
type FindDocumentsResponse struct {
	Database   string                   `json:"database" example:"mydb"`              // Database name
	Collection string                   `json:"collection" example:"users"`           // Collection name
	Documents  []map[string]interface{} `json:"documents" swaggertype:"array,object"` // Array of found documents
	Count      int                      `json:"count" example:"10"`                   // Number of documents returned
	TotalCount int64                    `json:"total_count" example:"100"`            // Total number of documents matching the filter
}

// FindOneDocumentResponse represents the response for finding one document
type FindOneDocumentResponse struct {
	Database   string                 `json:"database" example:"mydb"`       // Database name
	Collection string                 `json:"collection" example:"users"`    // Collection name
	Document   map[string]interface{} `json:"document" swaggertype:"object"` // The found document
}

// InsertDocumentResponse represents the response for inserting a document
type InsertDocumentResponse struct {
	Database   string                 `json:"database" example:"mydb"`                        // Database name
	Collection string                 `json:"collection" example:"users"`                     // Collection name
	InsertedID string                 `json:"inserted_id" example:"507f1f77bcf86cd799439011"` // The ID of the inserted document
	Document   map[string]interface{} `json:"document" swaggertype:"object"`                  // The inserted document
}

// UpdateDocumentResponse represents the response for updating a document
type UpdateDocumentResponse struct {
	Database      string `json:"database" example:"mydb"`                        // Database name
	Collection    string `json:"collection" example:"users"`                     // Collection name
	DocumentID    string `json:"document_id" example:"507f1f77bcf86cd799439011"` // Document ID
	MatchedCount  int64  `json:"matched_count" example:"1"`                      // Number of documents matched
	ModifiedCount int64  `json:"modified_count" example:"1"`                     // Number of documents modified
}

// DeleteDocumentResponse represents the response for deleting a document
type DeleteDocumentResponse struct {
	Database     string `json:"database" example:"mydb"`                        // Database name
	Collection   string `json:"collection" example:"users"`                     // Collection name
	DocumentID   string `json:"document_id" example:"507f1f77bcf86cd799439011"` // Document ID
	DeletedCount int64  `json:"deleted_count" example:"1"`                      // Number of documents deleted
}

// ListDatabases godoc
//
//	@Summary		List all databases
//	@Description	Returns a list of all database names
//	@Tags			databases
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	ListDatabasesResponse	"Successfully retrieved database list"
//	@Failure		401	{object}	map[string]string		"Unauthorized - missing or invalid api-key"
//	@Failure		500	{object}	map[string]string		"Internal server error"
//	@Router			/v1/databases [get]
func (h *MongoHandler) ListDatabases(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	databases, err := h.dbClient.ListDatabases(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"databases": databases,
		"count":     len(databases),
	})
}

// ListCollections godoc
//
//	@Summary		List collections in a database
//	@Description	Returns a list of all collection names in the specified database
//	@Tags			collections
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			db	path		string					true	"Database name"	example("mydb")
//	@Success		200	{object}	ListCollectionsResponse	"Successfully retrieved collection list"
//	@Failure		400	{object}	map[string]string		"Bad request - invalid database name"
//	@Failure		401	{object}	map[string]string		"Unauthorized - missing or invalid api-key"
//	@Failure		500	{object}	map[string]string		"Internal server error"
//	@Router			/v1/databases/{db}/collections [get]
func (h *MongoHandler) ListCollections(c echo.Context) error {
	dbName := c.Param("db")
	if dbName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Database name is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collections, err := h.dbClient.ListCollections(ctx, dbName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"database":    dbName,
		"collections": collections,
		"count":       len(collections),
	})
}

// FindDocuments godoc
//
//	@Summary		Find documents in a collection
//	@Description	Query documents from a collection with optional filter, limit, and skip
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			db			path		string					true	"Database name"					example("mydb")
//	@Param			collection	path		string					true	"Collection name"				example("users")
//	@Param			filter		query		string					false	"MongoDB filter (JSON string)"	example("{\"name\":\"John\"}")
//	@Param			limit		query		int						false	"Limit number of results"		default(100)	example(100)
//	@Param			skip		query		int						false	"Skip number of results"		default(0)		example(0)
//	@Param			sort		query		string					false	"Sort criteria (JSON string)"	example("{\"name\":1}")
//	@Success		200			{object}	FindDocumentsResponse	"Successfully retrieved documents"
//	@Failure		400			{object}	map[string]string		"Bad request - invalid filter, sort, limit, or skip"
//	@Failure		401			{object}	map[string]string		"Unauthorized - missing or invalid api-key"
//	@Failure		500			{object}	map[string]string		"Internal server error"
//	@Router			/v1/databases/{db}/collections/{collection}/documents [get]
func (h *MongoHandler) FindDocuments(c echo.Context) error {
	dbName := c.Param("db")
	collectionName := c.Param("collection")

	if dbName == "" || collectionName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Database and collection names are required",
		})
	}

	collection, err := h.dbClient.GetCollection(dbName, collectionName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get collection: " + err.Error(),
		})
	}

	// Parse query parameters
	filterStr := c.QueryParam("filter")
	limit := int64(100)
	skip := int64(0)
	sortStr := c.QueryParam("sort")

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := parseInt64(l); err == nil {
			limit = parsed
		}
	}
	if s := c.QueryParam("skip"); s != "" {
		if parsed, err := parseInt64(s); err == nil {
			skip = parsed
		}
	}

	// Build filter
	var filter bson.M
	if filterStr != "" {
		if err := bson.UnmarshalExtJSON([]byte(filterStr), true, &filter); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid filter JSON: " + err.Error(),
			})
		}
	} else {
		filter = bson.M{}
	}

	// Build sort
	var sort bson.D
	if sortStr != "" {
		if err := bson.UnmarshalExtJSON([]byte(sortStr), true, &sort); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid sort JSON: " + err.Error(),
			})
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build find options
	findOptions := options.Find().SetLimit(limit).SetSkip(skip)
	if len(sort) > 0 {
		findOptions.SetSort(sort)
	}

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Get total count
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"database":    dbName,
		"collection":  collectionName,
		"documents":   results,
		"count":       len(results),
		"total_count": count,
	})
}

// FindOne godoc
//
//	@Summary		Find one document in a collection
//	@Description	Query a single document from a collection with optional filter and sort
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			db			path		string					true	"Database name"					example("mydb")
//	@Param			collection	path		string					true	"Collection name"				example("users")
//	@Param			filter		query		string					false	"MongoDB filter (JSON string)"	example("{\"name\":\"John\"}")
//	@Param			sort		query		string					false	"Sort criteria (JSON string)"	example("{\"name\":1}")
//	@Success		200			{object}	FindOneDocumentResponse	"Successfully retrieved document"
//	@Failure		400			{object}	map[string]string		"Bad request - invalid filter or sort"
//	@Failure		401			{object}	map[string]string		"Unauthorized - missing or invalid api-key"
//	@Failure		404			{object}	map[string]string		"Not found - document not found"
//	@Failure		500			{object}	map[string]string		"Internal server error"
//	@Router			/v1/databases/{db}/collections/{collection}/document [get]
func (h *MongoHandler) FindOne(c echo.Context) error {
	dbName := c.Param("db")
	collectionName := c.Param("collection")

	if dbName == "" || collectionName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Database and collection names are required",
		})
	}

	collection, err := h.dbClient.GetCollection(dbName, collectionName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get collection: " + err.Error(),
		})
	}

	// Parse query parameters
	filterStr := c.QueryParam("filter")
	sortStr := c.QueryParam("sort")

	// Build filter
	var filter bson.M
	if filterStr != "" {
		if err := bson.UnmarshalExtJSON([]byte(filterStr), true, &filter); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid filter JSON: " + err.Error(),
			})
		}
	} else {
		filter = bson.M{}
	}

	// Build sort
	var sort bson.D
	if sortStr != "" {
		if err := bson.UnmarshalExtJSON([]byte(sortStr), true, &sort); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid sort JSON: " + err.Error(),
			})
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build find options
	findOptions := options.FindOne()
	if len(sort) > 0 {
		findOptions.SetSort(sort)
	}

	var result bson.M
	err = collection.FindOne(ctx, filter, findOptions).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Document not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"database":   dbName,
		"collection": collectionName,
		"document":   result,
	})
}

// InsertDocument godoc
//
//	@Summary		Insert a document
//	@Description	Insert a new document into a collection
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			db			path		string					true	"Database name"				example("mydb")
//	@Param			collection	path		string					true	"Collection name"			example("users")
//	@Param			document	body		object					true	"Document to insert (JSON)"	example({"name":"John","age":30})
//	@Success		201			{object}	InsertDocumentResponse	"Successfully inserted document"
//	@Failure		400			{object}	map[string]string		"Bad request - invalid JSON body"
//	@Failure		401			{object}	map[string]string		"Unauthorized - missing or invalid api-key"
//	@Failure		500			{object}	map[string]string		"Internal server error"
//	@Router			/v1/databases/{db}/collections/{collection}/documents [post]
func (h *MongoHandler) InsertDocument(c echo.Context) error {
	dbName := c.Param("db")
	collectionName := c.Param("collection")

	if dbName == "" || collectionName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Database and collection names are required",
		})
	}

	var document bson.M
	if err := c.Bind(&document); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON body: " + err.Error(),
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection, err := h.dbClient.GetCollection(dbName, collectionName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get collection: " + err.Error(),
		})
	}

	result, err := collection.InsertOne(ctx, document)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"database":    dbName,
		"collection":  collectionName,
		"inserted_id": result.InsertedID,
		"document":    document,
	})
}

// UpdateDocument godoc
//
//	@Summary		Update a document
//	@Description	Update a document by ID
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			db			path		string					true	"Database name"				example("mydb")
//	@Param			collection	path		string					true	"Collection name"			example("users")
//	@Param			id			path		string					true	"Document ID"				example("507f1f77bcf86cd799439011")
//	@Param			document	body		object					true	"Update document (JSON)"	example({"name":"Jane","age":31})
//	@Success		200			{object}	UpdateDocumentResponse	"Successfully updated document"
//	@Failure		400			{object}	map[string]string		"Bad request - invalid document ID or JSON body"
//	@Failure		401			{object}	map[string]string		"Unauthorized - missing or invalid api-key"
//	@Failure		404			{object}	map[string]string		"Not found - document not found"
//	@Failure		500			{object}	map[string]string		"Internal server error"
//	@Router			/v1/databases/{db}/collections/{collection}/documents/{id} [put]
func (h *MongoHandler) UpdateDocument(c echo.Context) error {
	dbName := c.Param("db")
	collectionName := c.Param("collection")
	docID := c.Param("id")

	if dbName == "" || collectionName == "" || docID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Database, collection, and document ID are required",
		})
	}

	objectID, err := primitive.ObjectIDFromHex(docID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid document ID: " + err.Error(),
		})
	}

	var updateDoc bson.M
	if err := c.Bind(&updateDoc); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON body: " + err.Error(),
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection, err := h.dbClient.GetCollection(dbName, collectionName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get collection: " + err.Error(),
		})
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": updateDoc}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	if result.MatchedCount == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Document not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"database":       dbName,
		"collection":     collectionName,
		"document_id":    docID,
		"matched_count":  result.MatchedCount,
		"modified_count": result.ModifiedCount,
	})
}

// DeleteDocument godoc
//
//	@Summary		Delete a document
//	@Description	Delete a document by ID
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			db			path		string					true	"Database name"		example("mydb")
//	@Param			collection	path		string					true	"Collection name"	example("users")
//	@Param			id			path		string					true	"Document ID"		example("507f1f77bcf86cd799439011")
//	@Success		200			{object}	DeleteDocumentResponse	"Successfully deleted document"
//	@Failure		400			{object}	map[string]string		"Bad request - invalid document ID"
//	@Failure		401			{object}	map[string]string		"Unauthorized - missing or invalid api-key"
//	@Failure		404			{object}	map[string]string		"Not found - document not found"
//	@Failure		500			{object}	map[string]string		"Internal server error"
//	@Router			/v1/databases/{db}/collections/{collection}/documents/{id} [delete]
func (h *MongoHandler) DeleteDocument(c echo.Context) error {
	dbName := c.Param("db")
	collectionName := c.Param("collection")
	docID := c.Param("id")

	if dbName == "" || collectionName == "" || docID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Database, collection, and document ID are required",
		})
	}

	objectID, err := primitive.ObjectIDFromHex(docID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid document ID: " + err.Error(),
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection, err := h.dbClient.GetCollection(dbName, collectionName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get collection: " + err.Error(),
		})
	}

	filter := bson.M{"_id": objectID}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	if result.DeletedCount == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Document not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"database":      dbName,
		"collection":    collectionName,
		"document_id":   docID,
		"deleted_count": result.DeletedCount,
	})
}

// GetDocument godoc
//
//	@Summary		Get a document by ID
//	@Description	Retrieve a single document by its ID
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			db			path		string					true	"Database name"		example("mydb")
//	@Param			collection	path		string					true	"Collection name"	example("users")
//	@Param			id			path		string					true	"Document ID"		example("507f1f77bcf86cd799439011")
//	@Success		200			{object}	map[string]interface{}	"Successfully retrieved document"
//	@Failure		400			{object}	map[string]string		"Bad request - invalid document ID"
//	@Failure		401			{object}	map[string]string		"Unauthorized - missing or invalid api-key"
//	@Failure		404			{object}	map[string]string		"Not found - document not found"
//	@Failure		500			{object}	map[string]string		"Internal server error"
//	@Router			/v1/databases/{db}/collections/{collection}/documents/{id} [get]
func (h *MongoHandler) GetDocument(c echo.Context) error {
	dbName := c.Param("db")
	collectionName := c.Param("collection")
	docID := c.Param("id")

	if dbName == "" || collectionName == "" || docID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Database, collection, and document ID are required",
		})
	}

	objectID, err := primitive.ObjectIDFromHex(docID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid document ID: " + err.Error(),
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection, err := h.dbClient.GetCollection(dbName, collectionName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get collection: " + err.Error(),
		})
	}

	var result bson.M
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Document not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}

// Helper function to parse int64
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
