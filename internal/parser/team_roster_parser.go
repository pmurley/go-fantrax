package parser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pmurley/go-fantrax/models"
)

// ParseTeamRosterResponse parses the raw API response into a simplified TeamRoster
func ParseTeamRosterResponse(data []byte) (*models.TeamRoster, error) {
	var response models.TeamRosterResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Responses) == 0 {
		return nil, fmt.Errorf("no responses in data")
	}

	rosterData := response.Responses[0].Data
	roster := &models.TeamRoster{
		LeagueTeams: rosterData.FantasyTeams,
	}

	// Extract team info
	roster.TeamInfo = extractTeamInfo(rosterData)

	// Extract claim budget
	roster.ClaimBudget = extractClaimBudget(rosterData.MiscData)

	// Parse roster tables - they are organized by player type, not roster status
	var allPlayers []models.RosterPlayer

	// Parse all tables (position players and pitchers)
	for _, table := range rosterData.Tables {
		players := parseRosterTable(table)
		allPlayers = append(allPlayers, players...)
	}

	// Separate players by roster status based on statusId
	for _, player := range allPlayers {
		switch player.Status {
		case "Active":
			roster.ActiveRoster = append(roster.ActiveRoster, player)
		case "Reserve":
			roster.ReserveRoster = append(roster.ReserveRoster, player)
		case "Injured Reserve":
			roster.InjuredReserve = append(roster.InjuredReserve, player)
		case "Minors":
			roster.MinorsRoster = append(roster.MinorsRoster, player)
		}
	}

	return roster, nil
}

func extractTeamInfo(data models.TeamRosterResponseData) models.TeamInfo {
	info := models.TeamInfo{
		LogoURL: data.Settings.LogoURL,
	}

	// Extract from TeamHeadingInfo
	if data.TeamHeadingInfo.Owners.Value != "" {
		info.OwnerName = data.TeamHeadingInfo.Owners.Value
	}
	if data.TeamHeadingInfo.H2HRecord.Value != "" {
		info.Record = data.TeamHeadingInfo.H2HRecord.Value
	}
	if data.TeamHeadingInfo.Rank.Value != "" {
		info.Rank = data.TeamHeadingInfo.Rank.Value
	}

	// Extract team ID from MyTeamIDs
	if len(data.MyTeamIDs) > 0 {
		info.TeamID = data.MyTeamIDs[0]
	}

	return info
}

func extractClaimBudget(miscData models.MiscData) float64 {
	for _, info := range miscData.SalaryInfo.Info {
		if info.Key == "claimBudget" {
			budget, err := strconv.ParseFloat(info.Value, 64)
			if err == nil {
				return budget
			}
		}
	}
	return 0
}

func parseRosterTable(table models.RosterTable) []models.RosterPlayer {
	var players []models.RosterPlayer

	for _, row := range table.Rows {
		// Skip empty roster slots
		if row.IsEmptyRosterSlot || row.Scorer.Name == "" {
			continue
		}

		player := models.RosterPlayer{
			PlayerID:        row.Scorer.ScorerID,
			Name:            row.Scorer.Name,
			ShortName:       row.Scorer.ShortName,
			TeamName:        row.Scorer.TeamName,
			TeamShortName:   row.Scorer.TeamShortName,
			TeamID:          row.Scorer.TeamID,
			Positions:       row.Scorer.PosIDs,
			PrimaryPosition: row.Scorer.PrimaryPosID,
			PosShortNames:   row.Scorer.PosShortNames,
			HeadshotURL:     row.Scorer.HeadshotURL,
			URLName:         row.Scorer.URLName,
			Rookie:          row.Scorer.Rookie,
			MinorsEligible:  row.Scorer.MinorsEligible,
			Status:          mapStatusID(row.StatusID),
			RosterPosition:  row.PosID,
			Stats:           &models.PlayerStats{},
		}

		// Extract age from first cell
		if len(row.Cells) > 0 {
			age, err := strconv.Atoi(row.Cells[0].Content)
			if err == nil {
				player.Age = age
			}
		}

		// Parse stats from cells
		player.Stats = parsePlayerStats(row.Cells, table.Header.Cells, row.Scorer.PosIDs)

		// Extract next game info
		player.NextGame = extractNextGame(row.Cells)

		players = append(players, player)
	}

	return players
}

func parsePlayerStats(cells []models.Cell, columns []models.Column, positionIDs []string) *models.PlayerStats {
	stats := &models.PlayerStats{}

	// Determine if this is a pitcher based on position IDs
	isPitching := isPitcher(positionIDs)

	if isPitching {
		stats.Pitching = &models.PitchingStats{}
	} else {
		stats.Batting = &models.BattingStats{}
	}

	// Parse stats from each column
	for i, cell := range cells {
		if i >= len(columns) || cell.Content == "" {
			continue
		}

		col := columns[i]
		// Skip non-stat columns (age, opponent)
		if col.Key == "age" || col.Key == "opponent" {
			continue
		}

		// Parse based on column key
		if isPitching {
			parsePitchingStatByKey(col.Key, cell.Content, stats.Pitching)
		} else {
			parseBattingStatByKey(col.Key, cell.Content, stats.Batting)
		}
	}

	return stats
}

// isPitcher determines if a player is a pitcher based on their position IDs
func isPitcher(positionIDs []string) bool {
	for _, posID := range positionIDs {
		if posID == "015" || posID == "016" { // SP or RP
			return true
		}
	}
	return false
}

// mapStatusID converts status ID to readable status string
func mapStatusID(statusID string) string {
	switch statusID {
	case "1":
		return "Active"
	case "2":
		return "Reserve"
	case "3":
		return "Injured Reserve"
	case "9":
		return "Minors"
	default:
		return "Unknown"
	}
}

// Helper functions to parse stat values
func parseIntStat(value string) *int {
	if value == "" || value == "-" {
		return nil
	}
	if intVal, err := strconv.Atoi(value); err == nil {
		return &intVal
	}
	return nil
}

func parseFloatStat(value string) *float64 {
	if value == "" || value == "-" {
		return nil
	}
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return &floatVal
	}
	return nil
}

// parseBattingStatByKey maps column keys to batting stat fields
func parseBattingStatByKey(key, value string, stats *models.BattingStats) {
	switch key {
	case "fptsPerGame":
		stats.FantasyPointsPerGame = parseFloatStat(value)
	case "10#0010#-1": // AB
		stats.AtBats = parseIntStat(value)
	case "10#0170#-1": // H
		stats.Hits = parseIntStat(value)
	case "10#0330#-1": // R
		stats.Runs = parseIntStat(value)
	case "10#0070#-1": // 2B
		stats.Doubles = parseIntStat(value)
	case "10#0420#-1": // 3B
		stats.Triples = parseIntStat(value)
	case "10#0200#-1": // HR
		stats.HomeRuns = parseIntStat(value)
	case "10#0310#-1": // RBI
		stats.RBI = parseIntStat(value)
	case "10#0430#-1": // BB
		stats.Walks = parseIntStat(value)
	case "10#0400#-1": // SO
		stats.Strikeouts = parseIntStat(value)
	case "10#0380#-1": // SB
		stats.StolenBases = parseIntStat(value)
	case "10#0040#-1": // CS
		stats.CaughtStealing = parseIntStat(value)
	case "10#0150#-1": // HBP
		stats.HitByPitch = parseIntStat(value)
	case "10#0130#-1": // GIDP
		stats.GIDP = parseIntStat(value)
	case "10#0090#-1": // E
		stats.Errors = parseIntStat(value)
	case "10#0050#-1": // CSA
		stats.CaughtStealingAgainst = parseIntStat(value)
	case "10#0065#-1": // DP
		stats.DoublePlays = parseIntStat(value)
	case "10#0005#-1": // A
		stats.Assists = parseIntStat(value)
	case "10#0006#-1": // AOF
		stats.AssistsOutfield = parseIntStat(value)
	case "10#029g#-1": // PO
		stats.Putouts = parseIntStat(value)
	case "10#029h#-1": // POOF
		stats.PutoutsOutfield = parseIntStat(value)
	case "10#0390#-1": // SBA
		stats.StolenBasesAgainst = parseIntStat(value)
	case "10#0280#-1": // PB
		stats.PassedBalls = parseIntStat(value)
	case "10#0100#-1": // GP
		stats.GamesPlayed = parseIntStat(value)
	}
}

// parsePitchingStatByKey maps column keys to pitching stat fields
func parsePitchingStatByKey(key, value string, stats *models.PitchingStats) {
	switch key {
	case "fptsPerGame":
		stats.FantasyPointsPerGame = parseFloatStat(value)
	case "20#0220#-1": // IP
		stats.InningsPitched = parseFloatStat(value)
	case "20#0300#-1": // QS
		stats.QualityStarts = parseIntStat(value)
	case "20#0360#-1": // SV
		stats.Saves = parseIntStat(value)
	case "20#0030#-1": // BS
		stats.BlownSaves = parseIntStat(value)
	case "20#0190#-1": // HLD
		stats.Holds = parseIntStat(value)
	case "20#0060#-1": // CG
		stats.CompleteGames = parseIntStat(value)
	case "20#0180#-1": // H
		stats.HitsAllowed = parseIntStat(value)
	case "20#0080#-1": // ER
		stats.EarnedRuns = parseIntStat(value)
	case "20#0440#-1": // BB
		stats.WalksAllowed = parseIntStat(value)
	case "20#0410#-1": // K
		stats.Strikeouts = parseIntStat(value)
	case "20#0490#-1": // ERA
		stats.ERA = parseFloatStat(value)
	case "20#0025#-1": // BK
		stats.Balks = parseIntStat(value)
	case "20#0450#-1": // WP
		stats.WildPitches = parseIntStat(value)
	case "20#0140#-1": // HB
		stats.HitBatsmen = parseIntStat(value)
	case "20#0370#-1": // SHO
		stats.Shutouts = parseIntStat(value)
	case "20#0291#-1": // PKO
		stats.Pickoffs = parseIntStat(value)
	case "20#0100#-1": // GP
		stats.GamesPlayed = parseIntStat(value)
	}
}

func extractNextGame(cells []models.Cell) *models.GameInfo {
	// Usually the second cell contains game info
	if len(cells) > 1 && cells[1].EventID != "" {
		gameInfo := &models.GameInfo{
			EventID: cells[1].EventID,
		}

		// Parse game content (e.g., "@PIT<br/>Thu 5:40PM")
		content := cells[1].Content
		parts := strings.Split(content, "<br/>")
		if len(parts) > 0 {
			gameInfo.Opponent = strings.TrimSpace(parts[0])
		}
		if len(parts) > 1 {
			gameInfo.DateTime = strings.TrimSpace(parts[1])
		}

		// Extract pitcher info from popover
		if cells[1].PopOver != nil {
			gameInfo.ProbablePitcher = &models.PitcherInfo{
				Name:      cells[1].PopOver.Scorer.Name,
				ShortName: cells[1].PopOver.Scorer.ShortName,
				Stats:     parsePitcherStats(cells[1].PopOver.Content),
			}
		}

		return gameInfo
	}

	return nil
}

func parsePitcherStats(content string) map[string]string {
	stats := make(map[string]string)

	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]+>`)
	cleanContent := re.ReplaceAllString(content, "")

	// Split by spaces and parse key-value pairs
	parts := strings.Fields(cleanContent)
	for i := 0; i < len(parts)-1; i += 2 {
		key := strings.TrimSpace(parts[i])
		if i+1 < len(parts) {
			value := strings.TrimSpace(parts[i+1])
			stats[key] = value
		}
	}

	return stats
}
