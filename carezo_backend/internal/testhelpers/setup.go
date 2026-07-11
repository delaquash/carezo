package testhelpers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/delaquash/carezo/configs"
	"github.com/delaquash/carezo/internal/database"
	"github.com/delaquash/carezo/internal/handlers"
	"github.com/delaquash/carezo/internal/middleware"
	"github.com/delaquash/carezo/internal/services"
	"github.com/delaquash/carezo/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// testApps holds everythibg a test needs to make http requests

type TestApp struct {
	Router *gin.Engine
	DB     *sqlx.DB
	Config *configs.Config
}

// setuptestapp creates a complete Gin router wired to a TEST DB
// called at the top of every test file

func SetUpTestApp(t *testing.T) *TestApp {
	t.Helper() // marks this as a helper — errors point to the caller, not here

	//  separate DB to test in env
	testDBURL := os.Getenv("TEST_DATABASE_URL")

	// fallback if testDB fails
	if testDBURL == "" {
		testDBURL = "postgres://carezo_user:Equarshie85@@localhost:5432/carezo_test_db?sslmode=disable"
	}

	cfg := &configs.Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "carezo_user",
		DBPassword: "password",
		DBName:     "carezo_test_db", // ← separate test DB, NOT carezo_db
		JWTSecret:  "test-secret-key",
		AppEnv:     "test",
	}

	db, err := database.ConnectPostgres(cfg)
	require.NoError(t, err, "failed to connect to test database")

	// Run migrations on the test DB so it has the right schema
	// (or just use the same schema as your main DB)

	gin.SetMode(gin.TestMode) // quieter output during tests

	// Build the same router as main.go but pointing to test DB
	router := gin.New()

	// inject the shared cloudinary service (use a mock in tests)
	cloudinarySvc := &MockCloudinaryService{}

	bookingHandler := handlers.NewBookingHandler()
	carHandler := handlers.NewCarHandler(cloudinarySvc)
	driverHandler := handlers.NewDriverHandler(cloudinarySvc)
	userHandler := handlers.NewUserHandler(cloudinarySvc)
	authHandler := handlers.NewAuthHandler(cfg)
	paymentHandler := handlers.NewPaymentHandler(cfg)

	api := router.Group("/api")

	// public routes
	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/verify-otp", authHandler.VerifyOTP)

	// public car/driver routes
	api.GET("/cars", carHandler.ListAllCars)
	api.GET("/cars/:id", carHandler.GetCar)
	api.GET("/drivers", driverHandler.ListAllDrivers)

	// webhook — must be public, no auth
	api.POST("/payments/webhook", paymentHandler.HandleWebhook)

	// protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))

	bookings := protected.Group("/bookings")
	bookings.POST("", bookingHandler.CreateBooking)
	bookings.GET("", bookingHandler.ListUserBooking)
	bookings.GET("/:id", bookingHandler.GetBooking)
	bookings.POST("/:id/cancel", bookingHandler.CancelBooking)
	bookings.PUT("/:id", bookingHandler.UpdateBooking)

	protected.GET("/user/get-profile", userHandler.GetMe)

	// admin routes
	admin := protected.Group("/admin")
	admin.Use(middleware.RequireRole("admin"))
	admin.POST("/cars", carHandler.CreateCar)
	admin.DELETE("/cars/:id", carHandler.DeleteCar)

	return &TestApp{Router: router, DB: db, Config: cfg}
}

// CleanupDB removes test data between tests so they don't interfere
func (app *TestApp) CleanUpDB(t *testing.T) {
	t.Helper()

	// order matters — foreign keys require deleting child tables first
	_, err := app.DB.Exec(`
        DELETE FROM bookings;
        DELETE FROM cars;
        DELETE FROM drivers;
        DELETE FROM users WHERE email NOT LIKE '%admin%';
    `)
	require.NoError(t, err)
}

// MakeRequest is a helper that fires an HTTP request at the test router
// and returns the response recorder so you can inspect status + body

func (app *TestApp) MakeRequest(
	method string,
	url string,
	body interface{}, // pass nill for GET requests
	token string, //pass "" for unauthenticated requests
) *httptest.ResponseRecorder {
	// convet body struct to JSON bytes
	var reqBody *bytes.Buffer

	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")

	// add auth token if provided
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// let response recorder capture the responsw
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)
	return w
}

// parseresponse to decode the JSON response body into a map
func ParseResponse(w *httptest.ResponseRecorder) map[string]interface{} {
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	return result
}

// GenerateTestToken creates a JWT for testing without going through login
func GenerateTestToken(userID string, role string, secret string) string {
	// use your existing JWT generation function
	// this avoids needing to call /api/auth/login in every test
	token, _ := utils.GenerateAccessToken(userID, role, secret)
	return token
}

// MockCloudinaryService — does nothing in tests (no real Cloudinary calls)
type MockCloudinaryService struct{}

func (m *MockCloudinaryService) UploadImage(file multipart.File, folder string) (*services.UploadResult, error) {
	return &services.UploadResult{URL: "https://mock.cloudinary.com/test.jpg", PublicID: "test/123"}, nil
}
func (m *MockCloudinaryService) DeleteImage(publicID string) error { return nil }
