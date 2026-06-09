package handlers

import (
	"net/http"

	"github.com/delaquash/carezo/internal/services"
	response "github.com/delaquash/carezo/pkg"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{
		notificationService: services.NewNotification(),
	}
}


// GET  /api/notifications
func(h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")

	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
	}

	notification, err := h.notificationService.GetUserNotification(userID.(string))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
	}

	// return unread count so FE can update badge
	unreadCount, _ := h.notificationService.GetUnreadCount(userID.(string))

	response.Success(c, http.StatusOK, "notifications retrieved", gin. H {
		"notifications": notification,
		"unread_count" : unreadCount,
	})
}


// PUT /api/notification/read-all
func (h *NotificationHandler) MarkAllRead(c *gin.Context){
	userID, exists := c.Get("user_id")

	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.notificationService.MarkAllRead(userID.(string)); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "all notification marked as read", nil)
}

// PUT /api/notifications/:id/read
func (h *NotificationHandler) MarkOneRead(c *gin.Context) {
	userID, exists := c.Get("user_id")

	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	notificationID := c.Param("id")

	if err := h.notificationService.MarkOneread(notificationID, userID.(string)); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "notification marked as read", nil)
}

// GET /api/notification/unread-count
func (h *NotificationHandler) GetUnreadCount(c *gin.Context){
	userID, exists := c.Get("user_id")

	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	count, err := h.notificationService.GetUnreadCount(userID.(string))

	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "unread count retrieved", gin.H {
		"unread_count":count,
	})
}


func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")

	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	notificationID := c.Param("id")

	if err := h.notificationService.DeleteNotification(notificationID, userID.(string)); err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "notification deleted", nil)

} 

func (s *NotificationHandler) DeleteAllNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")

	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := s.notificationService.DeleteAllNotification(userID.(string)); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "all notification deleted", nil)
}