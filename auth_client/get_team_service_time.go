package auth_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/pmurley/go-fantrax/models"
)

// statusIDToRosterStatus maps API status IDs to RosterStatus constants
var statusIDToRosterStatus = map[string]models.RosterStatus{
	"1": models.StatusActive,
	"2": models.StatusReserve,
	"3": models.StatusIR,
	"9": models.StatusMinors,
	"7": models.StatusNotOnTeam,
}

// GetTeamServiceTimeRequest represents the request payload for getTeamServiceTime
type GetTeamServiceTimeRequest struct {
	TeamID string `json:"teamId"`
}

// GetTeamServiceTimeRaw fetches the raw team service time response
func (c *Client) GetTeamServiceTimeRaw(teamID string) (*models.ServiceTimeResponse, error) {
	requestPayload := FantraxRequest{
		Msgs: []FantraxMessage{
			{
				Method: "getTeamServiceTime",
				Data: GetTeamServiceTimeRequest{
					TeamID: teamID,
				},
			},
		},
	}

	// Build refUrl
	refUrl := fmt.Sprintf("https://www.fantrax.com/fantasy/league/%s/team/service-time;teamId=%s", c.LeagueID, teamID)

	// Add common Fantrax request fields
	fullRequest := map[string]interface{}{
		"msgs":   requestPayload.Msgs,
		"uiv":    3,
		"refUrl": refUrl,
		"dt":     0,
		"at":     0,
		"av":     "0.0",
		"tz":     "America/Chicago",
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

	var response models.ServiceTimeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetTeamServiceTime fetches and parses service time data for a team
func (c *Client) GetTeamServiceTime(teamID string) (models.TeamServiceTimeResult, error) {
	rawResponse, err := c.GetTeamServiceTimeRaw(teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw service time: %w", err)
	}

	if len(rawResponse.Responses) == 0 {
		return nil, fmt.Errorf("no responses in service time response")
	}

	serviceTime := rawResponse.Responses[0].Data.ServiceTime
	return parseServiceTime(serviceTime)
}

// parseServiceTime converts raw service time data to a clean structure
func parseServiceTime(st models.ServiceTime) (models.TeamServiceTimeResult, error) {
	result := make(models.TeamServiceTimeResult)

	// Build a map of column index to period number for period columns
	// Headers structure: [Act, Res, IR, Min, ...positions..., 1, 2, 3, ... periodN]
	periodColumns := make(map[int]int) // column index -> period number
	for i, header := range st.Headers {
		// Period columns have shortName as an integer
		switch v := header.ShortName.(type) {
		case float64:
			periodColumns[i] = int(v)
		case int:
			periodColumns[i] = v
		}
	}

	// Process each player row
	for _, row := range st.Rows {
		scorer := row.Scorer
		player := models.PlayerServiceTime{
			ScorerID:         scorer.ScorerID,
			Name:             scorer.Name,
			ShortName:        scorer.ShortName,
			TeamName:         scorer.TeamName,
			TeamShortName:    scorer.TeamShortName,
			Positions:        scorer.PosShortNames,
			IsRookie:         scorer.Rookie,
			IsMinorsEligible: scorer.MinorsEligible,
			PeriodHistory:    make(map[int]models.PeriodStatus),
		}

		// Parse totals from first 4 columns (Act, Res, IR, Min)
		if len(row.Cells) > 0 {
			player.DaysActive = parseIntOrZero(row.Cells[0].Content)
		}
		if len(row.Cells) > 1 {
			player.DaysReserve = parseIntOrZero(row.Cells[1].Content)
		}
		if len(row.Cells) > 2 {
			player.DaysIR = parseIntOrZero(row.Cells[2].Content)
		}
		if len(row.Cells) > 3 {
			player.DaysMinors = parseIntOrZero(row.Cells[3].Content)
		}

		// Parse period history
		for colIdx, periodNum := range periodColumns {
			if colIdx < len(row.Cells) {
				cell := row.Cells[colIdx]
				status := models.StatusNotOnTeam
				if cell.StatusID != "" {
					if mappedStatus, ok := statusIDToRosterStatus[cell.StatusID]; ok {
						status = mappedStatus
					}
				}
				player.PeriodHistory[periodNum] = models.PeriodStatus{
					Status:   status,
					Position: cell.Content,
				}
			}
		}

		result[scorer.ScorerID] = player
	}

	return result, nil
}

// parseIntOrZero parses a string to int, returning 0 if empty or invalid
func parseIntOrZero(s string) int {
	if s == "" {
		return 0
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}
