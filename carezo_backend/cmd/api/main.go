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

	// handlers
	authHandler := handlers.NewAuthHandler(cfg)
	carHandler := handlers.NewCarHandler()
	driverHandler := handlers.NewDriverHandler()
	userHandler := handlers.UserHandler()
	bookingHandler := handlers.BookingHandler()

	// routes
	api := router.Group("/api") 
	{
		auth := api.Group("/auth") 
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/verify-otp", authHandler.VerifyOTP)
			auth.POST("/complete-profile", authHandler.CompleteProfile)
			auth.POST("/login", authHandler.Login)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

			// Public car routes (anyone can view cars)
		cars := api.Group("/cars")
		{
			cars.GET("", carHandler.ListAllCars)              // GET /api/cars?page=1&per_page=20
			cars.GET("/search", carHandler.SearchCars)        // GET /api/cars/search?brand=Toyota&transmission=automatic
			cars.GET("/available", carHandler.GetAvailableCars) // GET /api/cars/available?pickup_date=2024-01-15T10:00:00Z&return_date=2024-01-20T18:00:00Z
			cars.GET("/:id", carHandler.GetCar)               // GET /api/cars/{car_id}
		}


		// Public driver routes (anyone can view drivers)
		drivers := api.Group("/drivers")
		{
			drivers.GET("", driverHandler.ListAllDrivers)         // GET /api/drivers
			drivers.GET("/search", driverHandler.SearchDrivers)   // GET /api/drivers/search
			drivers.GET("/:id", driverHandler.GetDriver)          // GET /api/drivers/{id}
			drivers.GET("/:id/reviews", driverHandler.GetDriverReviews) // GET /api/drivers/{id}/reviews
		}

		// Protected routes (user must be authenticated)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{

			user := protected.Group("/user")
			{
				user.GET("/get-profile", userHandler.GetMe)
				user.PUT("/update-profile", userHandler.UpdateProfile)
				user.PUT("/complete-profile", userHandler.CompleteUserProfile)
				user.PUT("/change-password", userHandler.ChangePassword)
				user.DELETE("/delete-user", userHandler.DeactivateAcount)
			}
			// Review routes (authenticated users only)
			protected.POST("/reviews", driverHandler.CreateReview) // POST /api/reviews
		}

		// Admin routes (require admin role)
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRole("admin"))
		{
			// Car management (admin only)
			admin.POST("/cars", carHandler.CreateCar)	  
			admin.PUT("/cars/:id", carHandler.UpdateCar)   
			admin.DELETE("/cars/:id", carHandler.DeleteCar) 
				// Driver management (admin only)
			admin.POST("/drivers", driverHandler.CreateDriver)       
			admin.PUT("/drivers/:id", driverHandler.UpdateDriver)   
			admin.DELETE("/drivers/:id", driverHandler.DeleteDriver) 


			// User management (admin only)
			adminUsers := admin.Group("/users")
			{
				adminUsers.GET("", userHandler.ListUsers) 
				adminUsers.PUT("/:id/status", userHandler.UpdateUserStatus) 
				adminUsers.GET("/get-user/:id", userHandler.GetUserByID)  
			}
		
		}

		{
			// User Profile Routes
			protected.GET("/me", func(c *gin.Context) {
				userID := c.GetString("user_id")
				userEmail := c.GetString("user_email")
				c.JSON(http.StatusOK, gin.H{
					"user_id": userID,
					"email": userEmail,
				})
			})

			// admin route
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole("admin"))
			{
				// admin only
				admin.POST("/cars", carHandler.CreateCar)
				admin.PUT("/cars/:id", carHandler.UpdateCar)
				admin.DELETE("/cars/:id", carHandler.DeleteCar)
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

	go func() {
		log.Printf("🌐 Server listening on port %s", cfg.AppPort)
		log.Println("\n📝 Available Endpoints:")
		log.Println("  PUBLIC:")
		log.Println("    POST   /api/auth/register")
		log.Println("    POST   /api/auth/login")
		log.Println("    GET    /api/cars")
		log.Println("    GET    /api/cars/search")
		log.Println("    GET    /api/cars/available")
		log.Println("    GET    /api/cars/:id")
		log.Println("\n  USER (Requires Authentication):")
		log.Println("    GET    /api/me")
		log.Println("\n  ADMIN (Requires Admin Role):")
		log.Println("    POST   /api/admin/cars")
		log.Println("    PUT    /api/admin/cars/:id")
		log.Println("    DELETE /api/admin/cars/:id")
		log.Println()

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