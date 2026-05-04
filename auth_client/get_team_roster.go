package auth_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pmurley/go-fantrax/auth_client/parser"
	"github.com/pmurley/go-fantrax/models"
)

// GetTeamRosterInfoRequest represents the request payload for getTeamRosterInfo.
// ScoringCategoryType and StatsType are optional; when set they switch the
// response from the default "current daily lineup" to a per-period
// year-to-date stats view (with columns like GS, fpts, gp). The Fantrax API
// uses string codes for these — see WithScoringCategoryType / WithStatsType.
type GetTeamRosterInfoRequest struct {
	LeagueID            string `json:"leagueId"`
	Reload              string `json:"reload"`
	Period              string `json:"period"`
	TeamID              string `json:"teamId,omitempty"`
	ScoringCategoryType string `json:"scoringCategoryType,omitempty"`
	StatsType           string `json:"statsType,omitempty"`
}

// TeamRosterInfoOption configures an optional getTeamRosterInfo request param.
type TeamRosterInfoOption func(*teamRosterInfoOptions)

type teamRosterInfoOptions struct {
	scoringCategoryType string
	statsType           string
}

// WithScoringCategoryType sets the scoringCategoryType query param. The
// Fantrax API takes string codes (e.g. "1" for hitting/pitching scoring); the
// option is intentionally string-typed so callers can pass any value the API
// accepts without requiring a library update for new codes.
func WithScoringCategoryType(value string) TeamRosterInfoOption {
	return func(o *teamRosterInfoOptions) {
		o.scoringCategoryType = value
	}
}

// WithStatsType sets the statsType query param. Like WithScoringCategoryType,
// this is string-typed; common values include "2" for season YTD stats.
func WithStatsType(value string) TeamRosterInfoOption {
	return func(o *teamRosterInfoOptions) {
		o.statsType = value
	}
}

// GetTeamRosterInfoRaw fetches the raw team roster response without parsing.
// Pass WithScoringCategoryType / WithStatsType options to retrieve per-period
// YTD stats instead of the default daily lineup view.
func (c *Client) GetTeamRosterInfoRaw(period string, teamID string, opts ...TeamRosterInfoOption) (*models.TeamRosterResponse, error) {
	options := &teamRosterInfoOptions{}
	for _, opt := range opts {
		opt(options)
	}

	requestPayload := FantraxRequest{
		Msgs: []FantraxMessage{
			{
				Method: "getTeamRosterInfo",
				Data: GetTeamRosterInfoRequest{
					LeagueID:            c.LeagueID,
					Reload:              "1",
					Period:              period,
					TeamID:              teamID,
					ScoringCategoryType: options.scoringCategoryType,
					StatsType:           options.statsType,
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
		"tz":     "UTC",
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

// GetTeamRosterInfo fetches and parses the team roster into a simplified
// structure. Pass WithScoringCategoryType / WithStatsType options to retrieve
// per-period YTD stats instead of the default daily lineup view.
func (c *Client) GetTeamRosterInfo(period string, teamID string, opts ...TeamRosterInfoOption) (*models.TeamRoster, error) {
	// Get the raw response
	rawResponse, err := c.GetTeamRosterInfoRaw(period, teamID, opts...)
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

// GetCurrentPeriodTeamRosterInfo fetches the team roster for the current period.
// Accepts the same options as GetTeamRosterInfo.
func (c *Client) GetCurrentPeriodTeamRosterInfo(teamID string, opts ...TeamRosterInfoOption) (*models.TeamRoster, error) {
	// Empty string for period will get the current period
	return c.GetTeamRosterInfo("", teamID, opts...)
}

// GetCurrentPeriodTeamRosterInfoRaw fetches the raw team roster response for
// the current period. Accepts the same options as GetTeamRosterInfoRaw.
func (c *Client) GetCurrentPeriodTeamRosterInfoRaw(teamID string, opts ...TeamRosterInfoOption) (*models.TeamRosterResponse, error) {
	// Empty string for period will get the current period
	return c.GetTeamRosterInfoRaw("", teamID, opts...)
}

// GetMyTeamRosterInfo fetches the roster for the authenticated user's team.
// Accepts the same options as GetTeamRosterInfo.
func (c *Client) GetMyTeamRosterInfo(period string, opts ...TeamRosterInfoOption) (*models.TeamRoster, error) {
	// Empty string for teamID will get the user's own team
	return c.GetTeamRosterInfo(period, "", opts...)
}

// GetMyTeamRosterInfoRaw fetches the raw roster response for the authenticated
// user's team. Accepts the same options as GetTeamRosterInfoRaw.
func (c *Client) GetMyTeamRosterInfoRaw(period string, opts ...TeamRosterInfoOption) (*models.TeamRosterResponse, error) {
	// Empty string for teamID will get the user's own team
	return c.GetTeamRosterInfoRaw(period, "", opts...)
}
