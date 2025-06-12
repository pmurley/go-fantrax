package auth_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pmurley/go-fantrax/internal/live_scoring"
	"io"
	"net/http"
)

func (c *Client) GetLiveScoring() (*live_scoring.LiveScoringResponse, error) {
	var requestPayload = FantraxRequest{
		Msgs: []FantraxMessage{
			{
				Method: "getLiveScoringStats",
				Data: map[string]interface{}{
					"leagueId": c.LeagueID,
					"sppId":    "-1",
					"teamId":   "ALL",
					"period":   24,
					"date":     "2025-04-19",
					"newView":  true,
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

	// Check for non-200 status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var response live_scoring.LiveScoringResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}
