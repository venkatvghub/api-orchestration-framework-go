package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Mock User Service - demonstrates backend API that BFF orchestrates
func main() {
	router := gin.Default()

	// User profile endpoint
	router.GET("/api/users/:userId/profile", func(c *gin.Context) {
		userID := c.Param("userId")

		// Simulate processing delay
		time.Sleep(100 * time.Millisecond)

		profile := map[string]interface{}{
			"id":         userID,
			"first_name": "John",
			"last_name":  "Doe",
			"email":      "john.doe@example.com",
			"avatar":     "https://example.com/avatar.jpg",
			"created_at": time.Now().Add(-30 * 24 * time.Hour),
			"preferences": map[string]interface{}{
				"theme":         "dark",
				"notifications": true,
				"language":      "en",
			},
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"data":    profile,
		})
	})

	// User progress endpoint
	router.GET("/api/users/:userId/onboarding/progress", func(c *gin.Context) {
		userID := c.Param("userId")

		progress := map[string]interface{}{
			"user_id":           userID,
			"current_screen":    "personal_info",
			"completed_screens": []string{"welcome"},
			"progress":          0.25,
			"status":            "in_progress",
			"started_at":        time.Now().Add(-1 * time.Hour),
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"data":    progress,
		})
	})

	// Update user progress
	router.POST("/api/users/:userId/onboarding/progress", func(c *gin.Context) {
		userID := c.Param("userId")

		var request map[string]interface{}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		// Simulate saving progress
		time.Sleep(50 * time.Millisecond)

		c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"user_id": userID,
				"updated": true,
			},
		})
	})

	// Save user data
	router.POST("/api/users/:userId/data", func(c *gin.Context) {
		userID := c.Param("userId")

		var userData map[string]interface{}
		if err := c.ShouldBindJSON(&userData); err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		// Simulate data validation and saving
		time.Sleep(150 * time.Millisecond)

		c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"user_id": userID,
				"saved":   true,
				"fields":  len(userData),
			},
		})
	})

	fmt.Println("Mock User Service starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", router))
}
