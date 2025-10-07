package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/pmurley/go-fantrax"
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
	fmt.Println("Fetching league info...")
	leagueInfo, err := client.GetLeagueInfo(leagueID)
	if err != nil {
		log.Fatalf("Failed to get league info: %v", err)
	}

	// Generate markdown content
	markdown := generateMarkdown(leagueInfo)

	// Write to file
	filename := fmt.Sprintf("league_info_%s.md", leagueID)
	err = os.WriteFile(filename, []byte(markdown), 0644)
	if err != nil {
		log.Fatalf("Failed to write markdown file: %v", err)
	}

	fmt.Printf("Successfully wrote league info to %s\n", filename)
}

func generateMarkdown(info *fantrax.LeagueInfo) string {
	var sb strings.Builder

	// Title
	sb.WriteString("# League Information\n\n")

	// Draft Settings
	sb.WriteString("## Draft Settings\n\n")
	sb.WriteString(fmt.Sprintf("- **Draft Type**: %s\n\n", info.DraftType))

	// Pool Settings
	sb.WriteString("## Pool Settings\n\n")
	sb.WriteString(fmt.Sprintf("- **Player Source Type**: %s\n", info.PoolSettings.PlayerSourceType))
	sb.WriteString(fmt.Sprintf("- **Duplicate Player Type**: %s\n\n", info.PoolSettings.DuplicatePlayerType))

	// Roster Configuration
	sb.WriteString("## Roster Configuration\n\n")
	sb.WriteString(fmt.Sprintf("- **Max Total Players**: %d\n", info.RosterInfo.MaxTotalPlayers))
	sb.WriteString(fmt.Sprintf("- **Max Active Players**: %d\n", info.RosterInfo.MaxTotalActivePlayers))
	sb.WriteString(fmt.Sprintf("- **Max Reserve Players**: %d\n\n", info.RosterInfo.MaxTotalReservePlayers))

	// Position Constraints
	if len(info.RosterInfo.PositionConstraints) > 0 {
		sb.WriteString("### Position Constraints\n\n")
		sb.WriteString("| Position | Max Active |\n")
		sb.WriteString("|----------|------------|\n")
		for pos, constraint := range info.RosterInfo.PositionConstraints {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", pos, constraint.MaxActive))
		}
		sb.WriteString("\n")
	}

	// Teams
	if len(info.TeamInfo) > 0 {
		sb.WriteString("## Teams\n\n")

		// Convert map to slice for sorting
		teams := make([]fantrax.TeamInfo, 0, len(info.TeamInfo))
		for _, team := range info.TeamInfo {
			teams = append(teams, team)
		}

		// Sort teams by league, division, then name
		sort.Slice(teams, func(i, j int) bool {
			// Extract league and division from division string (e.g., "AL E" -> "AL", "E")
			divI := teams[i].Division
			divJ := teams[j].Division

			// Handle empty divisions
			if divI == "" {
				divI = "ZZ ZZ" // Sort empty divisions last
			}
			if divJ == "" {
				divJ = "ZZ ZZ"
			}

			partsI := strings.Split(divI, " ")
			partsJ := strings.Split(divJ, " ")

			leagueI := partsI[0]
			leagueJ := partsJ[0]

			// Compare league first
			if leagueI != leagueJ {
				return leagueI < leagueJ
			}

			// If same league, compare division
			if len(partsI) > 1 && len(partsJ) > 1 {
				if partsI[1] != partsJ[1] {
					return partsI[1] < partsJ[1]
				}
			}

			// If same division, sort by team name
			return teams[i].Name < teams[j].Name
		})

		sb.WriteString("| Team Name | Division | Team ID |\n")
		sb.WriteString("|-----------|----------|----------|\n")
		for _, team := range teams {
			division := team.Division
			if division == "" {
				division = "—"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | `%s` |\n", team.Name, division, team.ID))
		}
		sb.WriteString("\n")
	}

	// Scoring System
	sb.WriteString("## Scoring System\n\n")
	sb.WriteString(fmt.Sprintf("- **Type**: %s\n\n", info.ScoringSystem.Type))

	// Scoring Categories
	if info.ScoringSystem.ScoringCategories.HITTING != nil || info.ScoringSystem.ScoringCategories.PITCHING != nil {
		sb.WriteString("### Scoring Categories\n\n")

		if info.ScoringSystem.ScoringCategories.HITTING != nil {
			sb.WriteString("#### Hitting\n\n")
			for categoryID, categoryData := range info.ScoringSystem.ScoringCategories.HITTING {
				if name, ok := categoryData["name"]; ok {
					sb.WriteString(fmt.Sprintf("- **%s** (ID: `%s`)\n", name, categoryID))
				}
			}
			sb.WriteString("\n")
		}

		if info.ScoringSystem.ScoringCategories.PITCHING != nil {
			sb.WriteString("#### Pitching\n\n")
			for categoryID, categoryData := range info.ScoringSystem.ScoringCategories.PITCHING {
				if name, ok := categoryData["name"]; ok {
					sb.WriteString(fmt.Sprintf("- **%s** (ID: `%s`)\n", name, categoryID))
				}
			}
			sb.WriteString("\n")
		}
	}

	// Scoring Category Settings
	if len(info.ScoringSystem.ScoringCategorySettings) > 0 {
		sb.WriteString("### Scoring Configuration\n\n")

		for _, setting := range info.ScoringSystem.ScoringCategorySettings {
			if setting.Group.Name != "" {
				sb.WriteString(fmt.Sprintf("#### %s\n\n", setting.Group.Name))

				// Group configs by category for better readability
				categoryMap := make(map[string][]fantrax.ScoringConfig)
				for _, config := range setting.Configs {
					category := config.ScoringCategory.ShortName
					categoryMap[category] = append(categoryMap[category], config)
				}

				// Get sorted category names
				categories := make([]string, 0, len(categoryMap))
				for cat := range categoryMap {
					categories = append(categories, cat)
				}
				sort.Strings(categories)

				sb.WriteString("| Category | Position | Points |\n")
				sb.WriteString("|----------|----------|--------|\n")

				for _, category := range categories {
					configs := categoryMap[category]

					// Sort configs: Default first, then alphabetically by position
					sort.Slice(configs, func(i, j int) bool {
						if configs[i].Position.ShortName == "Default" {
							return true
						}
						if configs[j].Position.ShortName == "Default" {
							return false
						}
						return configs[i].Position.ShortName < configs[j].Position.ShortName
					})

					// Write the category rows
					for idx, config := range configs {
						categoryName := category
						if idx > 0 {
							categoryName = "" // Only show category name on first row
						}

						positionName := config.Position.ShortName
						if positionName == "Default" {
							positionName = "All"
						}

						sb.WriteString(fmt.Sprintf("| %s | %s | %.2f |\n",
							categoryName,
							positionName,
							config.Points))
					}
				}
				sb.WriteString("\n")
			}
		}
	}

	// Matchups/Schedule
	if len(info.Matchups) > 0 {
		sb.WriteString("## Schedule\n\n")

		// Show first few matchup periods as examples
		maxPeriods := 5
		if len(info.Matchups) < maxPeriods {
			maxPeriods = len(info.Matchups)
		}

		for i := 0; i < maxPeriods; i++ {
			matchup := info.Matchups[i]
			sb.WriteString(fmt.Sprintf("### Period %d\n\n", matchup.Period))

			if len(matchup.MatchupList) > 0 {
				sb.WriteString("| Home Team | Away Team |\n")
				sb.WriteString("|-----------|------------|\n")

				for _, m := range matchup.MatchupList {
					sb.WriteString(fmt.Sprintf("| %s | %s |\n", m.Home.Name, m.Away.Name))
				}
				sb.WriteString("\n")
			}
		}

		if len(info.Matchups) > maxPeriods {
			sb.WriteString(fmt.Sprintf("*... and %d more matchup periods*\n\n", len(info.Matchups)-maxPeriods))
		}
	}

	return sb.String()
}
