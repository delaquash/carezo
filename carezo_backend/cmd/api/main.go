package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	if err := r.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}

	log.Println("Server running on http://localhost:8081")
	r.Run("8081")	
}