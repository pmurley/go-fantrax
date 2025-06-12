package main

import (
	"fmt"
	"github.com/pmurley/go-fantrax/auth_client"
	"log"
	"os"
	"regexp"

	"github.com/pmurley/go-fantrax/models"
)

// stripHTML removes HTML tags from a string
func stripHTML(html string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	return re.ReplaceAllString(html, "")
}

func main() {
	// Get league ID from environment variable or use default
	leagueID := os.Getenv("FANTRAX_LEAGUE_ID")
	if leagueID == "" {
		leagueID = "q8lydqf5m4u30rca" // Default from the example
	}

	// Create client with caching enabled
	client := auth_client.NewClient(leagueID, true)

	// Example 1: Get my team's roster for current period
	fmt.Println("=== Fetching My Team's Current Roster ===")
	myRoster, err := client.GetMyTeamRosterInfo("")
	if err != nil {
		log.Fatalf("Failed to get my team roster: %v", err)
	}

	// Display team info
	fmt.Printf("\nTeam Information:\n")
	fmt.Printf("  Owner: %s\n", myRoster.TeamInfo.OwnerName)
	fmt.Printf("  Record: %s\n", myRoster.TeamInfo.Record)
	fmt.Printf("  Rank: %s\n", myRoster.TeamInfo.Rank)
	fmt.Printf("  Claim Budget: $%.2f\n", myRoster.ClaimBudget)

	// Display active roster summary
	fmt.Printf("\nActive Roster (%d players):\n", len(myRoster.ActiveRoster))
	for i, player := range myRoster.ActiveRoster {
		if i >= 5 { // Just show first 5 players
			fmt.Printf("  ... and %d more\n", len(myRoster.ActiveRoster)-5)
			break
		}
		fmt.Printf("  %2d. %-20s (age %2d) %-6s %s\n",
			i+1,
			player.Name,
			player.Age,
			player.TeamShortName,
			stripHTML(player.PosShortNames))

		// Show next game if available
		if player.NextGame != nil {
			fmt.Printf("      Next: %s %s", player.NextGame.Opponent, player.NextGame.DateTime)
			if player.NextGame.ProbablePitcher != nil {
				fmt.Printf(" vs %s", player.NextGame.ProbablePitcher.Name)
			}
			fmt.Println()
		}
	}

	// Display reserve roster summary
	fmt.Printf("\nReserve Roster (%d players):\n", len(myRoster.ReserveRoster))
	for i, player := range myRoster.ReserveRoster {
		if i >= 3 { // Just show first 3 reserves
			fmt.Printf("  ... and %d more\n", len(myRoster.ReserveRoster)-3)
			break
		}
		fmt.Printf("  %2d. %-20s (age %2d) %-6s %s\n",
			i+1,
			player.Name,
			player.Age,
			player.TeamShortName,
			stripHTML(player.PosShortNames))
	}

	// Display league teams
	fmt.Printf("\nLeague Teams (%d total):\n", len(myRoster.LeagueTeams))
	for i, team := range myRoster.LeagueTeams {
		if i >= 5 { // Just show first 5 teams
			fmt.Printf("  ... and %d more\n", len(myRoster.LeagueTeams)-5)
			break
		}
		fmt.Printf("  - %s (%s)\n", team.Name, team.ShortName)
	}

	// Example 2: Get a specific team's roster for a specific period
	fmt.Println("\n=== Fetching Specific Team's Roster (Period 73) ===")
	teamID := "j12wv4h4m6iakb28"
	period := "73"

	specificRoster, err := client.GetTeamRosterInfo(period, teamID)
	if err != nil {
		log.Fatalf("Failed to get specific team roster: %v", err)
	}

	// Comprehensive team information
	fmt.Printf("\n--- Team Information ---\n")
	fmt.Printf("Team ID:      %s\n", teamID)
	fmt.Printf("Owner:        %s\n", specificRoster.TeamInfo.OwnerName)
	fmt.Printf("Record:       %s\n", specificRoster.TeamInfo.Record)
	fmt.Printf("Rank:         %s\n", specificRoster.TeamInfo.Rank)
	fmt.Printf("Claim Budget: $%.2f\n", specificRoster.ClaimBudget)
	if specificRoster.TeamInfo.LogoURL != "" {
		fmt.Printf("Logo URL:     %s\n", specificRoster.TeamInfo.LogoURL)
	}

	// Active roster details
	fmt.Printf("\n--- Active Roster (%d players) ---\n", len(specificRoster.ActiveRoster))
	for i, player := range specificRoster.ActiveRoster {
		fmt.Printf("\n%d. %s (ID: %s)\n", i+1, player.Name, player.PlayerID)
		fmt.Printf("   Age: %d | Team: %s (%s)\n", player.Age, player.TeamName, player.TeamShortName)
		fmt.Printf("   Positions: %s | Primary: %s | Rostered at: %s\n",
			stripHTML(player.PosShortNames), player.PrimaryPosition, player.RosterPosition)

		// Player attributes
		attrs := []string{}
		if player.Rookie {
			attrs = append(attrs, "Rookie")
		}
		if player.MinorsEligible {
			attrs = append(attrs, "Minors Eligible")
		}
		if len(attrs) > 0 {
			fmt.Printf("   Attributes: %v\n", attrs)
		}

		// URLs
		if player.HeadshotURL != "" || player.URLName != "" {
			fmt.Printf("   URLs: headshot=%s, fantrax=%s\n", player.HeadshotURL, player.URLName)
		}

		// Next game
		if player.NextGame != nil {
			fmt.Printf("   Next Game: %s @ %s (Event ID: %s)\n",
				player.NextGame.Opponent, player.NextGame.DateTime, player.NextGame.EventID)
			if player.NextGame.ProbablePitcher != nil {
				fmt.Printf("   Probable Pitcher: %s", player.NextGame.ProbablePitcher.Name)
				if len(player.NextGame.ProbablePitcher.Stats) > 0 {
					fmt.Printf(" (")
					first := true
					for k, v := range player.NextGame.ProbablePitcher.Stats {
						if !first {
							fmt.Printf(", ")
						}
						fmt.Printf("%s: %s", k, v)
						first = false
					}
					fmt.Printf(")")
				}
				fmt.Println()
			}
		}

		// Stats
		printPlayerStats(player.Stats, "   ")
	}

	// Reserve roster details
	fmt.Printf("\n--- Reserve Roster (%d players) ---\n", len(specificRoster.ReserveRoster))
	for i, player := range specificRoster.ReserveRoster {
		fmt.Printf("\n%d. %s (ID: %s)\n", i+1, player.Name, player.PlayerID)
		fmt.Printf("   Age: %d | Team: %s (%s)\n", player.Age, player.TeamName, player.TeamShortName)
		fmt.Printf("   Positions: %s | Primary: %s | Rostered at: %s\n",
			stripHTML(player.PosShortNames), player.PrimaryPosition, player.RosterPosition)

		// Player attributes
		attrs := []string{}
		if player.Rookie {
			attrs = append(attrs, "Rookie")
		}
		if player.MinorsEligible {
			attrs = append(attrs, "Minors Eligible")
		}
		if len(attrs) > 0 {
			fmt.Printf("   Attributes: %v\n", attrs)
		}

		// Stats
		printPlayerStats(player.Stats, "   ")
	}

	// Injured Reserve roster details
	fmt.Printf("\n--- Injured Reserve (%d players) ---\n", len(specificRoster.InjuredReserve))
	for i, player := range specificRoster.InjuredReserve {
		fmt.Printf("\n%d. %s (ID: %s)\n", i+1, player.Name, player.PlayerID)
		fmt.Printf("   Age: %d | Team: %s (%s)\n", player.Age, player.TeamName, player.TeamShortName)
		fmt.Printf("   Positions: %s | Primary: %s | Rostered at: %s\n",
			stripHTML(player.PosShortNames), player.PrimaryPosition, player.RosterPosition)

		// Player attributes
		attrs := []string{}
		if player.Rookie {
			attrs = append(attrs, "Rookie")
		}
		if player.MinorsEligible {
			attrs = append(attrs, "Minors Eligible")
		}
		if len(attrs) > 0 {
			fmt.Printf("   Attributes: %v\n", attrs)
		}

		// Stats
		printPlayerStats(player.Stats, "   ")
	}

	// Minors roster details
	fmt.Printf("\n--- Minors Roster (%d players) ---\n", len(specificRoster.MinorsRoster))
	for i, player := range specificRoster.MinorsRoster {
		fmt.Printf("\n%d. %s (ID: %s)\n", i+1, player.Name, player.PlayerID)
		fmt.Printf("   Age: %d | Team: %s (%s)\n", player.Age, player.TeamName, player.TeamShortName)
		fmt.Printf("   Positions: %s | Primary: %s | Rostered at: %s\n",
			stripHTML(player.PosShortNames), player.PrimaryPosition, player.RosterPosition)

		// Player attributes
		attrs := []string{}
		if player.Rookie {
			attrs = append(attrs, "Rookie")
		}
		if player.MinorsEligible {
			attrs = append(attrs, "Minors Eligible")
		}
		if len(attrs) > 0 {
			fmt.Printf("   Attributes: %v\n", attrs)
		}

		// Stats
		printPlayerStats(player.Stats, "   ")
	}

	// League teams summary
	fmt.Printf("\n--- League Teams (%d total) ---\n", len(specificRoster.LeagueTeams))
	for _, team := range specificRoster.LeagueTeams {
		commissionerTag := ""
		if team.Commissioner {
			commissionerTag = " [Commissioner]"
		}
		fmt.Printf("  - %s (%s) ID: %s%s\n", team.Name, team.ShortName, team.ID, commissionerTag)
		if team.LogoURL256 != "" {
			fmt.Printf("    Logo: %s\n", team.LogoURL256)
		}
	}
}

// printPlayerStats displays strongly-typed player stats
func printPlayerStats(stats *models.PlayerStats, indent string) {
	if stats == nil {
		return
	}

	if stats.Batting != nil {
		printBattingStats(stats.Batting, indent)
	} else if stats.Pitching != nil {
		printPitchingStats(stats.Pitching, indent)
	}
}

// printBattingStats displays batting statistics
func printBattingStats(stats *models.BattingStats, indent string) {
	fmt.Printf("%sBatting Stats:\n", indent)

	// Core offensive stats
	if stats.FantasyPointsPerGame != nil {
		fmt.Printf("%s  Fantasy Points/Game: %.2f\n", indent, *stats.FantasyPointsPerGame)
	}
	if stats.GamesPlayed != nil {
		fmt.Printf("%s  Games Played: %d\n", indent, *stats.GamesPlayed)
	}
	if stats.AtBats != nil {
		fmt.Printf("%s  At Bats: %d\n", indent, *stats.AtBats)
	}
	if stats.Hits != nil {
		fmt.Printf("%s  Hits: %d\n", indent, *stats.Hits)
	}
	if stats.Runs != nil {
		fmt.Printf("%s  Runs: %d\n", indent, *stats.Runs)
	}
	if stats.Doubles != nil {
		fmt.Printf("%s  Doubles: %d\n", indent, *stats.Doubles)
	}
	if stats.Triples != nil {
		fmt.Printf("%s  Triples: %d\n", indent, *stats.Triples)
	}
	if stats.HomeRuns != nil {
		fmt.Printf("%s  Home Runs: %d\n", indent, *stats.HomeRuns)
	}
	if stats.RBI != nil {
		fmt.Printf("%s  RBI: %d\n", indent, *stats.RBI)
	}
	if stats.Walks != nil {
		fmt.Printf("%s  Walks: %d\n", indent, *stats.Walks)
	}
	if stats.Strikeouts != nil {
		fmt.Printf("%s  Strikeouts: %d\n", indent, *stats.Strikeouts)
	}
	if stats.StolenBases != nil {
		fmt.Printf("%s  Stolen Bases: %d\n", indent, *stats.StolenBases)
	}
	if stats.CaughtStealing != nil {
		fmt.Printf("%s  Caught Stealing: %d\n", indent, *stats.CaughtStealing)
	}
	if stats.HitByPitch != nil {
		fmt.Printf("%s  Hit By Pitch: %d\n", indent, *stats.HitByPitch)
	}
	if stats.GIDP != nil {
		fmt.Printf("%s  Grounded Into Double Plays: %d\n", indent, *stats.GIDP)
	}

	// Defensive stats
	if stats.Errors != nil {
		fmt.Printf("%s  Errors: %d\n", indent, *stats.Errors)
	}
	if stats.Assists != nil {
		fmt.Printf("%s  Assists: %d\n", indent, *stats.Assists)
	}
	if stats.AssistsOutfield != nil {
		fmt.Printf("%s  Assists (Outfield): %d\n", indent, *stats.AssistsOutfield)
	}
	if stats.Putouts != nil {
		fmt.Printf("%s  Putouts: %d\n", indent, *stats.Putouts)
	}
	if stats.PutoutsOutfield != nil {
		fmt.Printf("%s  Putouts (Outfield): %d\n", indent, *stats.PutoutsOutfield)
	}
	if stats.DoublePlays != nil {
		fmt.Printf("%s  Double Plays: %d\n", indent, *stats.DoublePlays)
	}

	// Catcher-specific stats
	if stats.PassedBalls != nil {
		fmt.Printf("%s  Passed Balls: %d\n", indent, *stats.PassedBalls)
	}
	if stats.CaughtStealingAgainst != nil {
		fmt.Printf("%s  Caught Stealing Against: %d\n", indent, *stats.CaughtStealingAgainst)
	}
	if stats.StolenBasesAgainst != nil {
		fmt.Printf("%s  Stolen Bases Against: %d\n", indent, *stats.StolenBasesAgainst)
	}
}

// printPitchingStats displays pitching statistics
func printPitchingStats(stats *models.PitchingStats, indent string) {
	fmt.Printf("%sPitching Stats:\n", indent)

	// Core pitching stats
	if stats.FantasyPointsPerGame != nil {
		fmt.Printf("%s  Fantasy Points/Game: %.2f\n", indent, *stats.FantasyPointsPerGame)
	}
	if stats.GamesPlayed != nil {
		fmt.Printf("%s  Games Played: %d\n", indent, *stats.GamesPlayed)
	}
	if stats.InningsPitched != nil {
		fmt.Printf("%s  Innings Pitched: %.1f\n", indent, *stats.InningsPitched)
	}
	if stats.ERA != nil {
		fmt.Printf("%s  ERA: %.2f\n", indent, *stats.ERA)
	}
	if stats.Strikeouts != nil {
		fmt.Printf("%s  Strikeouts: %d\n", indent, *stats.Strikeouts)
	}
	if stats.WalksAllowed != nil {
		fmt.Printf("%s  Walks Allowed: %d\n", indent, *stats.WalksAllowed)
	}
	if stats.HitsAllowed != nil {
		fmt.Printf("%s  Hits Allowed: %d\n", indent, *stats.HitsAllowed)
	}
	if stats.EarnedRuns != nil {
		fmt.Printf("%s  Earned Runs: %d\n", indent, *stats.EarnedRuns)
	}

	// Starter-specific stats
	if stats.QualityStarts != nil {
		fmt.Printf("%s  Quality Starts: %d\n", indent, *stats.QualityStarts)
	}
	if stats.CompleteGames != nil {
		fmt.Printf("%s  Complete Games: %d\n", indent, *stats.CompleteGames)
	}
	if stats.Shutouts != nil {
		fmt.Printf("%s  Shutouts: %d\n", indent, *stats.Shutouts)
	}

	// Relief-specific stats
	if stats.Saves != nil {
		fmt.Printf("%s  Saves: %d\n", indent, *stats.Saves)
	}
	if stats.BlownSaves != nil {
		fmt.Printf("%s  Blown Saves: %d\n", indent, *stats.BlownSaves)
	}
	if stats.Holds != nil {
		fmt.Printf("%s  Holds: %d\n", indent, *stats.Holds)
	}

	// Other pitching stats
	if stats.WildPitches != nil {
		fmt.Printf("%s  Wild Pitches: %d\n", indent, *stats.WildPitches)
	}
	if stats.HitBatsmen != nil {
		fmt.Printf("%s  Hit Batsmen: %d\n", indent, *stats.HitBatsmen)
	}
	if stats.Balks != nil {
		fmt.Printf("%s  Balks: %d\n", indent, *stats.Balks)
	}
	if stats.Pickoffs != nil {
		fmt.Printf("%s  Pickoffs: %d\n", indent, *stats.Pickoffs)
	}
}
