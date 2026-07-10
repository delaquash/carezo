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
	"github.com/delaquash/carezo/internal/services"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("MAIN FUNCTION STARTED")
	// load configs
	cfg := configs.LoadConfig()
	log.Printf("Starting %s in %s mode", cfg.AppName, cfg.AppEnv)

	// connecting to postgresql
	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer database.ClosePostgres()

	// Run seeder after DB is ready for admin user

	log.Println(" ABOUT TO RUN SEEDER")
	database.SeedAdminUser(db, cfg)
	log.Println(" SEEDER FINISHED")

	// connect to redis
	redis, err := database.ConnectRedis(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer database.CloseRedis()

	// cloudinary
	cloudinaryService, err := services.NewCloudinaryService(cfg)
	if err != nil {
		log.Fatalf("Failed to initiative Cloudinary: %v", err)
	}

	log.Println("Cloudinary initialised successfully")

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
	carHandler := handlers.NewCarHandler(cloudinaryService)
	driverHandler := handlers.NewDriverHandler(cloudinaryService)
	userHandler := handlers.NewUserHandler(cloudinaryService)
	bookingHandler := handlers.NewBookingHandler()
	paymentHandler := handlers.NewPaymentHandler(cfg)
	notificationHandler := handlers.NewNotificationHandler()
	reviewHandler := handlers.NewReviewHandler(cloudinaryService)
	uploaderHandler := handlers.NewUploadHandler(cloudinaryService)

	// routes
	api := router.Group("/api")
	{
		//  AUTH
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/verify-otp", authHandler.VerifyOTP)
			auth.POST("/resend-otp", authHandler.ResendOTP)
			auth.POST("/login", authHandler.Login)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		cars := api.Group("/cars")
		{
			cars.GET("", carHandler.ListAllCars)
			cars.GET("/search", carHandler.SearchCars)
			cars.GET("/available", carHandler.GetAvailableCars)
			cars.GET("/:id", carHandler.GetCar)
			cars.GET("/nearby", carHandler.GetNearByCars)
			cars.GET("/popular", carHandler.GetPopularCars)
		}

		drivers := api.Group("/drivers")
		{
			drivers.GET("", driverHandler.ListAllDrivers)
			drivers.GET("/search", driverHandler.SearchDrivers)
			drivers.GET("/:id", driverHandler.GetDriver)
			drivers.GET("/:id/reviews", driverHandler.GetDriverReviews)
		}

		payment := api.Group("/payments")
		{
			payment.POST("/webhook", paymentHandler.HandleWebhook)
			payment.POST("/initialize", middleware.AuthMiddleware(cfg), paymentHandler.InitializePayment)
		}

		// UPload
		upload := api.Group("/upload")
		upload.Use(middleware.AuthMiddleware(cfg))
		{
			upload.POST("", uploaderHandler.UploadSingleImage)
			upload.POST("/multiple", uploaderHandler.UploadMultipleImages)
		}

		//  PROTECTED
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{

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
			notifications := protected.Group("/notifications")
			{
				notifications.GET("", notificationHandler.GetNotifications)
				notifications.GET("/unread-count", notificationHandler.GetUnreadCount)
				notifications.PUT("/read-all", notificationHandler.MarkAllRead)
				notifications.PUT("/:id/read", notificationHandler.MarkOneRead)
				notifications.DELETE("/:id", notificationHandler.DeleteNotification)
				notifications.DELETE("", notificationHandler.DeleteAllNotification)
			}

			bookings := protected.Group("/bookings")
			{
				bookings.POST("", bookingHandler.CreateBooking)
				bookings.GET("", bookingHandler.ListUserBooking)
				bookings.GET("/:id", bookingHandler.GetBooking)
				bookings.POST("/:id/cancel", bookingHandler.CancelBooking)
			}

			// protected.POST("/reviews", driverHandler.CreateReview)
		}

		// Reviews
		reviews := protected.Group("/reviews")
		{
			reviews.POST("", driverHandler.CreateReview)
			reviews.GET("/:id", reviewHandler.GetReviewByID)
			reviews.PUT("/:id/images", reviewHandler.EditReviewImage)
		}

		// Admin routes
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRole("admin"))
		{
			// Car management
			cars := admin.Group("/cars")
			{
				cars.POST("", carHandler.CreateCar)
				cars.PUT("/:id", carHandler.UpdateCar)
				cars.DELETE("/:id", carHandler.DeleteCar)
			}

			// Driver management
			drivers := admin.Group("/drivers")
			{
				drivers.POST("", driverHandler.CreateDriver)
				drivers.PUT("/:id", driverHandler.UpdateDriver)
				drivers.DELETE("/:id", driverHandler.DeleteDriver)
			}

			// User management
			users := admin.Group("/users")
			{
				users.GET("", userHandler.ListAllUsers)
				users.GET("/:id", userHandler.GetUserByID)
				users.PUT("/:id/status", userHandler.UpdateUserStatus)
			}
		}

	}

	//  SERVER
	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server listening on port %s", cfg.AppPort)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// GRACEFUL SHUTDOWN
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
