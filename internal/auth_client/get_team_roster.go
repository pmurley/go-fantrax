package auth_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pmurley/go-fantrax/models"
	"io"
	"net/http"

	"github.com/pmurley/go-fantrax/internal/parser"
)

// GetTeamRosterInfoRequest represents the request payload for getTeamRosterInfo
type GetTeamRosterInfoRequest struct {
	LeagueID string `json:"leagueId"`
	Reload   string `json:"reload"`
	Period   string `json:"period"`
	TeamID   string `json:"teamId,omitempty"`
}

// GetTeamRosterInfoRaw fetches the raw team roster response without parsing
func (c *Client) GetTeamRosterInfoRaw(period string, teamID string) (*models.TeamRosterResponse, error) {
	requestPayload := FantraxRequest{
		Msgs: []FantraxMessage{
			{
				Method: "getTeamRosterInfo",
				Data: GetTeamRosterInfoRequest{
					LeagueID: c.LeagueID,
					Reload:   "1",
					Period:   period,
					TeamID:   teamID,
				},
			},
		},
	}

	// Build refUrl with optional parameters
	refUrl := fmt.Sprintf("https://www.fantrax.com/fantasy/league/%s/team/roster;reload=1", c.LeagueID)
	if period != "" {
		refUrl += fmt.Sprintf(";period=%s", period)
	}
	if teamID != "" {
		refUrl += fmt.Sprintf(";teamId=%s", teamID)
	}

	// Add common Fantrax request fields
	fullRequest := map[string]interface{}{
		"msgs":   requestPayload.Msgs,
		"uiv":    3,
		"refUrl": refUrl,
		"dt":     0,
		"at":     0,
		"av":     "0.0",
		"tz":     "America/Chicago",
		"v":      "167.0.1",
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

	// Check for non-200 status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response models.TeamRosterResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetTeamRosterInfo fetches and parses the team roster into a simplified structure
func (c *Client) GetTeamRosterInfo(period string, teamID string) (*models.TeamRoster, error) {
	// Get the raw response
	rawResponse, err := c.GetTeamRosterInfoRaw(period, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw team roster info: %w", err)
	}

	// Marshal the response back to JSON for the parser
	jsonData, err := json.Marshal(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response for parsing: %w", err)
	}

	// Parse the response
	roster, err := parser.ParseTeamRosterResponse(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse team roster response: %w", err)
	}

	return roster, nil
}

// GetCurrentPeriodTeamRosterInfo fetches the team roster for the current period
func (c *Client) GetCurrentPeriodTeamRosterInfo(teamID string) (*models.TeamRoster, error) {
	// Empty string for period will get the current period
	return c.GetTeamRosterInfo("", teamID)
}

// GetCurrentPeriodTeamRosterInfoRaw fetches the raw team roster response for the current period
func (c *Client) GetCurrentPeriodTeamRosterInfoRaw(teamID string) (*models.TeamRosterResponse, error) {
	// Empty string for period will get the current period
	return c.GetTeamRosterInfoRaw("", teamID)
}

// GetMyTeamRosterInfo fetches the roster for the authenticated user's team
func (c *Client) GetMyTeamRosterInfo(period string) (*models.TeamRoster, error) {
	// Empty string for teamID will get the user's own team
	return c.GetTeamRosterInfo(period, "")
}

// GetMyTeamRosterInfoRaw fetches the raw roster response for the authenticated user's team
func (c *Client) GetMyTeamRosterInfoRaw(period string) (*models.TeamRosterResponse, error) {
	// Empty string for teamID will get the user's own team
	return c.GetTeamRosterInfoRaw(period, "")
}
