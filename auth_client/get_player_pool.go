package auth_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/pmurley/go-fantrax/models"
)

var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

const (
	// MaxPlayersPerPage is the maximum number of players Fantrax returns per page
	MaxPlayersPerPage = 5000

	// StatusFilterAll includes all players (rostered and available)
	StatusFilterAll = "ALL"

	// StatusFilterAvailable includes only free agents and waiver players
	StatusFilterAvailable = "ALL_AVAILABLE"
)

// GetPlayerPoolRequest represents the request payload for getPlayerStats
type GetPlayerPoolRequest struct {
	StatusOrTeamFilter string `json:"statusOrTeamFilter,omitempty"`
	MaxResultsPerPage  int    `json:"maxResultsPerPage,omitempty"`
	PageNumber         string `json:"pageNumber,omitempty"` // Must be string per Fantrax API
}

// PlayerPoolOption is a functional option for configuring GetPlayerPool
type PlayerPoolOption func(*playerPoolConfig)

type playerPoolConfig struct {
	statusFilter string
}

// WithStatusFilter sets the status filter for the player pool query
// Use StatusFilterAll for all players or StatusFilterAvailable for only available players
func WithStatusFilter(filter string) PlayerPoolOption {
	return func(c *playerPoolConfig) {
		c.statusFilter = filter
	}
}

// GetPlayerPool fetches all players in the league's player pool
// By default, fetches ALL players (including rostered). Use WithStatusFilter(StatusFilterAvailable)
// to get only free agents and waiver players.
// This handles pagination automatically to retrieve all players.
func (c *Client) GetPlayerPool(opts ...PlayerPoolOption) ([]models.PoolPlayer, error) {
	// Apply options
	config := &playerPoolConfig{
		statusFilter: StatusFilterAll, // Default to all players
	}
	for _, opt := range opts {
		opt(config)
	}

	var allPlayers []models.PoolPlayer
	pageNumber := 1
	totalPages := 1 // Will be updated after first request

	for pageNumber <= totalPages {
		response, err := c.getPlayerPoolPage(config.statusFilter, pageNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page %d: %w", pageNumber, err)
		}

		if len(response.Responses) == 0 {
			return nil, fmt.Errorf("no responses in player pool response for page %d", pageNumber)
		}

		data := response.Responses[0].Data
		totalPages = data.PaginatedResultSet.TotalNumPages

		// Parse players from this page
		players, err := parseStatsTable(data.StatsTable, buildColumnIndex(data.TableHeader))
		if err != nil {
			return nil, fmt.Errorf("failed to parse players on page %d: %w", pageNumber, err)
		}

		allPlayers = append(allPlayers, players...)
		pageNumber++
	}

	return allPlayers, nil
}

// GetPlayerPoolRaw fetches a single page of the raw player pool response without parsing
func (c *Client) GetPlayerPoolRaw(statusFilter string, pageNumber int) (*models.PlayerPoolResponse, error) {
	return c.getPlayerPoolPage(statusFilter, pageNumber)
}

// getPlayerPoolPage fetches a single page of the player pool
func (c *Client) getPlayerPoolPage(statusFilter string, pageNumber int) (*models.PlayerPoolResponse, error) {
	requestData := GetPlayerPoolRequest{
		StatusOrTeamFilter: statusFilter,
		MaxResultsPerPage:  MaxPlayersPerPage,
		PageNumber:         strconv.Itoa(pageNumber),
	}

	fullRequest := map[string]interface{}{
		"msgs": []FantraxMessage{
			{
				Method: "getPlayerStats",
				Data:   requestData,
			},
		},
		"uiv":    3,
		"refUrl": fmt.Sprintf("https://www.fantrax.com/fantasy/league/%s/players", c.LeagueID),
		"dt":     0,
		"at":     0,
		"av":     "0.0",
		"tz":     c.getTimezone(),
		"v":      "179.0.1",
	}

	jsonStr, err := json.Marshal(fullRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://www.fantrax.com/fxpa/req?leagueId="+c.LeagueID, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response models.PlayerPoolResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// getTimezone returns the user's timezone or UTC as default
func (c *Client) getTimezone() string {
	if c.UserInfo != nil && c.UserInfo.Timezone != "" {
		return c.UserInfo.Timezone
	}
	return "UTC"
}

// columnIndex maps a stats-table column identifier to its cell index.
type columnIndex map[string]int

// buildColumnIndex indexes a player-pool table header by every stable
// identifier Fantrax exposes (key, sortType, shortName) so cells can be
// located by name instead of a fragile fixed position. Fantrax varies the
// column set (e.g. it drops %Drafted/ADP outside the draft season), which
// previously caused Age and other cells to be silently skipped.
func buildColumnIndex(h models.TableHeader) columnIndex {
	idx := make(columnIndex, len(h.Cells)*3)
	for i, col := range h.Cells {
		for _, id := range []string{col.Key, col.SortType, col.ShortName} {
			if id == "" {
				continue
			}
			if _, exists := idx[id]; !exists {
				idx[id] = i
			}
		}
	}
	return idx
}

// find returns the cell index for the first identifier that resolves, or -1.
func (c columnIndex) find(ids ...string) int {
	for _, id := range ids {
		if i, ok := c[id]; ok {
			return i
		}
	}
	return -1
}

// parseStatsTable converts raw stats table entries to PoolPlayer structs
func parseStatsTable(entries []models.StatsTableEntry, cols columnIndex) ([]models.PoolPlayer, error) {
	players := make([]models.PoolPlayer, 0, len(entries))

	for _, entry := range entries {
		player, err := parseStatsTableEntry(entry, cols)
		if err != nil {
			// Log warning but continue with other players
			continue
		}
		players = append(players, player)
	}

	return players, nil
}

// parseStatsTableEntry converts a single stats table entry to a PoolPlayer
func parseStatsTableEntry(entry models.StatsTableEntry, cols columnIndex) (models.PoolPlayer, error) {
	scorer := entry.Scorer
	cells := entry.Cells

	player := models.PoolPlayer{
		// Core identification
		PlayerID:  scorer.ScorerID,
		Name:      scorer.Name,
		ShortName: scorer.ShortName,
		URLName:   scorer.URLName,

		// MLB team info
		MLBTeamName:      scorer.TeamName,
		MLBTeamShortName: scorer.TeamShortName,
		MLBTeamID:        scorer.TeamID,

		// Player attributes
		Rookie:         scorer.Rookie,
		MinorsEligible: scorer.MinorsEligible,

		// Position info
		Positions:       scorer.PosIDs,
		PositionsNoFlex: scorer.PosIDsNoFlex,
		PrimaryPosID:    scorer.PrimaryPosID,
		DefaultPosID:    scorer.DefaultPosID,
		PosShortNames:   stripHTML(scorer.PosShortNames),
		MultiPositions:  entry.MultiPositions,

		// Rankings (from scorer)
		Rank: scorer.Rank,

		// Media
		HeadshotURL: scorer.HeadshotURL,

		// Icons
		Icons: scorer.Icons,
	}

	// Parse actions
	for _, action := range entry.Actions {
		player.Actions = append(player.Actions, action.TypeID)
	}

	// Parse cells by header-mapped position. Fantrax changes which columns
	// are present (e.g. %Drafted/ADP only appear during the draft season),
	// so each cell is located by its column identifier rather than a fixed
	// index/count. The old fixed `len(cells) >= 10` assumption silently
	// dropped Age (and every other cell) once Fantrax went to 8 columns.
	cell := func(ids ...string) (models.StatsTableCell, bool) {
		i := cols.find(ids...)
		if i < 0 || i >= len(cells) {
			return models.StatsTableCell{}, false
		}
		return cells[i], true
	}

	// Rank already comes from scorer.Rank.

	// Status - FA, W, or team abbreviation
	if c, ok := cell("status", "STATUS", "Sta"); ok {
		player.FantasyStatus = c.Content
		player.FantasyTeamID = c.TeamID
		player.FantasyTeamName = c.ToolTip
	}

	// Age
	if c, ok := cell("age", "AGE", "Age"); ok {
		if age, err := strconv.Atoi(strings.TrimSpace(c.Content)); err == nil {
			player.Age = age
		}
	}

	// Next opponent (may contain HTML like "@SF<br/>Wed 8:05PM")
	if c, ok := cell("opponent", "Opp"); ok {
		player.NextOpponent = stripHTML(c.Content)
	}

	// Fantasy Points
	if c, ok := cell("fpts", "SCORE", "FPts"); ok {
		player.FantasyPoints = parseFloat(c.Content)
	}

	// Fantasy Points Per Game
	if c, ok := cell("fptsPerGame", "FPTS_PER_GAME", "FP/G"); ok {
		player.FantasyPointsPerG = parseFloat(c.Content)
	}

	// % Drafted (absent outside draft season)
	if c, ok := cell("PERCENT_DRAFTED", "%D"); ok {
		player.PercentDrafted = parseFloat(c.Content)
	}

	// ADP (absent outside draft season)
	if c, ok := cell("ADP"); ok {
		player.ADP = parseFloat(c.Content)
	}

	// % Rostered (may have % suffix)
	if c, ok := cell("OVERVIEW_PERCENT_OWNED_2", "Ros"); ok {
		player.PercentRostered = parsePercentage(c.Content)
	}

	// Roster Change (may have +/- prefix and % suffix)
	if c, ok := cell("OVERVIEW_PLUS_MINUS_PERCENT_OWNED_2", "+/-"); ok {
		player.RosterChange = parsePercentage(c.Content)
	}

	return player, nil
}

// parseFloat parses a string to float64, returning 0 on error
func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// parsePercentage parses a percentage string like "97%" or "+1%" to float64
func parsePercentage(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "%")
	s = strings.TrimPrefix(s, "+")
	if s == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// stripHTML removes HTML tags from a string
func stripHTML(s string) string {
	return htmlTagRegex.ReplaceAllString(s, "")
}
