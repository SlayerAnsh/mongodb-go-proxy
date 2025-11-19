package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mongodb-go-proxy/database"
)

// DataAPIHandler handles MongoDB Data API format requests
type DataAPIHandler struct {
	dbClient *database.Client
}

// NewDataAPIHandler creates a new Data API handler
func NewDataAPIHandler(dbClient *database.Client) *DataAPIHandler {
	return &DataAPIHandler{
		dbClient: dbClient,
	}
}

// Base request fields shared by all actions
type baseRequest struct {
	Database   string `json:"database" example:"mydb"`    // Database name (required)
	Collection string `json:"collection" example:"users"` // Collection name (required)
}

// InsertOneRequest represents the request for insertOne action
//
//	@Description	Request body for insertOne action. Document is a MongoDB document object.
type InsertOneRequest struct {
	baseRequest
	Document map[string]interface{} `json:"document" swaggertype:"object"` // Document to insert (required). Example: {"name":"John","age":30}
}

// InsertManyRequest represents the request for insertMany action
//
//	@Description	Request body for insertMany action. Documents is an array of MongoDB document objects.
type InsertManyRequest struct {
	baseRequest
	Documents []map[string]interface{} `json:"documents" swaggertype:"array,object"` // Array of documents to insert (required). Example: [{"name":"John"},{"name":"Jane"}]
}

// FindOneRequest represents the request for findOne action
//
//	@Description	Request body for findOne action. Filter, sort, and projection are MongoDB query objects.
type FindOneRequest struct {
	baseRequest
	Filter     interface{} `json:"filter,omitempty" swaggertype:"object"`     // MongoDB filter query (optional). Example: {"name":"John"}
	Sort       interface{} `json:"sort,omitempty" swaggertype:"object"`       // Sort criteria (optional). Example: {"name":1}
	Projection interface{} `json:"projection,omitempty" swaggertype:"object"` // Fields to include/exclude (optional). Example: {"name":1,"age":1}
}

// FindRequest represents the request for find action
//
//	@Description	Request body for find action. Filter, sort, and projection are MongoDB query objects.
type FindRequest struct {
	baseRequest
	Filter     interface{} `json:"filter,omitempty" swaggertype:"object"`     // MongoDB filter query (optional). Example: {"name":"John"}
	Sort       interface{} `json:"sort,omitempty" swaggertype:"object"`       // Sort criteria (optional). Example: {"name":1}
	Limit      *int64      `json:"limit,omitempty" example:"100"`             // Maximum number of documents to return (optional, default: 100)
	Skip       *int64      `json:"skip,omitempty" example:"0"`                // Number of documents to skip (optional, default: 0)
	Projection interface{} `json:"projection,omitempty" swaggertype:"object"` // Fields to include/exclude (optional). Example: {"name":1,"age":1}
}

// UpdateOneRequest represents the request for updateOne action
//
//	@Description	Request body for updateOne action. Filter is a MongoDB query object. Update is a MongoDB update document (use $set, $unset, etc.).
type UpdateOneRequest struct {
	baseRequest
	Filter interface{} `json:"filter" swaggertype:"object"` // MongoDB filter query (required). Example: {"_id":"507f1f77bcf86cd799439011"}
	Update interface{} `json:"update" swaggertype:"object"` // Update document (required). Example: {"$set":{"name":"Jane"}}
}

// UpdateManyRequest represents the request for updateMany action
//
//	@Description	Request body for updateMany action. Filter is a MongoDB query object. Update is a MongoDB update document (use $set, $unset, etc.).
type UpdateManyRequest struct {
	baseRequest
	Filter interface{} `json:"filter" swaggertype:"object"` // MongoDB filter query (required). Example: {"status":"active"}
	Update interface{} `json:"update" swaggertype:"object"` // Update document (required). Example: {"$set":{"status":"inactive"}}
}

// DeleteOneRequest represents the request for deleteOne action
//
//	@Description	Request body for deleteOne action. Filter is a MongoDB query object.
type DeleteOneRequest struct {
	baseRequest
	Filter interface{} `json:"filter" swaggertype:"object"` // MongoDB filter query (required). Example: {"_id":"507f1f77bcf86cd799439011"}
}

// DeleteManyRequest represents the request for deleteMany action
//
//	@Description	Request body for deleteMany action. Filter is a MongoDB query object.
type DeleteManyRequest struct {
	baseRequest
	Filter interface{} `json:"filter" swaggertype:"object"` // MongoDB filter query (required). Example: {"status":"deleted"}
}

// Response structs for Swagger documentation

// InsertOneResponse represents the response for insertOne action
type InsertOneResponse struct {
	InsertedID string `json:"insertedId" example:"507f1f77bcf86cd799439011"` // The ID of the inserted document
}

// InsertManyResponse represents the response for insertMany action
type InsertManyResponse struct {
	InsertedIDs []string `json:"insertedIds" example:"[\"507f1f77bcf86cd799439011\",\"507f1f77bcf86cd799439012\"]"` // Array of IDs of inserted documents
}

// FindOneResponse represents the response for findOne action
type FindOneResponse struct {
	Document map[string]interface{} `json:"document" swaggertype:"object"` // The found document, or null if not found
}

// FindResponse represents the response for find action
type FindResponse struct {
	Documents  []map[string]interface{} `json:"documents" swaggertype:"array,object"` // Array of found documents
	Count      int                      `json:"count" example:"10"`                   // Number of documents returned
	TotalCount *int64                   `json:"totalCount,omitempty" example:"100"`   // Total number of documents matching the filter (optional)
	Skip       *int64                   `json:"skip,omitempty" example:"0"`           // Number of documents skipped (optional)
	Limit      *int64                   `json:"limit,omitempty" example:"100"`        // Maximum number of documents returned (optional)
}

// UpdateOneResponse represents the response for updateOne action
type UpdateOneResponse struct {
	MatchedCount  int64  `json:"matchedCount" example:"1"`                                // Number of documents matched
	ModifiedCount int64  `json:"modifiedCount" example:"1"`                               // Number of documents modified
	UpsertedID    string `json:"upsertedId,omitempty" example:"507f1f77bcf86cd799439011"` // ID of upserted document (if upsert occurred)
}

// UpdateManyResponse represents the response for updateMany action
type UpdateManyResponse struct {
	MatchedCount  int64  `json:"matchedCount" example:"5"`                                // Number of documents matched
	ModifiedCount int64  `json:"modifiedCount" example:"5"`                               // Number of documents modified
	UpsertedID    string `json:"upsertedId,omitempty" example:"507f1f77bcf86cd799439011"` // ID of upserted document (if upsert occurred)
}

// DeleteOneResponse represents the response for deleteOne action
type DeleteOneResponse struct {
	DeletedCount int64 `json:"deletedCount" example:"1"` // Number of documents deleted (0 or 1)
}

// DeleteManyResponse represents the response for deleteMany action
type DeleteManyResponse struct {
	DeletedCount int64 `json:"deletedCount" example:"5"` // Number of documents deleted
}

// InsertOne godoc
//
//	@Summary		Insert a single document
//	@Description	Inserts a single document into the specified collection
//	@Tags			data-api
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body		InsertOneRequest	true	"Insert one document request"
//	@Success		200		{object}	InsertOneResponse	"Successfully inserted document"
//	@Failure		400		{object}	map[string]string	"Bad request - missing required fields or invalid JSON"
//	@Failure		401		{object}	map[string]string	"Unauthorized - missing or invalid api-key"
//	@Failure		403		{object}	map[string]string	"Forbidden - invalid credentials"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Router			/v1/data-api/action/insertOne [post]
func (h *DataAPIHandler) InsertOne(c echo.Context) error {
	var req InsertOneRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	if req.Database == "" || req.Collection == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "database and collection are required",
		})
	}

	if req.Document == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "document is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection, err := h.dbClient.GetCollection(req.Database, req.Collection)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get collection: " + err.Error(),
		})
	}

	// Convert document to bson.M
	docBytes, err := bson.Marshal(req.Document)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid document: " + err.Error(),
		})
	}

	var doc bson.M
	if err := bson.Unmarshal(docBytes, &doc); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid document format: " + err.Error(),
		})
	}

	result, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Convert ObjectID to string for JSON response
	insertedID := result.InsertedID
	if oid, ok := insertedID.(primitive.ObjectID); ok {
		insertedID = oid.Hex()
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"insertedId": insertedID,
	})
}

// InsertMany godoc
//
//	@Summary		Insert multiple documents
//	@Description	Inserts multiple documents into the specified collection
//	@Tags			data-api
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body		InsertManyRequest	true	"Insert many documents request"
//	@Success		200		{object}	InsertManyResponse	"Successfully inserted documents"
//	@Failure		400		{object}	map[string]string	"Bad request - missing required fields or invalid JSON"
//	@Failure		401		{object}	map[string]string	"Unauthorized - missing or invalid api-key"
//	@Failure		403		{object}	map[string]string	"Forbidden - invalid credentials"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Router			/v1/data-api/action/insertMany [post]
func (h *DataAPIHandler) InsertMany(c echo.Context) error {
	var req InsertManyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	if req.Database == "" || req.Collection == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "database and collection are required",
		})
	}

	if len(req.Documents) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "documents array is required and cannot be empty",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection, err := h.dbClient.GetCollection(req.Database, req.Collection)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get collection: " + err.Error(),
		})
	}

	var docs []interface{}
	for _, doc := range req.Documents {
		docBytes, err := bson.Marshal(doc)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid document: " + err.Error(),
			})
		}

		var bsonDoc bson.M
		if err := bson.Unmarshal(docBytes, &bsonDoc); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid document format: " + err.Error(),
			})
		}
		docs = append(docs, bsonDoc)
	}

	result, err := collection.InsertMany(ctx, docs)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Convert ObjectIDs to strings
	insertedIds := make([]interface{}, len(result.InsertedIDs))
	for i, id := range result.InsertedIDs {
		if oid, ok := id.(primitive.ObjectID); ok {
			insertedIds[i] = oid.Hex()
		} else {
			insertedIds[i] = id
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"insertedIds": insertedIds,
	})
}

// FindOne godoc
//
//	@Summary		Find a single document
//	@Description	Finds a single document matching the filter criteria
//	@Tags			data-api
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body		FindOneRequest		true	"Find one document request"
//	@Success		200		{object}	FindOneResponse		"Successfully found document"
//	@Failure		400		{object}	map[string]string	"Bad request - invalid filter, sort, or projection"
//	@Failure		401		{object}	map[string]string	"Unauthorized - missing or invalid api-key"
//	@Failure		403		{object}	map[string]string	"Forbidden - invalid credentials"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Router			/v1/data-api/action/findOne [post]
func (h *DataAPIHandler) FindOne(c echo.Context) error {
	var req FindOneRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	if req.Database == "" || req.Collection == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "database and collection are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection, err := h.dbClient.GetCollection(req.Database, req.Collection)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get collection: " + err.Error(),
		})
	}

	filter, err := h.buildFilter(req.Filter)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid filter: " + err.Error(),
		})
	}

	findOptions := options.FindOne()
	if req.Sort != nil {
		sort, err := h.buildSort(req.Sort)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid sort: " + err.Error(),
			})
		}
		if len(sort) > 0 {
			findOptions.SetSort(sort)
		}
	}

	// Add projection support
	if req.Projection != nil {
		projection, err := h.buildProjection(req.Projection)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid projection: " + err.Error(),
			})
		}
		if projection != nil {
			findOptions.SetProjection(projection)
		}
	}

	var result bson.M
	err = collection.FindOne(ctx, filter, findOptions).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"document": nil,
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"document": result,
	})
}

// Find godoc
//
//	@Summary		Find multiple documents
//	@Description	Finds multiple documents matching the filter criteria with pagination support
//	@Tags			data-api
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body		FindRequest			true	"Find documents request"
//	@Success		200		{object}	FindResponse		"Successfully found documents"
//	@Failure		400		{object}	map[string]string	"Bad request - invalid filter, sort, limit, skip, or projection"
//	@Failure		401		{object}	map[string]string	"Unauthorized - missing or invalid api-key"
//	@Failure		403		{object}	map[string]string	"Forbidden - invalid credentials"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Router			/v1/data-api/action/find [post]
func (h *DataAPIHandler) Find(c echo.Context) error {
	var req FindRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	if req.Database == "" || req.Collection == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "database and collection are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection, err := h.dbClient.GetCollection(req.Database, req.Collection)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get collection: " + err.Error(),
		})
	}

	filter, err := h.buildFilter(req.Filter)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid filter: " + err.Error(),
		})
	}

	findOptions := options.Find()

	// Add limit
	if req.Limit != nil && *req.Limit > 0 {
		findOptions.SetLimit(*req.Limit)
	}

	// Add skip
	if req.Skip != nil && *req.Skip > 0 {
		findOptions.SetSkip(*req.Skip)
	}

	// Add sort support
	if req.Sort != nil {
		sort, err := h.buildSort(req.Sort)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid sort: " + err.Error(),
			})
		}
		if len(sort) > 0 {
			findOptions.SetSort(sort)
		}
	}

	// Add projection support
	if req.Projection != nil {
		projection, err := h.buildProjection(req.Projection)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid projection: " + err.Error(),
			})
		}
		if projection != nil {
			findOptions.SetProjection(projection)
		}
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

	response := map[string]interface{}{
		"documents": results,
		"count":     len(results),
	}
	if req.Skip != nil {
		response["skip"] = *req.Skip
	}
	if req.Limit != nil {
		response["limit"] = *req.Limit
	}

	// Get total count for the filter (for pagination info)
	totalCount, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		// If count fails, still return documents but without totalCount
		return c.JSON(http.StatusOK, response)
	}

	response["totalCount"] = totalCount

	return c.JSON(http.StatusOK, response)
}

// UpdateOne godoc
//
//	@Summary		Update a single document
//	@Description	Updates a single document matching the filter criteria
//	@Tags			data-api
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body		UpdateOneRequest	true	"Update one document request"
//	@Success		200		{object}	UpdateOneResponse	"Successfully updated document"
//	@Failure		400		{object}	map[string]string	"Bad request - missing required fields or invalid JSON"
//	@Failure		401		{object}	map[string]string	"Unauthorized - missing or invalid api-key"
//	@Failure		403		{object}	map[string]string	"Forbidden - invalid credentials"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Router			/v1/data-api/action/updateOne [post]
func (h *DataAPIHandler) UpdateOne(c echo.Context) error {
	var req UpdateOneRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	if req.Database == "" || req.Collection == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "database and collection are required",
		})
	}

	if req.Filter == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "filter is required",
		})
	}

	if req.Update == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "update is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection, err := h.dbClient.GetCollection(req.Database, req.Collection)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get collection: " + err.Error(),
		})
	}

	filter, err := h.buildFilter(req.Filter)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid filter: " + err.Error(),
		})
	}

	update, err := h.buildUpdate(req.Update)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid update: " + err.Error(),
		})
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	response := map[string]interface{}{
		"matchedCount":  result.MatchedCount,
		"modifiedCount": result.ModifiedCount,
	}

	// Add upsertedId if document was upserted
	if result.UpsertedID != nil {
		upsertedID := result.UpsertedID
		if oid, ok := upsertedID.(primitive.ObjectID); ok {
			upsertedID = oid.Hex()
		}
		response["upsertedId"] = upsertedID
	}

	return c.JSON(http.StatusOK, response)
}

// UpdateMany godoc
//
//	@Summary		Update multiple documents
//	@Description	Updates multiple documents matching the filter criteria
//	@Tags			data-api
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body		UpdateManyRequest	true	"Update many documents request"
//	@Success		200		{object}	UpdateManyResponse	"Successfully updated documents"
//	@Failure		400		{object}	map[string]string	"Bad request - missing required fields or invalid JSON"
//	@Failure		401		{object}	map[string]string	"Unauthorized - missing or invalid api-key"
//	@Failure		403		{object}	map[string]string	"Forbidden - invalid credentials"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Router			/v1/data-api/action/updateMany [post]
func (h *DataAPIHandler) UpdateMany(c echo.Context) error {
	var req UpdateManyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	if req.Database == "" || req.Collection == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "database and collection are required",
		})
	}

	if req.Filter == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "filter is required",
		})
	}

	if req.Update == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "update is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection, err := h.dbClient.GetCollection(req.Database, req.Collection)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get collection: " + err.Error(),
		})
	}

	filter, err := h.buildFilter(req.Filter)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid filter: " + err.Error(),
		})
	}

	update, err := h.buildUpdate(req.Update)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid update: " + err.Error(),
		})
	}

	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	response := map[string]interface{}{
		"matchedCount":  result.MatchedCount,
		"modifiedCount": result.ModifiedCount,
	}

	// Add upsertedId if document was upserted
	if result.UpsertedID != nil {
		upsertedID := result.UpsertedID
		if oid, ok := upsertedID.(primitive.ObjectID); ok {
			upsertedID = oid.Hex()
		}
		response["upsertedId"] = upsertedID
	}

	return c.JSON(http.StatusOK, response)
}

// DeleteOne godoc
//
//	@Summary		Delete a single document
//	@Description	Deletes a single document matching the filter criteria
//	@Tags			data-api
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body		DeleteOneRequest	true	"Delete one document request"
//	@Success		200		{object}	DeleteOneResponse	"Successfully deleted document"
//	@Failure		400		{object}	map[string]string	"Bad request - missing required fields or invalid JSON"
//	@Failure		401		{object}	map[string]string	"Unauthorized - missing or invalid api-key"
//	@Failure		403		{object}	map[string]string	"Forbidden - invalid credentials"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Router			/v1/data-api/action/deleteOne [post]
func (h *DataAPIHandler) DeleteOne(c echo.Context) error {
	var req DeleteOneRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	if req.Database == "" || req.Collection == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "database and collection are required",
		})
	}

	if req.Filter == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "filter is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection, err := h.dbClient.GetCollection(req.Database, req.Collection)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get collection: " + err.Error(),
		})
	}

	filter, err := h.buildFilter(req.Filter)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid filter: " + err.Error(),
		})
	}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"deletedCount": result.DeletedCount,
	})
}

// DeleteMany godoc
//
//	@Summary		Delete multiple documents
//	@Description	Deletes multiple documents matching the filter criteria
//	@Tags			data-api
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body		DeleteManyRequest	true	"Delete many documents request"
//	@Success		200		{object}	DeleteManyResponse	"Successfully deleted documents"
//	@Failure		400		{object}	map[string]string	"Bad request - missing required fields or invalid JSON"
//	@Failure		401		{object}	map[string]string	"Unauthorized - missing or invalid api-key"
//	@Failure		403		{object}	map[string]string	"Forbidden - invalid credentials"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Router			/v1/data-api/action/deleteMany [post]
func (h *DataAPIHandler) DeleteMany(c echo.Context) error {
	var req DeleteManyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	if req.Database == "" || req.Collection == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "database and collection are required",
		})
	}

	if req.Filter == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "filter is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection, err := h.dbClient.GetCollection(req.Database, req.Collection)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get collection: " + err.Error(),
		})
	}

	filter, err := h.buildFilter(req.Filter)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid filter: " + err.Error(),
		})
	}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"deletedCount": result.DeletedCount,
	})
}

// Helper functions to build MongoDB query objects

func (h *DataAPIHandler) buildFilter(filter interface{}) (bson.M, error) {
	if filter == nil {
		return bson.M{}, nil
	}

	filterBytes, err := bson.Marshal(filter)
	if err != nil {
		return nil, err
	}

	var result bson.M
	if err := bson.Unmarshal(filterBytes, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (h *DataAPIHandler) buildSort(sort interface{}) (bson.D, error) {
	if sort == nil {
		return bson.D{}, nil
	}

	sortBytes, err := bson.Marshal(sort)
	if err != nil {
		return nil, err
	}

	var result bson.D
	if err := bson.Unmarshal(sortBytes, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (h *DataAPIHandler) buildUpdate(update interface{}) (bson.M, error) {
	if update == nil {
		return nil, nil
	}

	updateBytes, err := bson.Marshal(update)
	if err != nil {
		return nil, err
	}

	var result bson.M
	if err := bson.Unmarshal(updateBytes, &result); err != nil {
		return nil, err
	}

	// If update doesn't have operators like $set, $unset, etc., wrap it in $set
	if !hasUpdateOperators(result) {
		return bson.M{"$set": result}, nil
	}

	return result, nil
}

// hasUpdateOperators checks if the update document contains MongoDB update operators
func hasUpdateOperators(update bson.M) bool {
	for key := range update {
		if len(key) > 0 && key[0] == '$' {
			return true
		}
	}
	return false
}

// buildProjection builds a projection document from the request
func (h *DataAPIHandler) buildProjection(projection interface{}) (bson.M, error) {
	if projection == nil {
		return nil, nil
	}

	projectionBytes, err := bson.Marshal(projection)
	if err != nil {
		return nil, err
	}

	var result bson.M
	if err := bson.Unmarshal(projectionBytes, &result); err != nil {
		return nil, err
	}

	return result, nil
}
