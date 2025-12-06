package auth_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// CreateClaimDropRequest represents the request payload for commissioner add/drop operations
// This is used for the /fxa/createClaimDrop endpoint which is different from the roster editing endpoint
type CreateClaimDropRequest struct {
	RosterLimitPeriod          string  `json:"rosterLimitPeriod"`          // The roster period (e.g., "1")
	ClaimScorerID              *string `json:"claimScorerId"`              // Player ID being added (null for drop-only)
	DropScorerID               *string `json:"dropScorerId"`               // Player ID being dropped (null for add-only)
	ClaimRosterActionID        *string `json:"claimRosterActionId"`        // Unknown - appears to be null in examples
	FantasyTeamID              string  `json:"fantasyTeamId"`              // Team ID performing the transaction
	TxDateTime                 string  `json:"txDateTime"`                 // Transaction date/time (e.g., "2026-03-24 23:05:00")
	FreeAgentBidAmount         *int    `json:"freeAgentBidAmount"`         // Bid amount (0 for add, null for drop)
	ClaimPosID                 *string `json:"claimPosId"`                 // Position ID for added player (null for drop)
	ClaimStatusID              *string `json:"claimStatusId"`              // Status ID for added player (null for drop)
	Future                     bool    `json:"future"`                     // Apply to future periods? Appears to be true in examples
	Override                   bool    `json:"override"`                   // Unknown - appears to be false in examples
	AdminModeProcessClaimNow   bool    `json:"adminModeProcessClaimNow"`   // Process immediately in commissioner mode (true for commissioner)
	AdminModeDropToStatusID    string  `json:"adminModeDropToStatusId"`    // Status for dropped player (e.g., "4" = Free Agent?)
	DoConfirm                  bool    `json:"doConfirm"`                  // Unknown - appears to be false in examples
	FAClaimSystem              string  `json:"faClaimSystem"`              // Free agent claim system (e.g., "BIDDING")
}

// CreateClaimDropResponse represents the response from the add/drop endpoint
type CreateClaimDropResponse struct {
	Code            string   `json:"code"`            // "EXECUTED" on success, "ERROR" on failure
	GenericMessage  string   `json:"genericMessage"`  // Human-readable message
	DetailMessages  []string `json:"detailMessages"`  // Detailed messages (HTML formatted)
	OtherMessages   []string `json:"otherMessages"`   // Additional messages
	TransactionID   string   `json:"transactionId"`   // Unique transaction ID
	Confirm         bool     `json:"confirm"`         // Whether confirmation is needed
	TransactionSet  *TransactionSet  `json:"transactionSet,omitempty"`  // Full transaction details
	FantasyItemOnTeam *interface{} `json:"fantasyItemOnTeam,omitempty"` // Player details (complex structure)
	FantasyItem     *interface{} `json:"fantasyItem,omitempty"`     // Player details (complex structure)
	Properties      map[string]string `json:"properties,omitempty"`    // Additional properties
}

// TransactionSet contains details about the transaction
type TransactionSet struct {
	ID                       string                 `json:"id"`
	LeagueID                 string                 `json:"leagueId"`
	CreatorUserID            string                 `json:"creatorUserId"`
	CreatorTeamID            string                 `json:"creatorTeamId"`
	DateCreated              string                 `json:"dateCreated"`
	DateProcessed            string                 `json:"dateProcessed,omitempty"`
	TimeProcessed            int64                  `json:"timeProcessed,omitempty"`
	ResolutionDate           string                 `json:"resolutionDate,omitempty"`
	ApplyToFuturePeriods     bool                   `json:"applyToFuturePeriods"`
	AdminMode                bool                   `json:"adminMode"`
	ServerID                 string                 `json:"serverId"`
	Status                   *TransactionStatus     `json:"status,omitempty"`
	Type                     *TransactionType       `json:"type,omitempty"`
	ClaimType                *ClaimType             `json:"claimType,omitempty"`
	Transactions             []Transaction          `json:"transactions"`
	ClaimPriority            int                    `json:"claimPriority,omitempty"`
	ClaimGroupNumber         int                    `json:"claimGroupNumber,omitempty"`
	MaxClaimsToExecute       int                    `json:"maxClaimsToExecute,omitempty"`
	FantasyTeamIdsWhoAccepted []string              `json:"fantasyTeamIdsWhoAccepted"`
	FantasyTeamIdsToAccept   []string               `json:"fantasyTeamIdsToAccept"`
	FantasyTeamIdsWhoObjected []string              `json:"fantasyTeamIdsWhoObjected"`
}

// TransactionStatus represents the status of a transaction
type TransactionStatus struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	ResourceKey string `json:"resourceKey"`
	SortOrder   int    `json:"sortOrder"`
	Pending     bool   `json:"pending"`
}

// TransactionType represents the type of transaction
type TransactionType struct {
	ID                   string   `json:"id"`
	Code                 string   `json:"code"`
	NameResource         string   `json:"nameResource"`
	ShortNameResource    string   `json:"shortNameResource"`
	DescriptionResource  string   `json:"descriptionResource"`
	HistoryTypes         []string `json:"historyTypes"`
	SortOrder            int      `json:"sortOrder"`
}

// ClaimType represents the type of claim (free agent, waiver, etc.)
type ClaimType struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	ResourceKey string `json:"resourceKey"`
	SortOrder   int    `json:"sortOrder"`
}

// Transaction represents an individual transaction (add or drop)
type Transaction struct {
	ScorerID           string  `json:"scorerId,omitempty"`
	DestinationTeamID  string  `json:"destinationTeamId,omitempty"`
	OriginTeamID       string  `json:"originTeamId,omitempty"`
	FreeAgentBidAmount float64 `json:"freeAgentBidAmount,omitempty"`
}

// IsSuccess returns true if the transaction was executed successfully
func (r *CreateClaimDropResponse) IsSuccess() bool {
	return r.Code == "EXECUTED"
}

// IsError returns true if there was an error executing the transaction
func (r *CreateClaimDropResponse) IsError() bool {
	return r.Code == "ERROR"
}

// commissionerAddWithStatus is a helper function that adds a player to a team with a specific status
// without needing to know the current period or the player's eligible positions.
//
// This function automatically:
//   - Fetches the current period
//   - Determines an appropriate position for the player
//   - Adds the player with the specified status
//
// The function uses intelligent position selection:
//   1. First attempts to add as a hitter (Utility position accepts all position players)
//   2. If that fails due to position eligibility, tries as a pitcher (Pitcher position accepts all pitchers)
//   3. Returns an error if neither position works
//
// Returns the API response or an error if the request failed.
func (c *Client) commissionerAddWithStatus(
	teamID string,
	playerID string,
	statusID string,
) (*CreateClaimDropResponse, error) {
	// Get current period
	period, err := c.GetCurrentPeriod()
	if err != nil {
		return nil, fmt.Errorf("failed to get current period: %w", err)
	}

	// Try adding as a hitter first (Utility position accepts all position players)
	response, err := c.CommissionerAdd(period, teamID, playerID, PosUtil, statusID)
	if err != nil {
		return nil, fmt.Errorf("failed to add player: %w", err)
	}

	// If successful, return
	if response.IsSuccess() {
		return response, nil
	}

	// If failed with position eligibility error, try as pitcher
	// Position eligibility errors have the pattern: "is not eligible as"
	if response.IsError() && len(response.DetailMessages) > 0 {
		isEligibilityError := false
		for _, msg := range response.DetailMessages {
			if strings.Contains(msg, "is not eligible as") {
				isEligibilityError = true
				break
			}
		}

		if isEligibilityError {
			// Try as pitcher (Pitcher position accepts all pitchers)
			pitcherResponse, pitcherErr := c.CommissionerAdd(period, teamID, playerID, PosP, statusID)
			if pitcherErr != nil {
				// If the second attempt also has a network error, return that
				return nil, fmt.Errorf("failed to add player as pitcher: %w", pitcherErr)
			}

			// Return the pitcher attempt response (success or error)
			return pitcherResponse, nil
		}
	}

	// Return the original response if it wasn't an eligibility error
	return response, nil
}

// CommissionerAddToReserve is a convenience function that adds a player to reserve
// without needing to know the current period or the player's eligible positions.
//
// Parameters:
//   - teamID: The fantasy team ID to add the player to
//   - playerID: The player ID (scorerId) to add
//
// Returns the API response or an error if the request failed.
func (c *Client) CommissionerAddToReserve(
	teamID string,
	playerID string,
) (*CreateClaimDropResponse, error) {
	return c.commissionerAddWithStatus(teamID, playerID, StatusReserve)
}

// CommissionerAddToMinors is a convenience function that adds a player to minors
// without needing to know the current period or the player's eligible positions.
//
// Parameters:
//   - teamID: The fantasy team ID to add the player to
//   - playerID: The player ID (scorerId) to add
//
// Returns the API response or an error if the request failed.
func (c *Client) CommissionerAddToMinors(
	teamID string,
	playerID string,
) (*CreateClaimDropResponse, error) {
	return c.commissionerAddWithStatus(teamID, playerID, StatusMinors)
}

// CommissionerAdd adds a player to a team's roster (commissioner mode only)
//
// This function is for commissioners/administrators to add players to any team.
// It uses a different endpoint than regular roster editing.
//
// Parameters:
//   - period: The roster period (week number) as an integer
//   - teamID: The fantasy team ID to add the player to
//   - playerID: The player ID (scorerId) to add
//   - positionID: The position slot ID (e.g., PosC, PosSS, PosUtil)
//   - statusID: The status ID (e.g., StatusActive, StatusReserve)
//
// The transaction date/time is automatically set to the current time in the user's timezone.
// The function uses hard-coded defaults for experimental/unknown fields.
//
// Returns the raw API response or an error if the request failed.
func (c *Client) CommissionerAdd(
	period int,
	teamID string,
	playerID string,
	positionID string,
	statusID string,
) (*CreateClaimDropResponse, error) {

	// Auto-generate transaction date/time in user's timezone
	// Format: "2006-01-02 15:04:05" (MySQL datetime format)
	var txDateTime string
	if c.UserInfo != nil && c.UserInfo.Timezone != "" {
		loc, err := time.LoadLocation(c.UserInfo.Timezone)
		if err != nil {
			// Fallback to UTC if timezone is invalid
			loc = time.UTC
		}
		txDateTime = time.Now().In(loc).Format("2006-01-02 15:04:05")
	} else {
		// Fallback to UTC if no user info
		txDateTime = time.Now().UTC().Format("2006-01-02 15:04:05")
	}

	// Build minimal request with hard-coded defaults for unknown fields
	bidAmount := 0
	requestPayload := CreateClaimDropRequest{
		RosterLimitPeriod:          fmt.Sprintf("%d", period),
		ClaimScorerID:              &playerID,
		DropScorerID:               nil, // No drop in add-only operation
		ClaimRosterActionID:        nil, // Unknown field - null in examples
		FantasyTeamID:              teamID,
		TxDateTime:                 txDateTime,
		FreeAgentBidAmount:         &bidAmount, // 0 for commissioner adds (no bidding)
		ClaimPosID:                 &positionID,
		ClaimStatusID:              &statusID,
		Future:                     true,  // Apply to future periods
		Override:                   false, // Unknown - false in examples
		AdminModeProcessClaimNow:   true,  // Process immediately (commissioner mode)
		AdminModeDropToStatusID:    "4",   // Status for drops - likely "4" = Free Agent
		DoConfirm:                  false, // Skip confirmation dialog
		FAClaimSystem:              "BIDDING", // TODO: May need to determine this from league settings
	}

	jsonStr, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal add request: %w", err)
	}

	// Use different endpoint than roster editing
	url := fmt.Sprintf("https://www.fantrax.com/fxa/createClaimDrop?leagueId=%s", c.LeagueID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, fmt.Errorf("failed to create add request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send add request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("add API returned non-200 status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read add response body: %w", err)
	}

	var response CreateClaimDropResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal add response: %w", err)
	}

	return &response, nil
}

// CommissionerDropToFreeAgent is a convenience function that drops a player to free agency
// without needing to know the current period.
//
// This function automatically fetches the current period and drops the player
// directly to the free agent pool (immediately available for pickup).
//
// Parameters:
//   - teamID: The fantasy team ID to drop the player from
//   - playerID: The player ID (scorerId) to drop
//
// Returns the API response or an error if the request failed.
func (c *Client) CommissionerDropToFreeAgent(
	teamID string,
	playerID string,
) (*CreateClaimDropResponse, error) {
	period, err := c.GetCurrentPeriod()
	if err != nil {
		return nil, fmt.Errorf("failed to get current period: %w", err)
	}
	return c.CommissionerDrop(period, teamID, playerID, false)
}

// CommissionerDropToWaivers is a convenience function that drops a player to waivers
// without needing to know the current period.
//
// This function automatically fetches the current period and drops the player
// to the waiver wire (subject to waiver claims before becoming a free agent).
//
// Parameters:
//   - teamID: The fantasy team ID to drop the player from
//   - playerID: The player ID (scorerId) to drop
//
// Returns the API response or an error if the request failed.
func (c *Client) CommissionerDropToWaivers(
	teamID string,
	playerID string,
) (*CreateClaimDropResponse, error) {
	period, err := c.GetCurrentPeriod()
	if err != nil {
		return nil, fmt.Errorf("failed to get current period: %w", err)
	}
	return c.CommissionerDrop(period, teamID, playerID, true)
}

// CommissionerDrop drops a player from a team's roster (commissioner mode only)
//
// This function is for commissioners/administrators to drop players from any team.
// It uses a different endpoint than regular roster editing.
//
// Parameters:
//   - period: The roster period (week number) as an integer
//   - teamID: The fantasy team ID to drop the player from
//   - playerID: The player ID (scorerId) to drop
//   - toWaivers: If true, player goes to waivers; if false, player becomes a free agent immediately
//
// The transaction date/time is automatically set to the current time in the user's timezone.
// The function uses hard-coded defaults for experimental/unknown fields.
//
// Returns the raw API response or an error if the request failed.
func (c *Client) CommissionerDrop(
	period int,
	teamID string,
	playerID string,
	toWaivers bool,
) (*CreateClaimDropResponse, error) {

	// Auto-generate transaction date/time in user's timezone
	var txDateTime string
	if c.UserInfo != nil && c.UserInfo.Timezone != "" {
		loc, err := time.LoadLocation(c.UserInfo.Timezone)
		if err != nil {
			loc = time.UTC
		}
		txDateTime = time.Now().In(loc).Format("2006-01-02 15:04:05")
	} else {
		txDateTime = time.Now().UTC().Format("2006-01-02 15:04:05")
	}

	// Determine drop destination status ID
	dropStatusID := DropToFreeAgent
	if toWaivers {
		dropStatusID = DropToWaivers
	}

	// Build minimal request for drop operation
	requestPayload := CreateClaimDropRequest{
		RosterLimitPeriod:          fmt.Sprintf("%d", period),
		ClaimScorerID:              nil, // No claim in drop-only operation
		DropScorerID:               &playerID,
		ClaimRosterActionID:        nil,
		FantasyTeamID:              teamID,
		TxDateTime:                 txDateTime,
		FreeAgentBidAmount:         nil, // null for drops
		ClaimPosID:                 nil, // null for drops
		ClaimStatusID:              nil, // null for drops
		Future:                     true,
		Override:                   false,
		AdminModeProcessClaimNow:   true,
		AdminModeDropToStatusID:    dropStatusID,
		DoConfirm:                  false,
		FAClaimSystem:              "BIDDING",
	}

	jsonStr, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal drop request: %w", err)
	}

	url := fmt.Sprintf("https://www.fantrax.com/fxa/createClaimDrop?leagueId=%s", c.LeagueID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, fmt.Errorf("failed to create drop request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send drop request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("drop API returned non-200 status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read drop response body: %w", err)
	}

	var response CreateClaimDropResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal drop response: %w", err)
	}

	return &response, nil
}
