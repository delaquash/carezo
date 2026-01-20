package response

import "github.com/gin-gonic/gin"

// this is for standardizing API responses

type Response struct {
	Success  bool    		`json:"success"`
	Message  string  		`json:"message,omitempty"`
	Data     interface{}	`json:"data, omitempty"`
	Error    string  		`json:"error,omitempty"`
}

// Sending a successful response
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response {
		Success: true,
		Message: message,
		Data: data,
	})
}


// Error message whn it fails
func Error (c *gin.Context, statusCode int, errMessage string) {
	c.JSON (statusCode, Response {
		Success: false,
		Error: errMessage,
	})
}