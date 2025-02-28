package db

import (
	"fmt"
	"log"
	"time"

	"github.com/ahasunos/caas/backend/internal/github"
	"github.com/ahasunos/caas/backend/internal/models"
)

// Function to insert profile into the database
func InsertProfileIntoDatabase(profile models.Profile) error {
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
func GetProfilesFromDatabase() ([]models.Profile, error) {
	rows, err := db.Query("SELECT id, name, url, description, stars, last_updated FROM inspec_profiles ORDER BY stars DESC")
	if err != nil {
		log.Printf("Error querying database: %v", err)
		return nil, err
	}
	defer rows.Close()

	var profiles []models.Profile
	for rows.Next() {
		var profile models.Profile
		if err := rows.Scan(&profile.ID, &profile.Name, &profile.URL, &profile.Description, &profile.Stars, &profile.LastUpdated); err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// Function to update profiles from GitHub
func UpdateProfilesFromGitHub() error {
	// Fetch profiles from GitHub API
	profiles, err := github.FetchProfilesFromGitHub()
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
