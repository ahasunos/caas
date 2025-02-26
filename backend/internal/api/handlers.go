package api

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/ahasunos/inspec-cloud/backend/internal/db"
	"github.com/ahasunos/inspec-cloud/backend/internal/github"
	"github.com/gin-gonic/gin"
)

// Endpoint: /
func welcomeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to InSpec as a Service!",
	})
}

// Endpoint: /fetch-profiles
func fetchProfilesHandler(c *gin.Context) {
	profiles, err := db.GetProfilesFromDatabase()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not fetch profiles from database.",
		})
		return
	}

	if len(profiles) == 0 {
		// No profiles found, fetch from GitHub
		if err := db.UpdateProfilesFromGitHub(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update profiles from GitHub.",
			})
			return
		}

		profiles, _ = db.GetProfilesFromDatabase() // Retry after updating
	}

	c.JSON(http.StatusOK, profiles)
}

// Endpoint: /update-profiles
func updateProfilesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile update in progress, please check back later.",
	})

	// Fetch and update profiles from GitHub
	if err := db.UpdateProfilesFromGitHub(); err != nil {
		log.Println("Error updating profiles:", err)
	}
}

// Endpoint: /add-profile
func addProfileHandler(c *gin.Context) {
	var request struct {
		URL string `json:"url"`
	}

	// Bind JSON request body to the URL field
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request payload.",
		})
		return
	}

	// Check if the repository at the URL contains an inspec.yml file
	if github.HasInSpecYML(request.URL) {
		// Fetch details of the repository
		profile, err := github.FetchProfileDetailsFromGitHub(request.URL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch profile details from GitHub.",
			})
			return
		}

		// Insert the profile into the database
		if err := db.InsertProfileIntoDatabase(profile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to insert profile into the database.",
			})
			return
		}

		// Return success message
		c.JSON(http.StatusOK, gin.H{
			"message": "Profile added successfully.",
			"profile": profile,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The provided repository is not a valid InSpec profile (missing inspec.yml).",
		})
	}
}

func executeProfileHandler(c *gin.Context) {
	// Parse request body
	var req struct {
		Hostname   string `json:"hostname"`
		Username   string `json:"username"`
		Profile    string `json:"profile"`
		PrivateKey string `json:"private_key"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Decode the Base64 encoded private key
	decodedKey, err := base64.StdEncoding.DecodeString(req.PrivateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode private key"})
		return
	}

	// Save the decoded private key to a temporary file
	privateKeyPath := "/tmp/temp-key.pem"
	if err := os.WriteFile(privateKeyPath, decodedKey, 0600); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save private key"})
		return
	}
	defer os.Remove(privateKeyPath) // Cleanup after execution

	log.Printf("Executing InSpec profile %s on %s as %s", req.Profile, req.Hostname, req.Username)
	log.Printf("Private key saved to %s", privateKeyPath)

	start := time.Now()
	// Construct InSpec command
	cmd := exec.Command("inspec", "exec", req.Profile, "-t", fmt.Sprintf("ssh://%s@%s", req.Username, req.Hostname), "-i", privateKeyPath, "--chef-license", "accept", "--chef-license-key", "free-833b40cf-336a-42ee-b71d-f14a078107b9-5090")

	// Execute command and capture output
	output, err := cmd.CombinedOutput()
	log.Printf("InSpec command executed in %s", time.Since(start))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Execution failed", "details": string(output)})
		return
	}

	// Return execution results
	c.JSON(http.StatusOK, gin.H{"output": string(output)})
}
