package models

import "time"

// Profile represents an InSpec profile.
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
