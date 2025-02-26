package github

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

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
