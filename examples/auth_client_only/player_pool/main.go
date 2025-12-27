package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pmurley/go-fantrax/auth_client"
)

func main() {
	// Get league ID from environment variable
	leagueID := os.Getenv("FANTRAX_LEAGUE_ID")
	if leagueID == "" {
		log.Fatal("FANTRAX_LEAGUE_ID must be set")
	}

	// Create client (caching disabled for fresh data)
	fmt.Println("Creating auth client...")
	client, err := auth_client.NewClient(leagueID, false)
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}

	// Fetch all players
	fmt.Println("Fetching all players in the player pool...")
	players, err := client.GetPlayerPool()
	if err != nil {
		log.Fatalf("Failed to get player pool: %v", err)
	}

	fmt.Printf("\nTotal players retrieved: %d\n", len(players))

	// Count by fantasy status
	faCount := 0
	rosteredCount := 0
	waiverCount := 0
	for _, p := range players {
		switch p.FantasyStatus {
		case "FA":
			faCount++
		case "W":
			waiverCount++
		default:
			rosteredCount++
		}
	}
	fmt.Printf("  Free Agents: %d\n", faCount)
	fmt.Printf("  On Waivers: %d\n", waiverCount)
	fmt.Printf("  Rostered: %d\n", rosteredCount)

	// Show top 10 players by rank
	fmt.Println("\n=== TOP 10 PLAYERS ===")
	for i := 0; i < 10 && i < len(players); i++ {
		p := players[i]
		fmt.Printf("%2d. %-25s %-4s Age:%-2d FPts:%-7.1f FP/G:%-5.2f Status:%-4s",
			p.Rank, p.Name, p.MLBTeamShortName, p.Age, p.FantasyPoints, p.FantasyPointsPerG, p.FantasyStatus)
		if p.FantasyTeamName != "" && p.FantasyStatus != "FA" {
			fmt.Printf(" (%s)", p.FantasyTeamName)
		}
		fmt.Println()
	}

	// Show some free agents
	fmt.Println("\n=== TOP FREE AGENTS (by rank) ===")
	shown := 0
	for _, p := range players {
		if p.FantasyStatus == "FA" {
			fmt.Printf("%3d. %-25s %-4s Age:%-2d FPts:%-7.1f FP/G:%-5.2f Pos:%s\n",
				p.Rank, p.Name, p.MLBTeamShortName, p.Age, p.FantasyPoints, p.FantasyPointsPerG, p.PosShortNames)
			shown++
			if shown >= 10 {
				break
			}
		}
	}

	// Show a sample player with all fields
	fmt.Println("\n=== SAMPLE PLAYER (full details) ===")
	if len(players) > 0 {
		p := players[8000]
		fmt.Printf("PlayerID:        %s\n", p.PlayerID)
		fmt.Printf("Name:            %s\n", p.Name)
		fmt.Printf("ShortName:       %s\n", p.ShortName)
		fmt.Printf("URLName:         %s\n", p.URLName)
		fmt.Printf("MLBTeamName:     %s\n", p.MLBTeamName)
		fmt.Printf("MLBTeamShortName:%s\n", p.MLBTeamShortName)
		fmt.Printf("MLBTeamID:       %s\n", p.MLBTeamID)
		fmt.Printf("Age:             %d\n", p.Age)
		fmt.Printf("Rookie:          %v\n", p.Rookie)
		fmt.Printf("MinorsEligible:  %v\n", p.MinorsEligible)
		fmt.Printf("Positions:       %v\n", p.Positions)
		fmt.Printf("PositionsNoFlex: %v\n", p.PositionsNoFlex)
		fmt.Printf("PrimaryPosID:    %s\n", p.PrimaryPosID)
		fmt.Printf("DefaultPosID:    %s\n", p.DefaultPosID)
		fmt.Printf("PosShortNames:   %s\n", p.PosShortNames)
		fmt.Printf("MultiPositions:  %s\n", p.MultiPositions)
		fmt.Printf("FantasyStatus:   %s\n", p.FantasyStatus)
		fmt.Printf("FantasyTeamID:   %s\n", p.FantasyTeamID)
		fmt.Printf("FantasyTeamName: %s\n", p.FantasyTeamName)
		fmt.Printf("Rank:            %d\n", p.Rank)
		fmt.Printf("FantasyPoints:   %.1f\n", p.FantasyPoints)
		fmt.Printf("FantasyPointsPerG:%.2f\n", p.FantasyPointsPerG)
		fmt.Printf("PercentDrafted:  %.0f\n", p.PercentDrafted)
		fmt.Printf("ADP:             %.1f\n", p.ADP)
		fmt.Printf("PercentRostered: %.0f\n", p.PercentRostered)
		fmt.Printf("RosterChange:    %.0f\n", p.RosterChange)
		fmt.Printf("NextOpponent:    %s\n", p.NextOpponent)
		fmt.Printf("HeadshotURL:     %s\n", p.HeadshotURL)
		fmt.Printf("Icons:           %v\n", p.Icons)
		fmt.Printf("Actions:         %v\n", p.Actions)
	}

	// Test with StatusFilterAvailable
	fmt.Println("\n=== TESTING StatusFilterAvailable ===")
	availablePlayers, err := client.GetPlayerPool(auth_client.WithStatusFilter(auth_client.StatusFilterAvailable))
	if err != nil {
		log.Fatalf("Failed to get available players: %v", err)
	}
	fmt.Printf("Available players only: %d\n", len(availablePlayers))
}
