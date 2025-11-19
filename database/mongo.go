package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// ConnectionTimeout is the idle timeout before closing a connection
	ConnectionTimeout = 5 * time.Minute
	// ConnectionCheckInterval is how often to check for stale connections
	ConnectionCheckInterval = 1 * time.Minute
)

// Client wraps the MongoDB client with dynamic connection management
type Client struct {
	uri          string
	client       *mongo.Client
	lastUsed     time.Time
	mu           sync.RWMutex
	connectionMu sync.Mutex // Protects connection creation to prevent race conditions
	stopCleanup  chan struct{}
	cleanupMu    sync.Mutex // Protects cleanup goroutine lifecycle
}

// NewClient creates a new MongoDB client with dynamic connection management
// The connection will be established lazily on first use
func NewClient(uri string) (*Client, error) {
	client := &Client{
		uri: uri,
	}

	return client, nil
}

// ensureConnection checks if connection exists and creates a new one if needed
// This method is thread-safe and prevents multiple goroutines from creating connections simultaneously
func (c *Client) ensureConnection(ctx context.Context) error {
	// First, check if we have a valid connection without locking for creation
	c.mu.RLock()
	hasConnection := c.client != nil
	c.mu.RUnlock()

	if hasConnection {
		// Update last used time and return existing connection
		c.mu.Lock()
		if c.client != nil {
			c.lastUsed = time.Now()
			c.mu.Unlock()
			return nil
		}
		c.mu.Unlock()
	}

	// Use connectionMu to ensure only one goroutine creates a connection at a time
	c.connectionMu.Lock()
	defer c.connectionMu.Unlock()

	// Double-check after acquiring the lock (another goroutine might have created it)
	c.mu.RLock()
	if c.client != nil {
		c.mu.RUnlock()
		c.mu.Lock()
		c.lastUsed = time.Now()
		c.mu.Unlock()
		return nil
	}
	c.mu.RUnlock()

	// Create new connection
	clientOptions := options.Client().ApplyURI(c.uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	} else {
		log.Println("Connected to MongoDB")
	}

	// Update state with new connection
	c.mu.Lock()
	c.client = client
	c.lastUsed = time.Now()
	c.mu.Unlock()

	// Start cleanup goroutine for this connection
	c.startCleanup()

	return nil
}

// GetConnection ensures a valid connection and returns the client
func (c *Client) GetConnection(ctx context.Context) (*mongo.Client, error) {
	if err := c.ensureConnection(ctx); err != nil {
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client, nil
}

// startCleanup starts the cleanup goroutine if not already running
func (c *Client) startCleanup() {
	c.cleanupMu.Lock()
	defer c.cleanupMu.Unlock()

	// Only start if we don't have a cleanup goroutine running
	if c.stopCleanup == nil {
		c.stopCleanup = make(chan struct{})
		go c.cleanupStaleConnections()
	}
}

// stopCleanup stops the cleanup goroutine
func (c *Client) stopCleanupGoroutine() {
	c.cleanupMu.Lock()
	defer c.cleanupMu.Unlock()

	if c.stopCleanup != nil {
		close(c.stopCleanup)
		c.stopCleanup = nil
	}
}

// cleanupStaleConnections periodically checks and closes stale connections
func (c *Client) cleanupStaleConnections() {
	ticker := time.NewTicker(ConnectionCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Checking for stale connections")
			c.mu.Lock()
			timeSinceLastUse := time.Since(c.lastUsed)
			hasConnection := c.client != nil

			if hasConnection && timeSinceLastUse > ConnectionTimeout {
				// Connection is stale, close it
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				if c.client != nil {
					c.client.Disconnect(ctx)
					log.Println("Disconnected from MongoDB")
				}
				cancel()
				c.client = nil
				c.mu.Unlock()

				// Stop cleanup goroutine since connection is closed
				c.stopCleanupGoroutine()
				return
			}
			c.mu.Unlock()

		case <-c.stopCleanup:
			return
		}
	}
}

// GetClient returns the MongoDB client (deprecated, use GetConnection instead)
func (c *Client) GetClient() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := c.GetConnection(ctx)
	if err != nil {
		return nil
	}
	return client
}

// ListDatabases returns a list of all database names
func (c *Client) ListDatabases(ctx context.Context) ([]string, error) {
	client, err := c.GetConnection(ctx)
	if err != nil {
		return nil, err
	}

	databases, err := client.ListDatabaseNames(ctx, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}
	return databases, nil
}

// ListCollections returns a list of collection names in the specified database
func (c *Client) ListCollections(ctx context.Context, dbName string) ([]string, error) {
	client, err := c.GetConnection(ctx)
	if err != nil {
		return nil, err
	}

	db := client.Database(dbName)
	collections, err := db.ListCollectionNames(ctx, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	return collections, nil
}

// GetCollection returns a collection from the specified database
func (c *Client) GetCollection(dbName, collectionName string) (*mongo.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := c.GetConnection(ctx)
	if err != nil {
		return nil, err
	}

	db := client.Database(dbName)
	return db.Collection(collectionName), nil
}

// Close closes the MongoDB connection and stops cleanup goroutine
func (c *Client) Close(ctx context.Context) error {
	// Stop cleanup goroutine
	c.stopCleanupGoroutine()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		return c.client.Disconnect(ctx)
	}
	return nil
}

// Ping checks the connection to MongoDB
func (c *Client) Ping(ctx context.Context) error {
	client, err := c.GetConnection(ctx)
	if err != nil {
		return err
	}
	return client.Ping(ctx, nil)
}
