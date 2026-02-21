package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pmurley/go-fantrax/auth_client"
)

func main() {
	leagueID := os.Getenv("FANTRAX_LEAGUE_ID")
	if leagueID == "" {
		log.Fatal("Please set FANTRAX_LEAGUE_ID environment variable")
	}

	// Create authenticated client
	client, err := auth_client.NewClient(leagueID, false)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	fmt.Printf("Logged in as: %s\n\n", client.UserInfo.Username)

	// Fetch and parse the league setup page
	fmt.Println("Fetching league setup matchups...")
	setup, err := client.GetLeagueSetupMatchups()
	if err != nil {
		log.Fatalf("Failed to get league setup matchups: %v", err)
	}

	// Print team summary
	fmt.Printf("\n=== Teams (%d) ===\n", len(setup.Teams))
	for _, team := range setup.Teams {
		fmt.Printf("  %-4s  %-40s  %s  (%d owners)\n", team.ShortName, team.Name, team.TeamID, len(team.Owners))
	}

	// Print division summary
	fmt.Printf("\n=== Divisions (%d) ===\n", len(setup.Divisions))
	for _, div := range setup.Divisions {
		fmt.Printf("  %s (%d teams)\n", div.Name, len(div.TeamIDs))
		for _, teamID := range div.TeamIDs {
			team := auth_client.GetTeamByID(setup, teamID)
			if team != nil {
				fmt.Printf("    - %s (%s)\n", team.Name, team.ShortName)
			} else {
				fmt.Printf("    - %s\n", teamID)
			}
		}
	}

	// Print matchup summary
	periods := auth_client.GetSortedPeriods(setup)
	fmt.Printf("\n=== Matchups (%d periods) ===\n", len(periods))
	if len(periods) > 0 {
		fmt.Printf("  Periods: %d to %d\n", periods[0], periods[len(periods)-1])
	}

	// Print period 1 matchups in detail
	period1 := auth_client.GetMatchupsByPeriod(setup, 1)
	if period1 != nil {
		fmt.Printf("\n=== Period 1 Matchups (%d pairs) ===\n", len(period1))
		for _, pair := range period1 {
			away := auth_client.GetTeamByID(setup, pair.AwayTeamID)
			awayName := pair.AwayTeamID
			if away != nil {
				awayName = away.Name
			}

			if pair.HomeTeamID == "-1" {
				fmt.Printf("  %-40s  BYE\n", awayName)
			} else {
				home := auth_client.GetTeamByID(setup, pair.HomeTeamID)
				homeName := pair.HomeTeamID
				if home != nil {
					homeName = home.Name
				}
				fmt.Printf("  %-40s  vs  %s\n", awayName, homeName)
			}
		}
	}

	// Print some form config details
	fmt.Printf("\n=== Form Config ===\n")
	fmt.Printf("  Hidden fields: %d\n", len(setup.FormConfig.HiddenFields))
	fmt.Printf("  Select fields: %d\n", len(setup.FormConfig.SelectFields))
	fmt.Printf("  Checkbox fields: %d\n", len(setup.FormConfig.CheckboxFields))
	fmt.Printf("  Team names: %d\n", len(setup.FormConfig.TeamNames))
	fmt.Printf("  Team short names: %d\n", len(setup.FormConfig.TeamShortNames))
	fmt.Printf("  Owner email fields: %d\n", len(setup.FormConfig.OwnerEmailFields))
	fmt.Printf("  Division names: %d\n", len(setup.FormConfig.DivisionNames))
	fmt.Printf("  ~~divisions entries: %d\n", len(setup.FormConfig.Divisions))
	for _, d := range setup.FormConfig.Divisions {
		fmt.Printf("    %s\n", d)
	}

	for name, value := range setup.FormConfig.SelectFields {
		fmt.Printf("  Select: %s = %s\n", name, value)
	}

	fmt.Println()
	for key, value := range setup.FormConfig.OwnerEmailFields {
		fmt.Printf("  OwnerEmail: %s = %s\n", key, value)
	}

	// Example: SetPeriodMatchups usage (commented out to avoid hitting live server)
	//
	// // Swap period 1 matchups: make team A play team B, team C play team D, etc.
	// newMatchups := []models.MatchupPair{
	//     {AwayTeamID: "teamA_id", HomeTeamID: "teamB_id"},
	//     {AwayTeamID: "teamC_id", HomeTeamID: "teamD_id"},
	// }
	// err = client.SetPeriodMatchups(setup, 1, newMatchups)
	// if err != nil {
	//     log.Fatalf("Failed to set period matchups: %v", err)
	// }
	// fmt.Println("Period 1 matchups updated successfully!")

	fmt.Println("\nDone!")
}
