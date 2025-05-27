package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Mock Content Service - manages onboarding screen configurations
func main() {
	router := gin.Default()

	// Screen configuration endpoint
	router.GET("/api/screens/:screenId", func(c *gin.Context) {
		screenID := c.Param("screenId")
		version := c.Query("version")
		if version == "" {
			version = "v1"
		}

		// Simulate content retrieval delay
		time.Sleep(80 * time.Millisecond)

		var screen map[string]interface{}

		switch screenID {
		case "welcome":
			screen = map[string]interface{}{
				"id":       "welcome",
				"version":  version,
				"title":    "Welcome to Our App!",
				"subtitle": "Let's get you started with a quick setup",
				"type":     "welcome",
				"fields":   []interface{}{},
				"actions": []map[string]interface{}{
					{
						"id":    "continue",
						"type":  "next",
						"label": "Get Started",
						"style": "primary",
					},
				},
				"validation": map[string]interface{}{
					"required_fields": []string{},
				},
				"next_screen": "personal_info",
			}

		case "personal_info":
			screen = map[string]interface{}{
				"id":       "personal_info",
				"version":  version,
				"title":    "Tell us about yourself",
				"subtitle": "We need some basic information to personalize your experience",
				"type":     "personal_info",
				"fields": []map[string]interface{}{
					{
						"id":          "first_name",
						"type":        "text",
						"label":       "First Name",
						"placeholder": "Enter your first name",
						"required":    true,
						"validation": map[string]interface{}{
							"min_length": 2,
							"max_length": 50,
						},
					},
					{
						"id":          "last_name",
						"type":        "text",
						"label":       "Last Name",
						"placeholder": "Enter your last name",
						"required":    true,
						"validation": map[string]interface{}{
							"min_length": 2,
							"max_length": 50,
						},
					},
					{
						"id":          "email",
						"type":        "email",
						"label":       "Email Address",
						"placeholder": "Enter your email",
						"required":    true,
						"validation": map[string]interface{}{
							"pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
						},
					},
					{
						"id":          "phone",
						"type":        "phone",
						"label":       "Phone Number",
						"placeholder": "Enter your phone number",
						"required":    false,
					},
				},
				"actions": []map[string]interface{}{
					{
						"id":    "back",
						"type":  "back",
						"label": "Back",
						"style": "secondary",
					},
					{
						"id":    "continue",
						"type":  "submit",
						"label": "Continue",
						"style": "primary",
					},
				},
				"validation": map[string]interface{}{
					"required_fields": []string{"first_name", "last_name", "email"},
				},
				"next_screen": "preferences",
			}

		case "preferences":
			screen = map[string]interface{}{
				"id":       "preferences",
				"version":  version,
				"title":    "Customize your experience",
				"subtitle": "Choose your preferences to get the most out of our app",
				"type":     "preferences",
				"fields": []map[string]interface{}{
					{
						"id":       "theme",
						"type":     "select",
						"label":    "Theme Preference",
						"required": true,
						"options": []map[string]interface{}{
							{"value": "light", "label": "Light"},
							{"value": "dark", "label": "Dark"},
							{"value": "auto", "label": "Auto"},
						},
						"default_value": "auto",
					},
					{
						"id":            "notifications",
						"type":          "checkbox",
						"label":         "Enable Notifications",
						"required":      false,
						"default_value": true,
					},
					{
						"id":       "language",
						"type":     "select",
						"label":    "Language",
						"required": true,
						"options": []map[string]interface{}{
							{"value": "en", "label": "English"},
							{"value": "es", "label": "Spanish"},
							{"value": "fr", "label": "French"},
						},
						"default_value": "en",
					},
				},
				"actions": []map[string]interface{}{
					{
						"id":    "back",
						"type":  "back",
						"label": "Back",
						"style": "secondary",
					},
					{
						"id":    "continue",
						"type":  "submit",
						"label": "Continue",
						"style": "primary",
					},
				},
				"validation": map[string]interface{}{
					"required_fields": []string{"theme", "language"},
				},
				"next_screen": "completion",
			}

		case "completion":
			screen = map[string]interface{}{
				"id":       "completion",
				"version":  version,
				"title":    "You're all set!",
				"subtitle": "Welcome to our community. Let's start exploring!",
				"type":     "completion",
				"fields":   []interface{}{},
				"actions": []map[string]interface{}{
					{
						"id":    "finish",
						"type":  "submit",
						"label": "Start Using App",
						"style": "primary",
					},
				},
				"validation": map[string]interface{}{
					"required_fields": []string{},
				},
				"next_screen": "",
			}

		default:
			c.JSON(http.StatusNotFound, map[string]interface{}{
				"success": false,
				"error":   "Screen not found",
			})
			return
		}

		// Add version-specific enhancements for v2
		if version == "v2" {
			screen["metadata"] = map[string]interface{}{
				"analytics_enabled": true,
				"ab_test_variant":   "default",
				"personalization":   true,
			}
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"data":    screen,
		})
	})

	// Flow configuration endpoint
	router.GET("/api/flows/:flowId", func(c *gin.Context) {
		flowID := c.Param("flowId")

		// Simulate configuration retrieval
		time.Sleep(60 * time.Millisecond)

		flow := map[string]interface{}{
			"id":      flowID,
			"name":    "Mobile Onboarding",
			"version": "1.0",
			"screens": []string{"welcome", "personal_info", "preferences", "completion"},
			"rules": map[string]interface{}{
				"allow_skip":       false,
				"required_screens": []string{"personal_info"},
				"conditional_flow": true,
				"timeout_minutes":  30,
			},
			"analytics": map[string]interface{}{
				"enabled":           true,
				"track_views":       true,
				"track_submissions": true,
				"track_timing":      true,
			},
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"data":    flow,
		})
	})

	// Analytics endpoint
	router.POST("/api/analytics/events", func(c *gin.Context) {
		var event map[string]interface{}
		if err := c.ShouldBindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		// Simulate event processing
		time.Sleep(30 * time.Millisecond)

		c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"event_id": "evt_" + fmt.Sprintf("%d", time.Now().Unix()),
				"recorded": true,
			},
		})
	})

	fmt.Println("Mock Content Service starting on :8082")
	log.Fatal(http.ListenAndServe(":8082", router))
}
