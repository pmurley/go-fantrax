package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/pmurley/go-fantrax/auth_client"
	"github.com/pmurley/go-fantrax/models"
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

	fmt.Printf("Got %d periods, %d teams, %d divisions\n",
		len(setup.Matchups), len(setup.Teams), len(setup.Divisions))

	// Show the original period 1 matchups
	period := 1
	originalPairs := setup.Matchups[period]
	fmt.Printf("\nOriginal period %d matchups (%d pairs):\n", period, len(originalPairs))
	for i, pair := range originalPairs {
		fmt.Printf("  %2d: %s vs %s\n", i, teamName(setup, pair.AwayTeamID), teamName(setup, pair.HomeTeamID))
	}

	// Swap: take the first two non-bye matchups and swap their away teams.
	// e.g., if matchup 0 is A_B and matchup 1 is C_D, make it C_B and A_D.
	newPairs := make([]models.MatchupPair, len(originalPairs))
	copy(newPairs, originalPairs)

	// Find first two non-bye matchups
	var swapIdx []int
	for i, p := range newPairs {
		if p.HomeTeamID != "-1" {
			swapIdx = append(swapIdx, i)
			if len(swapIdx) == 2 {
				break
			}
		}
	}

	if len(swapIdx) < 2 {
		log.Fatal("Could not find two non-bye matchups to swap")
	}

	i, j := swapIdx[0], swapIdx[1]
	fmt.Printf("\nSwapping away teams between matchup %d and %d:\n", i, j)
	fmt.Printf("  Before: %s vs %s  AND  %s vs %s\n",
		teamName(setup, newPairs[i].AwayTeamID), teamName(setup, newPairs[i].HomeTeamID),
		teamName(setup, newPairs[j].AwayTeamID), teamName(setup, newPairs[j].HomeTeamID))

	newPairs[i].AwayTeamID, newPairs[j].AwayTeamID = newPairs[j].AwayTeamID, newPairs[i].AwayTeamID

	fmt.Printf("  After:  %s vs %s  AND  %s vs %s\n",
		teamName(setup, newPairs[i].AwayTeamID), teamName(setup, newPairs[i].HomeTeamID),
		teamName(setup, newPairs[j].AwayTeamID), teamName(setup, newPairs[j].HomeTeamID))

	// Update the matchups in-place (same as SetPeriodMatchups does before building the form)
	setup.Matchups[period] = newPairs

	// Build the form body (but do NOT send it)
	form := auth_client.BuildFormBody(setup, period)

	// Save as decoded JSON (matching the format of matchup_save_post_decoded.json)
	decoded := make(map[string]interface{})
	for key, values := range form {
		if len(values) == 1 {
			decoded[key] = values[0]
		} else {
			decoded[key] = values
		}
	}

	jsonBytes, err := json.MarshalIndent(decoded, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	outPath := "/Users/pmurley/GolandProjects/go-fantrax/matchup_save_post_generated.json"
	if err := os.WriteFile(outPath, jsonBytes, 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
	fmt.Printf("\nSaved generated POST body to: %s\n", outPath)

	// Also save the raw URL-encoded form for reference
	rawPath := "/Users/pmurley/GolandProjects/go-fantrax/matchup_save_post_generated_raw.txt"
	if err := os.WriteFile(rawPath, []byte(form.Encode()), 0644); err != nil {
		log.Fatalf("Failed to write raw file: %v", err)
	}
	fmt.Printf("Saved raw URL-encoded form to: %s\n", rawPath)

	// Quick summary comparison
	fmt.Printf("\n=== Quick Comparison ===\n")
	fmt.Printf("Generated form has %d unique keys\n", len(form))

	// Count matchups entries
	fmt.Printf("  matchups entries: %d\n", len(form["matchups"]))
	fmt.Printf("  ~~divisions entries: %d\n", len(form["~~divisions"]))

	// List all keys sorted
	var keys []string
	for k := range form {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Printf("\nAll %d form keys:\n", len(keys))
	for _, k := range keys {
		vals := form[k]
		if len(vals) == 1 {
			v := vals[0]
			if len(v) > 80 {
				v = v[:80] + "..."
			}
			fmt.Printf("  %s = %s\n", k, v)
		} else {
			fmt.Printf("  %s = [%d values]\n", k, len(vals))
		}
	}

	fmt.Println("\nDone! Now compare with: diff <(jq -S . matchup_save_post_decoded.json) <(jq -S . matchup_save_post_generated.json)")
}

func teamName(setup *models.LeagueSetupMatchups, teamID string) string {
	if teamID == "-1" {
		return "BYE"
	}
	team := auth_client.GetTeamByID(setup, teamID)
	if team != nil {
		return team.ShortName
	}
	return teamID
}

// sortedKeys returns sorted keys from a string map for deterministic output.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// contains checks if a string is in a slice.
func contains(s []string, target string) bool {
	for _, v := range s {
		if v == target {
			return true
		}
	}
	return false
}

// join helper for display.
func join(parts []string) string {
	return strings.Join(parts, ", ")
}
