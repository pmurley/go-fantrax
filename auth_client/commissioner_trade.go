package auth_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TradeItem represents a single player movement in a trade
type TradeItem struct {
	PlayerID   string // The player ID (scorerId) being traded
	FromTeamID string // The team the player is coming from
	ToTeamID   string // The team the player is going to
}

// CreateTradeRequest represents the request payload for commissioner trade operations
type CreateTradeRequest struct {
	Transactions map[string]string `json:"transactions"` // Map of index -> "SC,playerID,fromTeam,toTeam,"
	TxDateTime   string            `json:"txDateTime"`   // Transaction date/time (format: "2006-01-02 15:04:05")
	Period       string            `json:"period"`       // Roster period as string
	AdminMode    bool              `json:"adminMode"`    // Commissioner mode (always true for this function)
	Future       bool              `json:"future"`       // Apply to future periods
	Override     bool              `json:"override"`     // Override roster limits (typically false)
	Msg          string            `json:"msg"`          // Optional trade message/notes
}

// CreateTradeResponse represents the response from the trade endpoint
type CreateTradeResponse struct {
	Code           string          `json:"code"`                      // "EXECUTED" on success, "ERROR" on failure
	GenericMessage string          `json:"genericMessage"`            // Human-readable message
	DetailMessages []string        `json:"detailMessages"`            // Detailed messages
	OtherMessages  []string        `json:"otherMessages"`             // Additional messages
	TransactionID  string          `json:"transactionId"`             // Unique transaction ID
	Confirm        bool            `json:"confirm"`                   // Whether confirmation is needed
	TransactionSet *TransactionSet `json:"transactionSet,omitempty"`  // Full transaction details
}

// IsSuccess returns true if the trade was executed successfully
func (r *CreateTradeResponse) IsSuccess() bool {
	return r.Code == "EXECUTED"
}

// IsError returns true if there was an error executing the trade
func (r *CreateTradeResponse) IsError() bool {
	return r.Code == "ERROR"
}

// CommissionerTrade executes a trade between teams (commissioner mode only)
//
// This function is for commissioners/administrators to execute trades between any teams.
// It can handle 2-team or multi-team trades with any number of players.
//
// Parameters:
//   - period: The roster period as an integer
//   - items: A slice of TradeItem structs, each representing one player movement
//   - message: Optional trade message/notes (can be empty string)
//
// The transaction date/time is automatically set to the current time in the user's timezone.
//
// Returns the API response or an error if the request failed.
func (c *Client) CommissionerTrade(
	period int,
	items []TradeItem,
	message string,
) (*CreateTradeResponse, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("at least one trade item is required")
	}

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

	// Build transactions map
	// Each entry format: "SC,playerID,fromTeamID,toTeamID,"
	// SC = Scorer Change, trailing comma is for an optional descriptor (not required)
	transactions := make(map[string]string)
	for i, item := range items {
		value := fmt.Sprintf("SC,%s,%s,%s,", item.PlayerID, item.FromTeamID, item.ToTeamID)
		transactions[fmt.Sprintf("%d", i)] = value
	}

	requestPayload := CreateTradeRequest{
		Transactions: transactions,
		TxDateTime:   txDateTime,
		Period:       fmt.Sprintf("%d", period),
		AdminMode:    true,
		Future:       true,
		Override:     false,
		Msg:          message,
	}

	jsonStr, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trade request: %w", err)
	}

	url := fmt.Sprintf("https://www.fantrax.com/fxa/createTrade?leagueId=%s", c.LeagueID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, fmt.Errorf("failed to create trade request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send trade request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read trade response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("trade API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response CreateTradeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal trade response: %w", err)
	}

	return &response, nil
}
