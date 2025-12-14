package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"todo-api/internal/config"
	"todo-api/internal/errors"
	"todo-api/internal/handler"
	authMiddleware "todo-api/internal/middleware"
	"todo-api/internal/model"
	"todo-api/internal/repository"
	"todo-api/internal/service"
	"todo-api/internal/validator"
	"todo-api/pkg/database"
)

func main() {
	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load .env file (optional, for local development)
	if err := godotenv.Load(); err != nil {
		log.Debug().Msg("No .env file found")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer database.Close(db)

	// Auto migrate models (development only)
	if cfg.IsDevelopment() {
		if err := db.AutoMigrate(
			&model.User{},
			&model.JwtDenylist{},
			&model.Category{},
			&model.Tag{},
			&model.Todo{},
			&model.TodoTag{},
			&model.Comment{},
			&model.TodoHistory{},
		); err != nil {
			log.Fatal().Err(err).Msg("Failed to auto migrate models")
		}
		log.Info().Msg("Database models migrated")
	}

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true

	// Set custom error handler
	e.HTTPErrorHandler = errors.ErrorHandler

	// Set custom validator
	validator.SetupValidator(e)

	// Middleware
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORS configuration (from environment)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.GetCORSOrigins(),
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		ExposeHeaders:    []string{echo.HeaderAuthorization},
		AllowCredentials: true,
		MaxAge:           cfg.CORSMaxAge,
	}))

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
		})
	})

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	tagRepo := repository.NewTagRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	historyRepo := repository.NewTodoHistoryRepository(db)

	// Initialize services
	todoService := service.NewTodoService(todoRepo, categoryRepo, historyRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userRepo, denylistRepo, cfg)
	todoHandler := handler.NewTodoHandler(todoService, todoRepo)
	categoryHandler := handler.NewCategoryHandler(categoryRepo)
	tagHandler := handler.NewTagHandler(tagRepo)
	commentHandler := handler.NewCommentHandler(commentRepo, todoRepo)
	historyHandler := handler.NewTodoHistoryHandler(historyRepo, todoRepo)

	// Auth routes (public)
	auth := e.Group("/auth")
	auth.POST("/sign_up", authHandler.SignUp)
	auth.POST("/sign_in", authHandler.SignIn)
	auth.DELETE("/sign_out", authHandler.SignOut, authMiddleware.JWTAuth(cfg, userRepo, denylistRepo))

	// API v1 routes (protected)
	api := e.Group("/api/v1", authMiddleware.JWTAuth(cfg, userRepo, denylistRepo))

	// Todo routes
	api.GET("/todos", todoHandler.List)
	api.GET("/todos/search", todoHandler.Search) // Must be before /todos/:id
	api.POST("/todos", todoHandler.Create)
	api.GET("/todos/:id", todoHandler.Show)
	api.PATCH("/todos/:id", todoHandler.Update)
	api.DELETE("/todos/:id", todoHandler.Delete)
	api.PATCH("/todos/update_order", todoHandler.UpdateOrder)

	// Category routes
	api.GET("/categories", categoryHandler.List)
	api.POST("/categories", categoryHandler.Create)
	api.GET("/categories/:id", categoryHandler.Show)
	api.PATCH("/categories/:id", categoryHandler.Update)
	api.DELETE("/categories/:id", categoryHandler.Delete)

	// Tag routes
	api.GET("/tags", tagHandler.List)
	api.POST("/tags", tagHandler.Create)
	api.GET("/tags/:id", tagHandler.Show)
	api.PATCH("/tags/:id", tagHandler.Update)
	api.DELETE("/tags/:id", tagHandler.Delete)

	// Comment routes (nested under todos)
	api.GET("/todos/:todo_id/comments", commentHandler.List)
	api.POST("/todos/:todo_id/comments", commentHandler.Create)
	api.PATCH("/todos/:todo_id/comments/:id", commentHandler.Update)
	api.DELETE("/todos/:todo_id/comments/:id", commentHandler.Delete)

	// History routes (nested under todos)
	api.GET("/todos/:todo_id/histories", historyHandler.List)

	// Log startup information
	log.Info().
		Str("port", cfg.Port).
		Str("env", cfg.Env).
		Msg("Starting server")

	// Start server in a goroutine
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited gracefully")
}
