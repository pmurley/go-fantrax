package main

import (
	"fmt"
	"os"

	"github.com/pmurley/go-fantrax"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Get league ID from environment variable or use default
	leagueID := os.Getenv("FANTRAX_LEAGUE_ID")
	if leagueID == "" {
		leagueID = "q8lydqf5m4u30rca" // Default from the example
	}

	// Create client with caching enabled
	client, err := fantrax.NewClient(leagueID, true)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Fetch league info
	//fmt.Println("Fetching league info...")
	//leagueInfo, err := client.GetLeagueInfo(leagueID)
	//if err != nil {
	//	log.Fatalf("Failed to get league info: %v", err)
	//}

	rosterInfo, err := client.GetTeamRosters()
	if err != nil {
		log.Fatalf("Failed to get team rosters: %v", err)
	}

	for _, roster := range rosterInfo.Rosters {
		for _, player := range roster.RosterItems {
			fmt.Printf("Team: %s, PlayerId: %s, Status: %s\n", roster.TeamName, player.ID, player.Status)
		}
	}

}
