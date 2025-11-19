package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	"mongodb-go-proxy/config"
	"mongodb-go-proxy/database"
	swagger_docs "mongodb-go-proxy/docs" // swagger docs
	"mongodb-go-proxy/handlers"
	auth "mongodb-go-proxy/middleware"
)

//	@title			MongoDB Go Proxy API
//	@version		1.0
//	@description	REST API proxy for MongoDB operations
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.email	support@example.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8080
//	@BasePath	/api

// @schemes					http https
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						api-key
func main() {

	// Load configuration
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	swagger_docs.SwaggerInfo.Host = config.GetEnv("SWAGGER_HOST", "localhost:8080") // ex: "api.example.com"
	log.Println("Swagger Host:", swagger_docs.SwaggerInfo.Host)
	// Initialize MongoDB client (connection will be established lazily on first use)
	dbClient, err := database.NewClient(cfg.MongoURI)
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())

	// CORS middleware
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: []string{"*"}, // In production, specify exact origins
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "api-secret", "api-key"},
	}))

	// Initialize handlers
	mongoHandler := handlers.NewMongoHandler(dbClient)
	dataAPIHandler := handlers.NewDataAPIHandler(dbClient)

	api := e.Group("/api")
	// Public routes (no auth required)
	api.GET("/health", healthCheck)
	database := api.Group("/v1/databases")
	// Setup routes with appropriate authentication
	setupMongoRoutes(database, mongoHandler, cfg.APISecret, cfg.ReadOnlyAPISecret)

	// MongoDB Data API routes (compatible with mongo-rest-client npm package)
	dataApi := api.Group("/v1/data-api")
	// MongoDB Data API routes (compatible with mongo-rest-client npm package)
	setupDataAPIRoutes(dataApi, dataAPIHandler, cfg.APISecret, cfg.ReadOnlyAPISecret)

	// Swagger documentation (no auth for easier access)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Start server
	port := ":" + cfg.ServerPort
	e.Logger.Fatal(e.Start(port))
}

// setupMongoRoutes configures all MongoDB proxy routes with appropriate authentication
func setupMongoRoutes(api *echo.Group, handler *handlers.MongoHandler, apiSecret, readOnlyAPISecret string) {
	// Read routes - accept both API_SECRET and READONLY_API_SECRET
	readRoutes := api.Group("")
	readRoutes.Use(auth.ReadAuth(apiSecret, readOnlyAPISecret))
	{
		// Database routes (read)
		readRoutes.GET("", handler.ListDatabases)

		// Collection routes (read)
		readRoutes.GET("/:db/collections", handler.ListCollections)

		// Document read routes
		readRoutes.GET("/:db/collections/:collection/documents", handler.FindDocuments)
		readRoutes.GET("/:db/collections/:collection/documents/:id", handler.GetDocument)
		readRoutes.GET("/:db/collections/:collection/document", handler.FindOne)
	}

	// Write routes - only accept API_SECRET
	writeRoutes := api.Group("")
	writeRoutes.Use(auth.WriteAuth(apiSecret))
	{
		// Document write routes
		writeRoutes.POST("/:db/collections/:collection/documents", handler.InsertDocument)
		writeRoutes.PUT("/:db/collections/:collection/documents/:id", handler.UpdateDocument)
		writeRoutes.DELETE("/:db/collections/:collection/documents/:id", handler.DeleteDocument)
	}
}

// setupDataAPIRoutes configures MongoDB Data API routes (compatible with mongo-rest-client npm package)
func setupDataAPIRoutes(api *echo.Group, handler *handlers.DataAPIHandler, apiSecret, readOnlyAPISecret string) {
	actionRoute := api.Group("/action")

	// Read actions - accept both API_SECRET and READONLY_API_SECRET
	readRoutes := actionRoute.Group("")
	readRoutes.Use(auth.ReadAuth(apiSecret, readOnlyAPISecret))
	{
		readRoutes.POST("/findOne", handler.FindOne)
		readRoutes.POST("/find", handler.Find)
	}

	// Write actions - only accept API_SECRET
	writeRoutes := actionRoute.Group("")
	writeRoutes.Use(auth.WriteAuth(apiSecret))
	{
		writeRoutes.POST("/insertOne", handler.InsertOne)
		writeRoutes.POST("/insertMany", handler.InsertMany)
		writeRoutes.POST("/updateOne", handler.UpdateOne)
		writeRoutes.POST("/updateMany", handler.UpdateMany)
		writeRoutes.POST("/deleteOne", handler.DeleteOne)
		writeRoutes.POST("/deleteMany", handler.DeleteMany)
	}
}

// healthCheck godoc
//
//	@Summary		Health check endpoint
//	@Description	Returns the health status of the API and MongoDB connection
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "API is running",
	})
}
