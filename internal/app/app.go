// @title Finance Tracker API
// @version 1.0
// @description API for managing users, accounts, and transactions
// @host localhost:8080
// @BasePath /

package app

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"finance-tracker/db/queries"
	_ "finance-tracker/docs"
	"finance-tracker/pkg/handler"
	"finance-tracker/pkg/repository"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Run() {
	// DATABASE CONNECTION STRING
	dbURL := getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5435/finance_tracker?sslmode=disable")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		log.Fatal("Database unreachable:", err)
	}

	log.Println("Connected to PostgreSQL")

	q := queries.New(db) // ✅ db implements DBTX

	// REPOSITORIES
	userRepo := repository.NewUserRepository(q)
	accountRepo := repository.NewAccountRepository(q)
	transactionRepo := repository.NewTransactionRepository(q)

	// HANDLERs
	userHandler := handler.NewUserHandler(userRepo)
	accountHandler := handler.NewAccountHandler(accountRepo)
	transactionHandler := handler.NewTransactionHandler(transactionRepo)

	// GIN ROUTER
	router := gin.Default()

	// Swagger UI
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(302, "/docs/index.html")
	})
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API ROUTES
	api := router.Group("")
	{
		// User
		api.POST("/register", userHandler.Register)

		api.GET("/users", userHandler.List)

		api.GET("/users/:id", userHandler.GetByID)

		api.PUT("/users/:id", userHandler.Update)

		api.DELETE("/users/:id", userHandler.Delete)

		// Accounts
		api.POST("/accounts", accountHandler.Create)

		// Transactions
		api.GET("/transactions", transactionHandler.List)
	}

	// START SERVER
	port := getenv("PORT", "8080")

	log.Println("Server running on port", port)
	router.Run(":" + port)
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
