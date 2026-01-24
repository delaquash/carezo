package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/delaquash/carezo/configs"
	"github.com/delaquash/carezo/internal/database"
	"github.com/delaquash/carezo/internal/handlers"
	"github.com/delaquash/carezo/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// load configs
	cfg := configs.LoadConfig()
	log.Printf("Starting %s in %s mode", cfg.AppName, cfg.AppEnv)

	// connecting to postgresql
	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatalf("Failedto connect to Postgresql: %v", err)
	}

	defer database.ClosePostgres()

	// connect to redis
	redis, err := database.ConnectRedis(cfg)

	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}

	defer database.CloseRedis()

	// Set gin mode
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// gin router
	router := gin.Default()

	// middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.RateLimitMiddleware(cfg))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H {
			"status": "healthy",
			"time":    time.Now(),
			"db":	   db.Ping() == nil,
			"redis":   redis.Ping(context.Background()).Err() == nil,
		})
	})

	// routes
	api := router.Group("/api") 
	{
		authHandler := handlers.NewAuthHandler(cfg)
		auth := api.Group("/auth") 
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/verify-otp", authHandler.VerifyOTP)
			auth.POST("/complete-profile", authHandler.CompleteProfile)
			auth.POST("/login", authHandler.Login)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// protected route
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			// User Profile Routes
			protected.GET("/me", func(c *gin.Context) {
				userID := c.GetString("user_id")
				c.JSON(http.StatusOK, gin.H{"user_id": userID})
			})

			// admin route
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole("admin"))
			{
				admin.GET("/users", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Admin users list"})
				})
			}
		}
	}

	// creating http server
	srv := &http.Server {
		Addr: 			":" + cfg.AppPort,
		Handler: 		router,
		ReadTimeout: 	15 * time.Second,
		WriteTimeout: 	15 * time.Second,
		IdleTimeout: 	60 * time.Second,
	}

	go func ()  {
		log.Printf("Server listening on port %s", cfg.AppPort)
		log.Printf("API Documentation: http://localhost:%s/health", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// shutting down
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}