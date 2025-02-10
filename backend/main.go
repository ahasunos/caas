package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Database connection details
var db *sql.DB

func init() {
	var err error
	// db, err = sql.Open("postgres", "host=localhost port=5432 user=postgres password=password123 dbname=inspec sslmode=disable")
	db, err = sql.Open("postgres", "host=inspec-postgres port=5432 user=postgres password=password123 dbname=inspec sslmode=disable")

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
}

// Profile struct to hold profile data
type Profile struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Stars       int       `json:"stars"`
	LastUpdated time.Time `json:"last_updated"`
}

// GitHubSearchResult struct to parse GitHub API search response
type GitHubSearchResult struct {
	Items []GitHubRepo `json:"items"`
}

// GitHubRepo struct represents a GitHub repository
type GitHubRepo struct {
	Name        string `json:"name"`
	HTMLURL     string `json:"html_url"`
	Description string `json:"description"`
	Stars       int    `json:"stargazers_count"`
}

// Endpoint: /
func welcomeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to InSpec as a Service!",
	})
}

// Endpoint: /fetch-profiles
func fetchProfilesHandler(c *gin.Context) {
	profiles, err := getProfilesFromDatabase()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not fetch profiles from database.",
		})
		return
	}

	if len(profiles) == 0 {
		// No profiles found, fetch from GitHub
		if err := updateProfilesFromGitHub(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update profiles from GitHub.",
			})
			return
		}

		profiles, _ = getProfilesFromDatabase() // Retry after updating
	}

	c.JSON(http.StatusOK, profiles)
}

// Endpoint: /update-profiles
func updateProfilesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile update in progress, please check back later.",
	})

	// Fetch and update profiles from GitHub
	if err := updateProfilesFromGitHub(); err != nil {
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
	if hasInSpecYML(request.URL) {
		// Fetch details of the repository
		profile, err := fetchProfileDetailsFromGitHub(request.URL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch profile details from GitHub.",
			})
			return
		}

		// Insert the profile into the database
		if err := insertProfileIntoDatabase(profile); err != nil {
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

// Function to fetch profile details from GitHub
func fetchProfileDetailsFromGitHub(repoURL string) (Profile, error) {
	// Extract repository owner and name from URL
	repoParts := strings.Split(repoURL, "/")
	owner := repoParts[len(repoParts)-2]
	repo := repoParts[len(repoParts)-1]
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)

	// Fetch repository details from GitHub
	resp, err := http.Get(apiURL)
	if err != nil {
		return Profile{}, fmt.Errorf("failed to fetch repository details: %v", err)
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		return Profile{}, fmt.Errorf("GitHub repository not found or not accessible")
	}

	var repoDetails struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		HTMLURL     string `json:"html_url"`
		Stargazers  int    `json:"stargazers_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repoDetails); err != nil {
		return Profile{}, fmt.Errorf("failed to decode repository details: %v", err)
	}

	// Construct Profile object
	return Profile{
		Name:        repoDetails.Name,
		URL:         repoDetails.HTMLURL,
		Description: repoDetails.Description,
		Stars:       repoDetails.Stargazers,
		LastUpdated: time.Now(),
	}, nil
}

// Function to insert profile into the database
func insertProfileIntoDatabase(profile Profile) error {
	// Check if the profile already exists in the database
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM inspec_profiles WHERE url = $1", profile.URL).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check if profile exists: %v", err)
	}

	if count > 0 {
		// Profile already exists, no need to insert
		return nil
	}

	// Insert new profile into the database
	_, err = db.Exec("INSERT INTO inspec_profiles (name, url, description, stars, last_updated) VALUES ($1, $2, $3, $4, $5)",
		profile.Name, profile.URL, profile.Description, profile.Stars, profile.LastUpdated)
	if err != nil {
		return fmt.Errorf("failed to insert profile into the database: %v", err)
	}

	return nil
}

// Function to get profiles from database
func getProfilesFromDatabase() ([]Profile, error) {
	rows, err := db.Query("SELECT id, name, url, description, stars, last_updated FROM inspec_profiles ORDER BY stars DESC")
	if err != nil {
		log.Printf("Error querying database: %v", err)
		return nil, err
	}
	defer rows.Close()

	var profiles []Profile
	for rows.Next() {
		var profile Profile
		if err := rows.Scan(&profile.ID, &profile.Name, &profile.URL, &profile.Description, &profile.Stars, &profile.LastUpdated); err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// Function to update profiles from GitHub
func updateProfilesFromGitHub() error {
	// Fetch profiles from GitHub API
	profiles, err := fetchProfilesFromGitHub()
	if err != nil {
		return err
	}

	// Update or insert profiles in database
	for _, profile := range profiles {
		// Check if profile already exists
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM inspec_profiles WHERE url = $1", profile.URL).Scan(&count)
		if err != nil {
			return err
		}

		if count > 0 {
			// Update existing profile
			_, err = db.Exec("UPDATE inspec_profiles SET stars = $1, description = $2, last_updated = $3 WHERE url = $4",
				profile.Stars, profile.Description, time.Now(), profile.URL)
			if err != nil {
				return err
			}
		} else {
			// Insert new profile
			_, err = db.Exec("INSERT INTO inspec_profiles (name, url, description, stars, last_updated) VALUES ($1, $2, $3, $4, $5)",
				profile.Name, profile.URL, profile.Description, profile.Stars, time.Now())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Function to fetch profiles from GitHub API
func fetchProfilesFromGitHub() ([]Profile, error) {
	var allProfiles []Profile
	page := 1
	perPage := 10

	for {
		url := fmt.Sprintf("https://api.github.com/search/repositories?q=inspec+profile&sort=stars&per_page=%d&page=%d", perPage, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}
		req.Header.Set("User-Agent", "InSpecService")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch profiles from GitHub: %v", err)
		}
		defer resp.Body.Close()

		var result GitHubSearchResult
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return nil, fmt.Errorf("failed to decode response from GitHub: %v", err)
		}

		// Exit if no more profiles are found
		if len(result.Items) == 0 {
			break
		}

		// Loop through the items and add valid profiles that have inspec.yml
		for _, repo := range result.Items {
			// Check if inspec.yml exists in the repository's root
			if repo.Description != "" && hasInSpecYML(repo.HTMLURL) {
				allProfiles = append(allProfiles, Profile{
					Name:        repo.Name,
					URL:         repo.HTMLURL,
					Description: repo.Description,
					Stars:       repo.Stars,
				})
			}
		}

		// Move to the next page
		page++
	}

	return allProfiles, nil
}

// Function to check if the inspec.yml file exists in the repository's root
// Function to check if the inspec.yml file exists in the repository's root
func hasInSpecYML(repoURL string) bool {
	// Construct the API URL to get the contents of the repo

	repoParts := strings.Split(repoURL, "/")
	owner := repoParts[len(repoParts)-2]
	repo := repoParts[len(repoParts)-1]
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/inspec.yml", owner, repo)

	// Send GET request to check for inspec.yml file
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Printf("Error fetching %s: %v", apiURL, err)
		return false
	}
	defer resp.Body.Close()

	// Debugging log to check the status code and response
	log.Printf("GitHub API response for %s: %v", apiURL, resp.StatusCode)

	// Check if the status code is 200 (OK) and the content is the expected 'inspec.yml' file
	if resp.StatusCode == http.StatusOK {
		log.Printf("inspec.yml found in repository %s", repoURL)
		return true
	}

	// Handle other status codes (e.g., 404 if the file is not found)
	if resp.StatusCode == http.StatusNotFound {
		log.Printf("inspec.yml not found in repository %s", repoURL)
	} else {
		log.Printf("Unexpected status code %d from GitHub API for %s", resp.StatusCode, apiURL)
	}
	return false
}

func main() {
	r := gin.Default()

	// Define routes
	r.GET("/", welcomeHandler)
	r.GET("/fetch-profiles", fetchProfilesHandler)
	r.GET("/update-profiles", updateProfilesHandler)
	r.POST("/add-profile", addProfileHandler) // New endpoint to add profile
	r.POST("/execute-profile", executeProfileHandler)

	// Run the server
	r.Run(":8080")
}
