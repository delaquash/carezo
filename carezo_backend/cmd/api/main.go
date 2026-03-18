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
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
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
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now(),
			"db":     db.Ping() == nil,
			"redis":  redis.Ping(context.Background()).Err() == nil,
		})
	})

	// handlers
	authHandler := handlers.NewAuthHandler(cfg)
	carHandler := handlers.NewCarHandler()
	driverHandler := handlers.NewDriverHandler()
	userHandler := handlers.NewUserHandler()
	bookingHandler := handlers.NewBookingHandler()

	// routes
	api := router.Group("/api")
	{
		// ================= AUTH =================
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/verify-otp", authHandler.VerifyOTP)
			auth.POST("/resend-otp", authHandler.ResendOTP)
			auth.POST("/login", authHandler.Login)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// ================= PUBLIC CARS =================
		cars := api.Group("/cars")
		{
			cars.GET("", carHandler.ListAllCars)
			cars.GET("/search", carHandler.SearchCars)
			cars.GET("/available", carHandler.GetAvailableCars)
			cars.GET("/:id", carHandler.GetCar)
		}

		// ================= PUBLIC DRIVERS =================
		drivers := api.Group("/drivers")
		{
			drivers.GET("", driverHandler.ListAllDrivers)
			drivers.GET("/search", driverHandler.SearchDrivers)
			drivers.GET("/:id", driverHandler.GetDriver)
			drivers.GET("/:id/reviews", driverHandler.GetDriverReviews)
		}

		// ================= PROTECTED =================
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			// ✅ MOVED HERE (correct placement)
			protected.GET("/me", func(c *gin.Context) {
				userID := c.GetString("user_id")
				userEmail := c.GetString("user_email")
				c.JSON(http.StatusOK, gin.H{
					"user_id": userID,
					"email":   userEmail,
				})
			})

			user := protected.Group("/user")
			{
				user.GET("/get-profile", userHandler.GetMe)
				user.PUT("/update-profile", userHandler.UpdateProfile)
				user.PUT("/complete-profile", userHandler.CompleteUserProfile)
				user.DELETE("/delete-user", userHandler.DeleteAccount)
			}

			bookings := protected.Group("/bookings")
			{
				bookings.POST("", bookingHandler.CreateBooking)
				bookings.GET("", bookingHandler.ListUserBooking)
				bookings.GET("/:id", bookingHandler.GetBooking)
				bookings.POST("/:id/cancel", bookingHandler.CancelBooking)
			}

			protected.POST("/reviews", driverHandler.CreateReview)
		}

		// ================= ADMIN =================
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRole("admin"))
		{
			// Car management
			admin.POST("/cars", carHandler.CreateCar)
			admin.PUT("/cars/:id", carHandler.UpdateCar)
			admin.DELETE("/cars/:id", carHandler.DeleteCar)

			// Driver management
			admin.POST("/drivers", driverHandler.CreateDriver)
			admin.PUT("/drivers/:id", driverHandler.UpdateDriver)
			admin.DELETE("/drivers/:id", driverHandler.DeleteDriver)

			// User management
			adminUsers := admin.Group("/users")
			{
				adminUsers.GET("", userHandler.ListAllUsers) // ✅ correct route
				adminUsers.PUT("/:id/status", userHandler.UpdateUserStatus)
				adminUsers.GET("/get-user/:id", userHandler.GetUserByID)
			}
		}

		////////////////////////////////////////////////////////////////////////////////
		// ❌ DUPLICATE BLOCK (COMMENTED OUT)
		// Reason:
		// - Re-registers /api/admin/users → causes Gin panic
		// - /me route wrongly placed here → moved above
		////////////////////////////////////////////////////////////////////////////////

		/*
		{
			// ❌ MOVED to protected block above
			protected.GET("/me", func(c *gin.Context) {
				userID := c.GetString("user_id")
				userEmail := c.GetString("user_email")
				c.JSON(http.StatusOK, gin.H{
					"user_id": userID,
					"email": userEmail,
				})
			})

			// ❌ DUPLICATE admin group
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole("admin"))
			{
				// ❌ DUPLICATE ROUTE: /api/admin/users
				admin.GET("/users", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Admin users list"})
				})
			}
		}
		*/
	}

	// ================= SERVER =================
	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🌐 Server listening on port %s", cfg.AppPort)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// ================= GRACEFUL SHUTDOWN =================
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