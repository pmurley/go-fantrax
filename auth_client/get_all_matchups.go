package auth_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// AllMatchupsResult contains all matchups for a season with team info for lookups
type AllMatchupsResult struct {
	Matchups []Matchup              `json:"matchups"`
	Teams    map[string]FantasyTeam `json:"teams"` // keyed by teamId
}

// GetAllMatchups returns all matchups for the season using the SCHEDULE view
func (c *Client) GetAllMatchups() (*AllMatchupsResult, error) {
	var requestPayload = FantraxRequest{
		Msgs: []FantraxMessage{
			{
				Method: "getStandings",
				Data: map[string]string{
					"leagueId": c.LeagueID,
					"view":     "SCHEDULE",
				},
			},
		},
	}

	jsonStr, err := json.Marshal(requestPayload)
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

	var response StandingsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Responses) == 0 {
		return nil, fmt.Errorf("no response data found")
	}

	responseData := response.Responses[0].Data

	result := &AllMatchupsResult{
		Matchups: make([]Matchup, 0),
		Teams:    responseData.FantasyTeamInfo,
	}

	// Process all matchup tables (H2hPointsBased3 from SCHEDULE view)
	for _, table := range responseData.TableList {
		if table.TableType != "H2hPointsBased3" {
			continue
		}

		period := 0
		date := ""

		// Parse period number from caption (e.g., "Scoring Period 42")
		if strings.HasPrefix(table.Caption, "Scoring Period ") {
			parts := strings.Split(table.Caption, " ")
			if len(parts) >= 3 {
				period, _ = strconv.Atoi(parts[2])
			}
		}

		// Extract date from subCaption (e.g., "(Sat Apr 19, 2025)")
		if len(table.SubCaption) > 2 {
			date = strings.Trim(table.SubCaption, "()")
		}

		for _, row := range table.Rows {
			if len(row.Cells) < 8 {
				continue
			}

			awayTeamID := row.Cells[0].TeamID
			awayPoints, _ := strconv.ParseFloat(row.Cells[1].Content, 64)
			awayAdj, _ := strconv.ParseFloat(row.Cells[2].Content, 64)
			awayTotal, _ := strconv.ParseFloat(row.Cells[3].Content, 64)

			homeTeamID := row.Cells[4].TeamID
			homePoints, _ := strconv.ParseFloat(row.Cells[5].Content, 64)
			homeAdj, _ := strconv.ParseFloat(row.Cells[6].Content, 64)
			homeTotal, _ := strconv.ParseFloat(row.Cells[7].Content, 64)

			matchup := Matchup{
				ScoringPeriod: period,
				Date:          date,
				AwayTeam: MatchTeam{
					TeamID:     awayTeamID,
					Points:     awayPoints,
					Adjustment: awayAdj,
					Total:      awayTotal,
				},
				HomeTeam: MatchTeam{
					TeamID:     homeTeamID,
					Points:     homePoints,
					Adjustment: homeAdj,
					Total:      homeTotal,
				},
			}

			result.Matchups = append(result.Matchups, matchup)
		}
	}

	return result, nil
}
