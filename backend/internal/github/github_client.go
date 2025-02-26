package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ahasunos/inspec-cloud/backend/internal/models"
)

// Function to fetch profile details from GitHub
func FetchProfileDetailsFromGitHub(repoURL string) (models.Profile, error) {
	// Extract repository owner and name from URL
	repoParts := strings.Split(repoURL, "/")
	owner := repoParts[len(repoParts)-2]
	repo := repoParts[len(repoParts)-1]
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)

	// Fetch repository details from GitHub
	resp, err := http.Get(apiURL)
	if err != nil {
		return models.Profile{}, fmt.Errorf("failed to fetch repository details: %v", err)
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		return models.Profile{}, fmt.Errorf("GitHub repository not found or not accessible")
	}

	var repoDetails struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		HTMLURL     string `json:"html_url"`
		Stargazers  int    `json:"stargazers_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repoDetails); err != nil {
		return models.Profile{}, fmt.Errorf("failed to decode repository details: %v", err)
	}

	// Construct Profile object
	return models.Profile{
		Name:        repoDetails.Name,
		URL:         repoDetails.HTMLURL,
		Description: repoDetails.Description,
		Stars:       repoDetails.Stargazers,
		LastUpdated: time.Now(),
	}, nil
}

// Function to fetch profiles from GitHub API
func FetchProfilesFromGitHub() ([]models.Profile, error) {
	var allProfiles []models.Profile
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

		var result models.GitHubSearchResult
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
			if repo.Description != "" && HasInSpecYML(repo.HTMLURL) {
				allProfiles = append(allProfiles, models.Profile{
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
