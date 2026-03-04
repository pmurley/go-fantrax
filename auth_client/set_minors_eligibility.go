package auth_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MinorsEligibilityRequest represents the request payload for setting minors eligibility
type MinorsEligibilityRequest struct {
	PlayerID                string `json:"playerId"`
	MinorsIneligibilityDate string `json:"minorsIneligibilityDate"` // "YYYY-MM-DD" to mark ineligible, "" to mark eligible
}

// MinorsEligibilityResponse represents the response from the minors eligibility endpoint
type MinorsEligibilityResponse struct {
	Msg string `json:"msg"`
}

// SetMinorsIneligible marks a player as minors-ineligible starting from today's date.
//
// Parameters:
//   - playerID: The player ID (scorerId) to mark as minors-ineligible
//
// Returns the API response or an error if the request failed.
func (c *Client) SetMinorsIneligible(playerID string) (*MinorsEligibilityResponse, error) {
	today := time.Now().Format("2006-01-02")
	return c.saveMinorsEligibility(playerID, today)
}

// SetMinorsEligible marks a player as minors-eligible (removes any ineligibility override).
//
// Parameters:
//   - playerID: The player ID (scorerId) to mark as minors-eligible
//
// Returns the API response or an error if the request failed.
func (c *Client) SetMinorsEligible(playerID string) (*MinorsEligibilityResponse, error) {
	return c.saveMinorsEligibility(playerID, "")
}

// saveMinorsEligibility is the internal function that calls the Fantrax minors eligibility override endpoint.
func (c *Client) saveMinorsEligibility(playerID string, ineligibilityDate string) (*MinorsEligibilityResponse, error) {
	requestPayload := MinorsEligibilityRequest{
		PlayerID:                playerID,
		MinorsIneligibilityDate: ineligibilityDate,
	}

	jsonStr, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal minors eligibility request: %w", err)
	}

	url := fmt.Sprintf("https://www.fantrax.com/fxa/saveMinorsEligibilityOverrideChanges?leagueId=%s", c.LeagueID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, fmt.Errorf("failed to create minors eligibility request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send minors eligibility request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("minors eligibility API returned non-200 status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read minors eligibility response body: %w", err)
	}

	var response MinorsEligibilityResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal minors eligibility response: %w", err)
	}

	return &response, nil
}
