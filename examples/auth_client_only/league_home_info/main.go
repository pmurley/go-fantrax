package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/pmurley/go-fantrax/auth_client"
)

func main() {
	// Get league ID from environment variable or use default
	leagueID := os.Getenv("FANTRAX_LEAGUE_ID")
	if leagueID == "" {
		leagueID = "q8lydqf5m4u30rca" // Default from the example
	}

	// Create client with caching enabled
	client, err := auth_client.NewClient(leagueID, true)
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}

	fmt.Println("=== Fetching League Home Info ===")

	// Fetch league home info
	homeInfo, err := client.GetLeagueHomeInfo()
	if err != nil {
		log.Fatalf("Failed to get league home info: %v", err)
	}

	// Display league settings
	fmt.Printf("\n--- League Settings ---\n")
	fmt.Printf("League Name: %s\n", homeInfo.Settings.LeagueName)
	fmt.Printf("Year: %s\n", homeInfo.Settings.Year)
	fmt.Printf("Sport ID: %s\n", homeInfo.Settings.SportID)
	fmt.Printf("Premium Type: %s\n", homeInfo.Settings.PremiumLeagueType)
	if homeInfo.Settings.LogoUploaded {
		fmt.Printf("Logo URL: %s\n", homeInfo.Settings.LogoURL)
	}

	// Display teams
	fmt.Printf("\n--- Fantasy Teams (%d total) ---\n", len(homeInfo.Teams))
	for i, team := range homeInfo.Teams {
		if i >= 5 {
			fmt.Printf("  ... and %d more teams\n", len(homeInfo.Teams)-5)
			break
		}
		commish := ""
		if team.Commissioner {
			commish = " [Commissioner]"
		}
		fmt.Printf("  %s (%s)%s\n", team.Name, team.ShortName, commish)
		if team.LogoURL256 != "" {
			fmt.Printf("    Logo: %s\n", team.LogoURL256)
		}
	}

	// Display standings by division
	fmt.Printf("\n--- Standings by Division ---\n")
	for _, division := range homeInfo.Standings {
		fmt.Printf("\n%s:\n", division.DivisionName)
		for _, team := range division.Teams {
			commish := ""
			if team.Commissioner {
				commish = "*"
			}
			fmt.Printf("  %d. %-30s %s  Win%%: %s  GB: %s  Pts: %s%s\n",
				team.Rank,
				team.TeamName,
				team.Record,
				team.WinPercentage,
				team.GamesBack,
				team.Points,
				commish)
		}
	}

	// Display matchups
	fmt.Printf("\n--- Current Matchups ---\n")
	fmt.Printf("Period: %s\n", homeInfo.Matchups.PeriodInfo)
	fmt.Printf("Live: %v\n", homeInfo.Matchups.Live)
	if len(homeInfo.Matchups.Games) == 0 {
		fmt.Printf("Message: %s\n", homeInfo.Matchups.NoMatchupsMsg)
	} else {
		for _, game := range homeInfo.Matchups.Games {
			fmt.Printf("  %s (%s) vs %s (%s)\n",
				game.AwayTeamName, game.AwayTeamScore,
				game.HomeTeamName, game.HomeTeamScore)
		}
	}

	// Save to JSON file
	outputFile := "league_home_info.json"
	jsonData, err := json.MarshalIndent(homeInfo, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal to JSON: %v", err)
	} else {
		err = os.WriteFile(outputFile, jsonData, 0644)
		if err != nil {
			log.Printf("Failed to write output file: %v", err)
		} else {
			fmt.Printf("\nProcessed data saved to: %s\n", outputFile)
		}
	}
}