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

	// Create authenticated client (must be commissioner account)
	client, err := auth_client.NewClient(leagueID, false)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== Commissioner Add/Drop Example ===\n")
	fmt.Printf("League ID: %s\n", leagueID)
	fmt.Printf("Logged in as: %s\n", client.UserInfo.Username)
	fmt.Printf("Target Team ID: %s\n\n", targetTeamID)

	// Get current period
	period, err := client.GetCurrentPeriod()
	if err != nil {
		log.Fatalf("Failed to get current period: %v", err)
	}
	fmt.Printf("Current period: %d\n\n", period)

	// Track if we need to clean up
	var addedPlayerID string

	// Example 1: Commissioner Add
	playerToAddID := os.Getenv("FANTRAX_PLAYER_ID_TO_ADD")
	if playerToAddID != "" {
		fmt.Println("=== Example 1: Commissioner Add ===")
		fmt.Printf("Adding player %s to team %s\n\n", playerToAddID, targetTeamID)

		// Read roster before add
		fmt.Println("Reading roster before add...")
		rosterBefore, err := client.GetTeamRosterInfo(fmt.Sprintf("%d", period), targetTeamID)
		if err != nil {
			log.Fatalf("Failed to get roster before add: %v", err)
		}

		// Count all players on roster
		allPlayersBefore := append(rosterBefore.ActiveRoster, rosterBefore.ReserveRoster...)
		allPlayersBefore = append(allPlayersBefore, rosterBefore.InjuredReserve...)
		allPlayersBefore = append(allPlayersBefore, rosterBefore.MinorsRoster...)
		playerCountBefore := len(allPlayersBefore)
		fmt.Printf("Roster before: %d players\n", playerCountBefore)

		// Check if player is already on roster
		playerExists := false
		for _, p := range allPlayersBefore {
			if p.PlayerID == playerToAddID {
				playerExists = true
				fmt.Printf("⚠️  Warning: Player %s (%s) is already on the roster\n", p.Name, p.PlayerID)
				break
			}
		}
		fmt.Println()

		if !playerExists {
			// Sleep before API call
			time.Sleep(1 * time.Second)

			// Add player to minors using convenience function
			// This automatically detects the current period and appropriate position
			fmt.Println("Executing commissioner add...")
			response, err := client.CommissionerAddToMinors(
				targetTeamID,
				playerToAddID,
			)
			if err != nil {
				log.Fatalf("Failed to add player: %v", err)
			}

			// Check response
			if response.IsSuccess() {
				fmt.Printf("✓ Success: %s\n", response.GenericMessage)
				if len(response.DetailMessages) > 0 {
					fmt.Println("Details:")
					for _, msg := range response.DetailMessages {
						fmt.Printf("  - %s\n", msg)
					}
				}
				fmt.Printf("Transaction ID: %s\n", response.TransactionID)
				// Track that we added this player for cleanup
				addedPlayerID = playerToAddID
			} else if response.IsError() {
				fmt.Printf("✗ Error: %s\n", response.GenericMessage)
				if len(response.DetailMessages) > 0 {
					fmt.Println("Details:")
					for _, msg := range response.DetailMessages {
						fmt.Printf("  - %s\n", msg)
					}
				}
			}
			fmt.Println()

			// Sleep before verification read
			fmt.Println("Waiting 1 second before verification...")
			time.Sleep(1 * time.Second)

			// Read roster after add
			fmt.Println("Reading roster after add...")
			rosterAfter, err := client.GetTeamRosterInfo(fmt.Sprintf("%d", period), targetTeamID)
			if err != nil {
				log.Fatalf("Failed to get roster after add: %v", err)
			}

			// Count all players on roster
			allPlayersAfter := append(rosterAfter.ActiveRoster, rosterAfter.ReserveRoster...)
			allPlayersAfter = append(allPlayersAfter, rosterAfter.InjuredReserve...)
			allPlayersAfter = append(allPlayersAfter, rosterAfter.MinorsRoster...)
			playerCountAfter := len(allPlayersAfter)
			fmt.Printf("Roster after: %d players\n", playerCountAfter)

			// Verify player was added
			playerFound := false
			for _, p := range allPlayersAfter {
				if p.PlayerID == playerToAddID {
					playerFound = true
					fmt.Printf("✓ Verified: Player %s (%s) is now on the roster\n", p.Name, p.PlayerID)
					fmt.Printf("  Status: %s\n", p.Status)
					break
				}
			}

			if !playerFound {
				fmt.Printf("✗ Verification failed: Player %s not found on roster after add\n", playerToAddID)
			}

			if playerCountAfter != playerCountBefore+1 {
				fmt.Printf("⚠️  Warning: Expected %d players, got %d\n", playerCountBefore+1, playerCountAfter)
			}
			fmt.Println()
		}
	} else {
		fmt.Println("=== Example 1: Commissioner Add ===")
		fmt.Println("Skipped: Set FANTRAX_PLAYER_ID_TO_ADD to test adding a player")
		fmt.Println("Example: export FANTRAX_PLAYER_ID_TO_ADD=03pp9\n")
	}

	// Example 2: Commissioner Drop
	playerToDropID := os.Getenv("FANTRAX_PLAYER_ID_TO_DROP")
	if playerToDropID != "" {
		fmt.Println("=== Example 2: Commissioner Drop ===")
		fmt.Printf("Dropping player %s from team %s\n\n", playerToDropID, targetTeamID)

		// Read roster before drop
		fmt.Println("Reading roster before drop...")
		rosterBefore, err := client.GetTeamRosterInfo(fmt.Sprintf("%d", period), targetTeamID)
		if err != nil {
			log.Fatalf("Failed to get roster before drop: %v", err)
		}

		// Count all players on roster
		allPlayersBefore := append(rosterBefore.ActiveRoster, rosterBefore.ReserveRoster...)
		allPlayersBefore = append(allPlayersBefore, rosterBefore.InjuredReserve...)
		allPlayersBefore = append(allPlayersBefore, rosterBefore.MinorsRoster...)
		playerCountBefore := len(allPlayersBefore)
		fmt.Printf("Roster before: %d players\n", playerCountBefore)

		// Check if player is on roster
		playerExists := false
		var playerName string
		for _, p := range allPlayersBefore {
			if p.PlayerID == playerToDropID {
				playerExists = true
				playerName = p.Name
				fmt.Printf("Found player: %s (%s) on roster\n", p.Name, p.PlayerID)
				fmt.Printf("  Current position: %s, status: %s\n", p.RosterPosition, p.Status)
				break
			}
		}

		if !playerExists {
			fmt.Printf("⚠️  Warning: Player %s is not on the roster\n", playerToDropID)
		}
		fmt.Println()

		if playerExists {
			// Sleep before API call
			time.Sleep(1 * time.Second)

			// Drop player using convenience function
			// This automatically detects the current period
			fmt.Println("Executing commissioner drop...")
			response, err := client.CommissionerDropFromRoster(
				targetTeamID,
				playerToDropID,
			)
			if err != nil {
				log.Fatalf("Failed to drop player: %v", err)
			}

			// Check response
			if response.IsSuccess() {
				fmt.Printf("✓ Success: %s\n", response.GenericMessage)
				if len(response.DetailMessages) > 0 {
					fmt.Println("Details:")
					for _, msg := range response.DetailMessages {
						fmt.Printf("  - %s\n", msg)
					}
				}
				fmt.Printf("Transaction ID: %s\n", response.TransactionID)
			} else if response.IsError() {
				fmt.Printf("✗ Error: %s\n", response.GenericMessage)
				if len(response.DetailMessages) > 0 {
					fmt.Println("Details:")
					for _, msg := range response.DetailMessages {
						fmt.Printf("  - %s\n", msg)
					}
				}
			}
			fmt.Println()

			// Sleep before verification read
			fmt.Println("Waiting 1 second before verification...")
			time.Sleep(1 * time.Second)

			// Read roster after drop
			fmt.Println("Reading roster after drop...")
			rosterAfter, err := client.GetTeamRosterInfo(fmt.Sprintf("%d", period), targetTeamID)
			if err != nil {
				log.Fatalf("Failed to get roster after drop: %v", err)
			}

			// Count all players on roster
			allPlayersAfter := append(rosterAfter.ActiveRoster, rosterAfter.ReserveRoster...)
			allPlayersAfter = append(allPlayersAfter, rosterAfter.InjuredReserve...)
			allPlayersAfter = append(allPlayersAfter, rosterAfter.MinorsRoster...)
			playerCountAfter := len(allPlayersAfter)
			fmt.Printf("Roster after: %d players\n", playerCountAfter)

			// Verify player was dropped
			playerFound := false
			for _, p := range allPlayersAfter {
				if p.PlayerID == playerToDropID {
					playerFound = true
					break
				}
			}

			if !playerFound {
				fmt.Printf("✓ Verified: Player %s (%s) has been removed from the roster\n", playerName, playerToDropID)
			} else {
				fmt.Printf("✗ Verification failed: Player %s (%s) is still on roster after drop\n", playerName, playerToDropID)
			}

			if playerCountAfter != playerCountBefore-1 {
				fmt.Printf("⚠️  Warning: Expected %d players, got %d\n", playerCountBefore-1, playerCountAfter)
			}
			fmt.Println()
		}
	} else {
		fmt.Println("=== Example 2: Commissioner Drop ===")
		fmt.Println("Skipped: Set FANTRAX_PLAYER_ID_TO_DROP to test dropping a player")
		fmt.Println("Example: export FANTRAX_PLAYER_ID_TO_DROP=03pp9\n")
	}

	// Cleanup: Drop any player that was added during testing
	if addedPlayerID != "" {
		fmt.Println("=== Cleanup ===")
		fmt.Printf("Dropping player %s to restore original roster state...\n", addedPlayerID)
		time.Sleep(1 * time.Second)

		cleanupResp, err := client.CommissionerDropFromRoster(targetTeamID, addedPlayerID)
		if err != nil {
			fmt.Printf("⚠️  Cleanup failed: %v\n", err)
		} else if cleanupResp.IsSuccess() {
			fmt.Println("✓ Player dropped - roster restored")
		} else {
			fmt.Printf("⚠️  Cleanup failed: %s\n", cleanupResp.GenericMessage)
		}
		fmt.Println()
	}

	// Summary
	fmt.Println("=== Summary ===")
	fmt.Println("This example demonstrates commissioner add/drop operations.")
	fmt.Println("Note: These operations require a commissioner account.\n")
	fmt.Println("Available position constants:")
	fmt.Printf("  - auth_client.PosC (Catcher): %s\n", auth_client.PosC)
	fmt.Printf("  - auth_client.Pos1B (First Base): %s\n", auth_client.Pos1B)
	fmt.Printf("  - auth_client.Pos3B (Third Base): %s\n", auth_client.Pos3B)
	fmt.Printf("  - auth_client.PosSS (Shortstop): %s\n", auth_client.PosSS)
	fmt.Printf("  - auth_client.PosUtil (Utility): %s\n", auth_client.PosUtil)
	fmt.Printf("  - auth_client.PosSP (Starting Pitcher): %s\n", auth_client.PosSP)
	fmt.Printf("  - auth_client.PosRP (Relief Pitcher): %s\n", auth_client.PosRP)
	fmt.Println("\nAvailable status constants:")
	fmt.Printf("  - auth_client.StatusActive (Active): %s\n", auth_client.StatusActive)
	fmt.Printf("  - auth_client.StatusReserve (Reserve): %s\n", auth_client.StatusReserve)
	fmt.Printf("  - auth_client.StatusMinors (Minors): %s\n", auth_client.StatusMinors)
	fmt.Println("\n✓ Example completed!")
}
