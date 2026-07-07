package auth_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// GamesPerPositionResponse is the raw top-level Fantrax response for
// getTeamRosterInfo?view=GAMES_PER_POS (the "Min/Max" tab of the Team
// Roster screen).
type GamesPerPositionResponse struct {
	Responses []struct {
		Data GamesPerPositionResponseData `json:"data"`
	} `json:"responses"`
}

// GamesPerPositionResponseData holds the two tables the GAMES_PER_POS view
// returns: per-fielding-position games played, and per-scoring-category
// totals (e.g. "Games Started - Pitching (GS)").
type GamesPerPositionResponseData struct {
	GamePlayedPerPosData struct {
		TableData []PositionCountRow `json:"tableData"`
	} `json:"gamePlayedPerPosData"`
	ScMinMaxData struct {
		TableData []CategoryLimitRow `json:"tableData"`
	} `json:"scMinMaxData"`
}

// PositionCountRow is one row of the per-position games-played table.
type PositionCountRow struct {
	Pos      string `json:"pos"`
	PosShort string `json:"posShort"`
	GP       int    `json:"gp"`
	Min      string `json:"min"` // numeric string, or "No min"
	Max      string `json:"max"` // numeric string, or "No max"
}

// CategoryLimitRow is one row of the per-scoring-category totals table.
type CategoryLimitRow struct {
	ScoringCategory string `json:"scoringCategory"`
	Total           string `json:"total"`
	Min             string `json:"min"` // numeric string, or "No min"
	Max             string `json:"max"` // numeric string, or "No max"
}

// PositionCount is a single fielding position's games-played count for a
// scoring period, with its configured min/max (nil = no limit configured).
type PositionCount struct {
	Name      string
	ShortName string
	GP        int
	Min       *int
	Max       *int
}

// CategoryLimit is a single scoring category's (e.g. "Games Started -
// Pitching (GS)") total and configured min/max for a scoring period (nil =
// no limit configured).
type CategoryLimit struct {
	Category string
	Total    int
	Min      *int
	Max      *int
}

// GamesPerPosition is the parsed "Min/Max" tab of the Team Roster screen:
// per-position games-played counts plus per-scoring-category totals, each
// with their configured min/max.
type GamesPerPosition struct {
	Positions      []PositionCount
	CategoryLimits []CategoryLimit
}

// parseMinMax converts a Fantrax min/max cell to *int. Fantrax returns
// either a numeric string ("10") or the literal sentinel "No min"/"No max"
// when the position/category has no configured limit.
func parseMinMax(s string) *int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &n
}

// ProcessGamesPerPosition converts the raw API response into the typed
// GamesPerPosition result.
func ProcessGamesPerPosition(raw *GamesPerPositionResponse) (*GamesPerPosition, error) {
	if len(raw.Responses) == 0 {
		return nil, fmt.Errorf("no response data found")
	}
	data := raw.Responses[0].Data
	result := &GamesPerPosition{}
	for _, row := range data.GamePlayedPerPosData.TableData {
		result.Positions = append(result.Positions, PositionCount{
			Name:      row.Pos,
			ShortName: row.PosShort,
			GP:        row.GP,
			Min:       parseMinMax(row.Min),
			Max:       parseMinMax(row.Max),
		})
	}
	for _, row := range data.ScMinMaxData.TableData {
		total, _ := strconv.Atoi(row.Total)
		result.CategoryLimits = append(result.CategoryLimits, CategoryLimit{
			Category: row.ScoringCategory,
			Total:    total,
			Min:      parseMinMax(row.Min),
			Max:      parseMinMax(row.Max),
		})
	}
	return result, nil
}

// GetTeamRosterPositionCounts fetches and parses the "Min/Max" tab
// (getTeamRosterInfo?view=GAMES_PER_POS) for the given team and scoring
// period. scoringPeriod is the weekly Scoring Period number — the same
// numbering GetStandings(WithStandingsView(StandingsViewSchedule)) returns
// — pass "" for the current period.
func (c *Client) GetTeamRosterPositionCounts(teamID, scoringPeriod string) (*GamesPerPosition, error) {
	data := map[string]string{
		"leagueId": c.LeagueID,
		"teamId":   teamID,
		"view":     "GAMES_PER_POS",
	}
	if scoringPeriod != "" {
		data["scoringPeriod"] = scoringPeriod
	}

	fullRequest := buildFullRequest(
		[]FantraxMessage{{Method: "getTeamRosterInfo", Data: data}},
		fmt.Sprintf("https://www.fantrax.com/fantasy/league/%s/team/roster", c.LeagueID),
	)

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

	body, err := readBody(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var raw GamesPerPositionResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return ProcessGamesPerPosition(&raw)
}
