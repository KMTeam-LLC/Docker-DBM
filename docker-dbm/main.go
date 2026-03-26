// Docker-DBM (Docker DataBase Manager)
// An ephemeral init container for provisioning databases in Docker Compose environments.
//
// This tool checks if a requested database and user already exist on a centralized
// database server. If they do, it exits with an error (aborting the deployment).
// If they do not, it creates them with strict multi-tenant isolation.
//
// Powered by KMTeam LLC
package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/KMTeam-LLC/Docker-DBM/docker-dbm/provisioner"
)

const (
	// defaultRetryAttempts is the number of times to retry connecting to the database
	defaultRetryAttempts = 5
	// defaultRetryDelay is the time to wait between retry attempts
	defaultRetryDelay = 2 * time.Second
)

func main() {
	log.Println("========================================")
	log.Println("  Docker-DBM - Database Manager")
	log.Println("  Powered by KMTeam LLC")
	log.Println("========================================")

	// Load configuration from environment variables
	config, err := provisioner.LoadConfigFromEnv()
	if err != nil {
		log.Printf("ERROR: Configuration error: %v", err)
		os.Exit(1)
	}

	log.Printf("Database Type: %s", config.DBType)
	log.Printf("Target Database: %s", config.AppDBName)
	log.Printf("Target User: %s", config.AppDBUser)
	log.Printf("Server: %s:%s", config.DBHost, config.DBPort)

	// Get the appropriate provisioner
	prov, err := provisioner.GetProvisioner(config.DBType)
	if err != nil {
		log.Printf("ERROR: %v", err)
		os.Exit(1)
	}

	log.Printf("Using %s provisioner", prov.Name())
	log.Println("----------------------------------------")

	// Get retry configuration from environment (optional)
	retryAttempts := getEnvInt("DB_CONNECT_RETRIES", defaultRetryAttempts)
	retryDelay := time.Duration(getEnvInt("DB_CONNECT_RETRY_DELAY", int(defaultRetryDelay.Seconds()))) * time.Second

	// Execute provisioning with retry logic for cold start scenarios
	var provisionErr error
	for attempt := 1; attempt <= retryAttempts; attempt++ {
		provisionErr = prov.Provision(config)
		if provisionErr == nil {
			break
		}

		// Check if this is a connection error that might resolve with a retry
		if isConnectionError(provisionErr) && attempt < retryAttempts {
			log.Printf("Connection attempt %d/%d failed: %v", attempt, retryAttempts, provisionErr)
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
			continue
		}

		// Non-connection error or last attempt - don't retry
		break
	}

	if provisionErr != nil {
		log.Println("----------------------------------------")
		log.Printf("PROVISIONING FAILED: %v", provisionErr)
		log.Println("Deployment aborted to prevent conflicts.")
		os.Exit(1)
	}

	log.Println("----------------------------------------")
	log.Println("SUCCESS: Database provisioning complete!")
	log.Println("The application container may now start.")
	os.Exit(0)
}

// getEnvInt returns an environment variable as an integer, or a default value if not set.
func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// isConnectionError checks if the error is likely a database connection error
// that might be resolved by retrying (e.g., during container cold start).
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	// Check for common connection-related error patterns
	connectionPatterns := []string{
		"connection refused",
		"no such host",
		"i/o timeout",
		"network is unreachable",
		"failed to connect",
		"failed to ping",
		"dial tcp",
	}
	for _, pattern := range connectionPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	return false
}
