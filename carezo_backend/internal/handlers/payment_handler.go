package handlers

import (
	"net/http"

	"github.com/delaquash/carezo/configs"
	"github.com/delaquash/carezo/internal/services"
	response "github.com/delaquash/carezo/pkg"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *services.PaymentService
	cfg			*configs.Config
}


func NewPaymentHandler(cfg *configs.Config) *PaymentHandler {
	return &PaymentHandler{
		paymentService: services.NewPaymentService(cfg.PaystackSecretKey),
		cfg:			cfg,
	}
}

// POST /api/payments/initialize
func (h *PaymentHandler) InitializePayment(c *gin.Context) {
	_, exist := c.Get("user_id")
	if !exist {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	userEmail, _ := c.Get("user_email")

	var req struct {
		BookingID string `json:"booking_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	result, err := h.paymentService.InitializePayment(req.BookingID, userEmail.(string))

	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "payment initialize", result)
}

func (h *PaymentHandler) VerifyPayment(c *gin.Context) {
	_, exists := c.Get("user_id")

	if !exists {
		response.Error(c, http.StatusBadRequest, "unauthorized")
		return
	}

	reference := c.Param("reference")

	if reference == "" {
		response.Error(c, http.StatusBadRequest, "payment reference is required")
		return
	}

	if err := h.paymentService.VerifyPayment(reference); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "payment verifies successfully", nil)
}
