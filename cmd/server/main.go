package main

import (
	"embed"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"stripe-go-spike/internal/api"
	dbpkg "stripe-go-spike/internal/db"
	"stripe-go-spike/internal/payments"
)

//go:embed frontend
var frontendAssets embed.FS

func main() {
	// Load environment variables from .env files if present
	loadDotenv()

	// Optionally run database migrations
	if shouldRunMigrations() {
		if err := runMigrations(); err != nil {
			log.Fatalf("migrations failed: %v", err)
		}
		log.Printf("migrations applied successfully")
	}

	// Initialize the payments service (mock for spike)
	payService := payments.NewService(payments.Config{
		SecretKey:      os.Getenv("STRIPE_SECRET_KEY"),
		PublishableKey: os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		WebhookSecret:  os.Getenv("STRIPE_WEBHOOK_SECRET"),
	})

	// Create and run the Gin server
	router := api.NewRouter(payService, frontendAssets)

	addr := getServerAddr()

	log.Printf("Starting server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

// loadDotenv loads environment variables from .env files with sensible precedence.
// OS environment variables always take precedence over file values.
func loadDotenv() {
	// Determine environment mode
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("GO_ENV")
	}

	// Load highest-precedence files first, using godotenv.Load which does not overwrite existing env
	var files []string
	if env != "" {
		files = append(files, ".env."+env+".local")
	}
	files = append(files, ".env.local")
	if env != "" {
		files = append(files, ".env."+env)
	}
	files = append(files, ".env")

	for _, f := range files {
		_ = godotenv.Load(f)
	}
}

func shouldRunMigrations() bool {
	v := os.Getenv("RUN_MIGRATION")
	if v == "" {
		return false
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		// any non-empty, non-boolean value will be treated as true as a convenience
		return true
	}
	return b
}

func runMigrations() error {
	// Connect to the database using the same logic as the migrate command
	database, err := dbpkg.NewConnection()
	if err != nil {
		return err
	}
	defer database.Close()

	mfs := os.DirFS("db/migrations")
	return dbpkg.RunMigrations(database, mfs, "")
}

func getServerAddr() string {
	addr := os.Getenv("APP_ADDR")
	if addr == "" {
		addr = ":8060"
	}

	port := os.Getenv("PORT")
	if port != "" {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			// addr is just a port, so we can ignore the error and use an empty host
			host = ""
		}
		addr = net.JoinHostPort(host, port)
	}
	return addr
}
