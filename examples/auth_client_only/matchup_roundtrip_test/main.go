// matchup_roundtrip_test performs a live round-trip test of SetPeriodMatchups:
//
//  1. GET  — fetch current matchups, record original state
//  2. POST — swap two teams in period 1
//  3. GET  — verify the swap took effect
//  4. POST — revert to original matchups
//  5. GET  — verify the reversion matches the original
//
// Each step prints detailed output. The script aborts on any failure
// and leaves matchups in whatever state they were in at the time of failure.
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pmurley/go-fantrax/auth_client"
	"github.com/pmurley/go-fantrax/models"
)

const testPeriod = 1

func main() {
	leagueID := os.Getenv("FANTRAX_LEAGUE_ID")
	if leagueID == "" {
		log.Fatal("Please set FANTRAX_LEAGUE_ID environment variable")
	}

	client, err := auth_client.NewClient(leagueID, false)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	fmt.Printf("Logged in as: %s\n\n", client.UserInfo.Username)

	// ── Step 1: Fetch original matchups ──────────────────────────────────
	fmt.Println("=== Step 1: Fetch original matchups ===")
	setup, err := client.GetLeagueSetupMatchups()
	if err != nil {
		log.Fatalf("Failed to get league setup matchups: %v", err)
	}

	originalPairs := setup.Matchups[testPeriod]
	if len(originalPairs) == 0 {
		log.Fatalf("Period %d has no matchups", testPeriod)
	}

	fmt.Printf("Period %d has %d matchup pairs\n", testPeriod, len(originalPairs))
	printMatchups(setup, originalPairs)

	// Save a deep copy of the original for later comparison
	savedOriginal := copyPairs(originalPairs)

	// Find first two non-bye matchups to swap
	var swapIdx []int
	for i, p := range originalPairs {
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

	// ── Step 2: Swap and POST ────────────────────────────────────────────
	fmt.Printf("\n=== Step 2: Swap matchups %d and %d, then POST ===\n", i, j)

	swappedPairs := copyPairs(originalPairs)
	swappedPairs[i].AwayTeamID, swappedPairs[j].AwayTeamID = swappedPairs[j].AwayTeamID, swappedPairs[i].AwayTeamID

	fmt.Printf("Swapping away teams:\n")
	fmt.Printf("  Matchup %d: %s vs %s  ->  %s vs %s\n", i,
		teamName(setup, originalPairs[i].AwayTeamID), teamName(setup, originalPairs[i].HomeTeamID),
		teamName(setup, swappedPairs[i].AwayTeamID), teamName(setup, swappedPairs[i].HomeTeamID))
	fmt.Printf("  Matchup %d: %s vs %s  ->  %s vs %s\n", j,
		teamName(setup, originalPairs[j].AwayTeamID), teamName(setup, originalPairs[j].HomeTeamID),
		teamName(setup, swappedPairs[j].AwayTeamID), teamName(setup, swappedPairs[j].HomeTeamID))

	fmt.Println("POSTing swap...")
	err = client.SetPeriodMatchups(setup, testPeriod, swappedPairs)
	if err != nil {
		log.Fatalf("FAILED to POST swap: %v", err)
	}
	fmt.Println("POST returned 302 — success!")

	// Brief pause to let the server process
	time.Sleep(2 * time.Second)

	// ── Step 3: Verify the swap ──────────────────────────────────────────
	fmt.Println("\n=== Step 3: Fetch matchups and verify swap ===")
	setup2, err := client.GetLeagueSetupMatchups()
	if err != nil {
		log.Fatalf("Failed to re-fetch matchups: %v", err)
	}

	verifyPairs := setup2.Matchups[testPeriod]
	printMatchups(setup2, verifyPairs)

	if !matchupsEqual(verifyPairs, swappedPairs) {
		fmt.Println("\nEXPECTED:")
		printMatchups(setup2, swappedPairs)
		fmt.Println("\nGOT:")
		printMatchups(setup2, verifyPairs)
		log.Fatal("VERIFICATION FAILED: fetched matchups do not match the swapped matchups")
	}
	fmt.Println("VERIFIED: matchups match the swap!")

	// ── Step 4: Revert to original ───────────────────────────────────────
	fmt.Println("\n=== Step 4: Revert to original matchups ===")
	fmt.Println("POSTing revert...")
	err = client.SetPeriodMatchups(setup2, testPeriod, savedOriginal)
	if err != nil {
		log.Fatalf("FAILED to POST revert: %v", err)
	}
	fmt.Println("POST returned 302 — success!")

	time.Sleep(2 * time.Second)

	// ── Step 5: Verify the reversion ─────────────────────────────────────
	fmt.Println("\n=== Step 5: Fetch matchups and verify reversion ===")
	setup3, err := client.GetLeagueSetupMatchups()
	if err != nil {
		log.Fatalf("Failed to re-fetch matchups after revert: %v", err)
	}

	revertPairs := setup3.Matchups[testPeriod]
	printMatchups(setup3, revertPairs)

	if !matchupsEqual(revertPairs, savedOriginal) {
		fmt.Println("\nEXPECTED (original):")
		printMatchups(setup3, savedOriginal)
		fmt.Println("\nGOT:")
		printMatchups(setup3, revertPairs)
		log.Fatal("VERIFICATION FAILED: matchups do not match the original after revert")
	}
	fmt.Println("VERIFIED: matchups match the original!")

	fmt.Println("\n=== ALL STEPS PASSED ===")
}

func printMatchups(setup *models.LeagueSetupMatchups, pairs []models.MatchupPair) {
	for idx, pair := range pairs {
		away := teamName(setup, pair.AwayTeamID)
		home := teamName(setup, pair.HomeTeamID)
		if pair.HomeTeamID == "-1" {
			fmt.Printf("  %2d: %-20s  BYE\n", idx, away)
		} else {
			fmt.Printf("  %2d: %-20s vs  %s\n", idx, away, home)
		}
	}
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

func copyPairs(pairs []models.MatchupPair) []models.MatchupPair {
	out := make([]models.MatchupPair, len(pairs))
	copy(out, pairs)
	return out
}

func matchupsEqual(a, b []models.MatchupPair) bool {
	if len(a) != len(b) {
		return false
	}
	for idx := range a {
		if a[idx].AwayTeamID != b[idx].AwayTeamID || a[idx].HomeTeamID != b[idx].HomeTeamID {
			return false
		}
	}
	return true
}
