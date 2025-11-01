package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pmurley/go-fantrax/auth_client"
)

func main() {
	// Get league ID from environment variable
	leagueID := os.Getenv("FANTRAX_LEAGUE_ID")
	if leagueID == "" {
		log.Fatal("Please set FANTRAX_LEAGUE_ID environment variable")
	}

	// Get team ID from environment variable
	targetTeamID := os.Getenv("FANTRAX_TEAM_ID")
	if targetTeamID == "" {
		log.Fatal("Please set FANTRAX_TEAM_ID environment variable")
	}

	// Create authenticated client
	client, err := auth_client.NewClient(leagueID, false)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== Simple Roster Editing Example ===\n")
	fmt.Printf("League ID: %s\n", leagueID)
	fmt.Printf("Logged in as: %s\n\n", client.UserInfo.Username)

	// Create a roster editor (fetches current roster automatically)
	period := 1

	fmt.Println("Fetching roster...")
	editor, err := client.NewRosterEditor(period, targetTeamID, true, false)
	if err != nil {
		log.Fatalf("Failed to create roster editor: %v", err)
	}

	fmt.Printf("Team ID: %s\n", targetTeamID)
	fmt.Printf("Period: %d\n\n", period)

	// Show current roster summary
	activePlayers := editor.GetActivePlayers()
	reservePlayers := editor.GetReservePlayers()
	minorsPlayers := editor.GetMinorsPlayers()

	fmt.Printf("Current Roster:\n")
	fmt.Printf("  Active: %d players\n", len(activePlayers))
	fmt.Printf("  Reserve: %d players\n", len(reservePlayers))
	fmt.Printf("  Minors: %d players\n\n", len(minorsPlayers))

	// Find a player to move (first active player)
	if len(activePlayers) == 0 {
		log.Fatal("No active players found")
	}

	playerToMove := activePlayers[0]
	originalPosition := playerToMove.PositionID
	fmt.Printf("Test player: %s (currently Active at %s)\n\n", playerToMove.Name, originalPosition)

	// Step 1: Move player to Reserve
	fmt.Println("=== Step 1: Moving player to Reserve ===")
	if err := editor.MoveToReserve(playerToMove.PlayerID); err != nil {
		log.Fatalf("Failed to queue change: %v", err)
	}

	// Preview pending changes
	changes := editor.GetPendingChanges()
	fmt.Println("Pending changes:")
	for _, change := range changes {
		fmt.Printf("  - %s\n", change)
	}
	fmt.Println()

	// Apply changes
	fmt.Println("Applying changes...")
	result, err := editor.Apply(true)
	if err != nil {
		log.Fatalf("Failed to apply changes: %v", err)
	}

	if !result.Success {
		log.Fatalf("Changes failed: %s", result.ErrorMessage)
	}

	fmt.Println("✓ Changes applied successfully!")
	if result.IsCommissioner {
		fmt.Println("✓ Commissioner action confirmed")
	}
	if len(result.Changes) > 0 {
		fmt.Printf("API reported changes: %v\n", result.Changes)
	}

	// Wait before verification
	fmt.Println("\nWaiting 3 seconds...")
	time.Sleep(3 * time.Second)

	// Verify the change
	fmt.Println("\n=== Verifying player is now on Reserve ===")
	verifyEditor1, err := client.NewRosterEditor(period, targetTeamID, true, false)
	if err != nil {
		log.Fatalf("Failed to fetch roster for verification: %v", err)
	}

	reservePlayersAfter := verifyEditor1.GetReservePlayers()
	found := false
	for _, p := range reservePlayersAfter {
		if p.PlayerID == playerToMove.PlayerID {
			found = true
			fmt.Printf("✓ Verified: %s is now on Reserve\n", p.Name)
			break
		}
	}
	if !found {
		log.Fatalf("✗ Verification failed: Player not found on Reserve")
	}

	// Wait before next operation
	fmt.Println("\nWaiting 3 seconds...")
	time.Sleep(3 * time.Second)

	// Step 2: Move player back to Active
	fmt.Println("\n=== Step 2: Moving player back to Active ===")
	editor2, err := client.NewRosterEditor(period, targetTeamID, true, false)
	if err != nil {
		log.Fatalf("Failed to create second editor: %v", err)
	}

	if err := editor2.MoveToActive(playerToMove.PlayerID, originalPosition); err != nil {
		log.Fatalf("Failed to queue restore: %v", err)
	}

	fmt.Println("Pending changes:")
	for _, change := range editor2.GetPendingChanges() {
		fmt.Printf("  - %s\n", change)
	}
	fmt.Println()

	fmt.Println("Applying restore...")
	result2, err := editor2.Apply(true)
	if err != nil {
		log.Fatalf("Failed to restore: %v", err)
	}

	if !result2.Success {
		log.Fatalf("Restore failed: %s", result2.ErrorMessage)
	}

	fmt.Println("✓ Player restored successfully!")
	if len(result2.Changes) > 0 {
		fmt.Printf("API reported changes: %v\n", result2.Changes)
	}

	// Wait before final verification
	fmt.Println("\nWaiting 3 seconds...")
	time.Sleep(3 * time.Second)

	// Final verification
	fmt.Println("\n=== Verifying player is back on Active ===")
	verifyEditor2, err := client.NewRosterEditor(period, targetTeamID, true, false)
	if err != nil {
		log.Fatalf("Failed to fetch roster for final verification: %v", err)
	}

	activePlayersAfter := verifyEditor2.GetActivePlayers()
	found = false
	for _, p := range activePlayersAfter {
		if p.PlayerID == playerToMove.PlayerID {
			found = true
			if p.PositionID == originalPosition {
				fmt.Printf("✓ Verified: %s is back on Active at %s\n", p.Name, p.PositionID)
			} else {
				log.Fatalf("✗ Verification failed: Expected position %s, got %s", originalPosition, p.PositionID)
			}
			break
		}
	}
	if !found {
		log.Fatalf("✗ Verification failed: Player not found on Active roster")
	}

	// Summary
	fmt.Println("\n=== Summary ===")
	fmt.Printf("✓ All operations successful\n")
	fmt.Printf("✓ Player: %s\n", playerToMove.Name)
	fmt.Printf("✓ Sequence: Active → Reserve → Active\n")
	fmt.Printf("✓ Position preserved: %s\n", originalPosition)
	fmt.Println("\n✓ Example completed successfully!")
}
