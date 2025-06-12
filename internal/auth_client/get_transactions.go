package auth_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pmurley/go-fantrax/internal/parser"
	"github.com/pmurley/go-fantrax/models"
)

// GetTransactionDetailsHistoryRequest represents the request payload for getTransactionDetailsHistory
type GetTransactionDetailsHistoryRequest struct {
	LeagueID          string `json:"leagueId"`
	MaxResultsPerPage string `json:"maxResultsPerPage"`
	ExecutedOnly      bool   `json:"executedOnly,omitempty"`
	IncludeDeleted    bool   `json:"includeDeleted,omitempty"`
	View              string `json:"view,omitempty"` // "CLAIM_DROP" or "TRADE"
	PageNumber        string `json:"pageNumber,omitempty"`
}

// GetTransactionDetailsHistoryRaw fetches the raw transaction history response without parsing
func (c *Client) GetTransactionDetailsHistoryRaw(maxResultsPerPage string) (json.RawMessage, error) {
	// Build the request payload matching the example
	fullRequest := map[string]interface{}{
		"msgs": []FantraxMessage{
			{
				Method: "getTransactionDetailsHistory",
				Data: GetTransactionDetailsHistoryRequest{
					LeagueID:          c.LeagueID,
					MaxResultsPerPage: maxResultsPerPage,
				},
			},
		},
		"uiv":    3,
		"refUrl": fmt.Sprintf("https://www.fantrax.com/fantasy/league/%s/transactions/history;maxResultsPerPage=%s", c.LeagueID, maxResultsPerPage),
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

	return json.RawMessage(body), nil
}

// GetTransactionDetailsHistory fetches transaction history with a default of 250 results per page
func (c *Client) GetTransactionDetailsHistory() (json.RawMessage, error) {
	return c.GetTransactionDetailsHistoryRaw("250")
}

// GetTransactionHistory fetches and parses the transaction history
func (c *Client) GetTransactionHistory(maxResultsPerPage string) ([]models.Transaction, error) {
	// Get raw response
	rawResponse, err := c.GetTransactionDetailsHistoryRaw(maxResultsPerPage)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw transaction history: %w", err)
	}

	// Parse the response
	historyResponse, err := parser.ParseTransactionHistoryResponse(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction history response: %w", err)
	}

	// Convert to simplified transactions
	transactions, err := parser.ParseTransactions(historyResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transactions: %w", err)
	}

	return transactions, nil
}

// GetAllTransactions fetches all claim/drop transactions across all pages
func (c *Client) GetAllTransactions() ([]models.Transaction, error) {
	var allTransactions []models.Transaction
	pageNumber := 1
	var expectedTotal int

	for {
		// Build request for this page
		req := GetTransactionDetailsHistoryRequest{
			LeagueID:          c.LeagueID,
			MaxResultsPerPage: "250",
			ExecutedOnly:      true,
			IncludeDeleted:    false,
			View:              "CLAIM_DROP",
			PageNumber:        fmt.Sprintf("%d", pageNumber),
		}

		// Get raw response
		rawResponse, err := c.GetTransactionDetailsHistoryFullRaw(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get transaction history page %d: %w", pageNumber, err)
		}

		// Parse the response
		historyResponse, err := parser.ParseTransactionHistoryResponse(rawResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to parse transaction history response page %d: %w", pageNumber, err)
		}

		// Convert to simplified transactions
		transactions, err := parser.ParseTransactions(historyResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to parse transactions page %d: %w", pageNumber, err)
		}

		// Get pagination info and set expected total on first page
		if len(historyResponse.Responses) > 0 {
			pagination := historyResponse.Responses[0].Data.PaginatedResultSet
			if pageNumber == 1 {
				expectedTotal = pagination.TotalNumResults
			}

			// Add transactions, but don't exceed expected total
			for _, tx := range transactions {
				if len(allTransactions) >= expectedTotal {
					break // Stop if we've reached the expected total
				}
				allTransactions = append(allTransactions, tx)
			}

			// Check if we have more pages or reached expected total
			if pageNumber >= pagination.TotalNumPages || len(allTransactions) >= expectedTotal {
				break
			}
		} else {
			break // No response data
		}

		pageNumber++
	}

	return allTransactions, nil
}

// GetTransactionDetailsHistoryFullRaw fetches the raw transaction history with all parameters
func (c *Client) GetTransactionDetailsHistoryFullRaw(req GetTransactionDetailsHistoryRequest) (json.RawMessage, error) {
	// Build refUrl with all parameters
	refUrl := fmt.Sprintf("https://www.fantrax.com/fantasy/league/%s/transactions/history;maxResultsPerPage=%s",
		c.LeagueID, req.MaxResultsPerPage)

	if req.View != "" {
		refUrl += fmt.Sprintf(";view=%s", req.View)
	}
	if req.PageNumber != "" {
		refUrl += fmt.Sprintf(";pageNumber=%s", req.PageNumber)
	}
	// Add executedOnly and includeDeleted to refUrl
	refUrl += fmt.Sprintf(";executedOnly=%t;includeDeleted=%t", req.ExecutedOnly, req.IncludeDeleted)

	// Build the request payload
	fullRequest := map[string]interface{}{
		"msgs": []FantraxMessage{
			{
				Method: "getTransactionDetailsHistory",
				Data:   req,
			},
		},
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

	req2, err := http.NewRequest("POST", "https://www.fantrax.com/fxpa/req?leagueId="+c.LeagueID, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.Do(req2)
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

	return json.RawMessage(body), nil
}

// GetTradesRaw fetches the raw trade history response
func (c *Client) GetTradesRaw(maxResultsPerPage string, pageNumber string, executedOnly bool) (json.RawMessage, error) {
	req := GetTransactionDetailsHistoryRequest{
		LeagueID:          c.LeagueID,
		MaxResultsPerPage: maxResultsPerPage,
		ExecutedOnly:      executedOnly,
		IncludeDeleted:    false,
		View:              "TRADE",
		PageNumber:        pageNumber,
	}
	return c.GetTransactionDetailsHistoryFullRaw(req)
}

// GetAllTradesRaw fetches all trades with default settings
func (c *Client) GetAllTradesRaw() (json.RawMessage, error) {
	return c.GetTradesRaw("250", "1", true)
}

// GetTrades fetches and parses trade history
func (c *Client) GetTrades(maxResultsPerPage string, pageNumber string, executedOnly bool) ([]models.Transaction, error) {
	// Get raw response
	rawResponse, err := c.GetTradesRaw(maxResultsPerPage, pageNumber, executedOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw trade history: %w", err)
	}

	// Parse the response
	historyResponse, err := parser.ParseTransactionHistoryResponse(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trade history response: %w", err)
	}

	// Convert to simplified transactions
	transactions, err := parser.ParseTransactions(historyResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trades: %w", err)
	}

	return transactions, nil
}

// GetAllTrades fetches all trade transactions across all pages
func (c *Client) GetAllTrades() ([]models.Transaction, error) {
	var allTrades []models.Transaction
	pageNumber := 1
	var expectedTotal int

	for {
		// Build request for this page
		req := GetTransactionDetailsHistoryRequest{
			LeagueID:          c.LeagueID,
			MaxResultsPerPage: "250",
			ExecutedOnly:      true,
			IncludeDeleted:    false,
			View:              "TRADE",
			PageNumber:        fmt.Sprintf("%d", pageNumber),
		}

		// Get raw response
		rawResponse, err := c.GetTransactionDetailsHistoryFullRaw(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get trade history page %d: %w", pageNumber, err)
		}

		// Parse the response
		historyResponse, err := parser.ParseTransactionHistoryResponse(rawResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to parse trade history response page %d: %w", pageNumber, err)
		}

		// Convert to simplified transactions
		transactions, err := parser.ParseTransactions(historyResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to parse trades page %d: %w", pageNumber, err)
		}

		// Get pagination info and set expected total on first page
		if len(historyResponse.Responses) > 0 {
			pagination := historyResponse.Responses[0].Data.PaginatedResultSet
			if pageNumber == 1 {
				expectedTotal = pagination.TotalNumResults
			}

			// Add transactions, but don't exceed expected total
			for _, tx := range transactions {
				if len(allTrades) >= expectedTotal {
					break // Stop if we've reached the expected total
				}
				allTrades = append(allTrades, tx)
			}

			// Check if we have more pages or reached expected total
			if pageNumber >= pagination.TotalNumPages || len(allTrades) >= expectedTotal {
				break
			}
		} else {
			break // No response data
		}

		pageNumber++
	}

	return allTrades, nil
}

// GetAllTransactionsIncludingTrades fetches both claims/drops and trades across all pages
func (c *Client) GetAllTransactionsIncludingTrades() ([]models.Transaction, error) {
	// Get claims and drops
	claimsDrops, err := c.GetAllTransactions()
	if err != nil {
		return nil, fmt.Errorf("failed to get claims/drops: %w", err)
	}

	// Get trades
	trades, err := c.GetAllTrades()
	if err != nil {
		return nil, fmt.Errorf("failed to get trades: %w", err)
	}

	// Combine all transactions
	allTransactions := append(claimsDrops, trades...)

	return allTransactions, nil
}

// GetTransactionsPaginated fetches transactions with pagination info
func (c *Client) GetTransactionsPaginated(view string, pageNumber int, maxResults int, executedOnly bool) ([]models.Transaction, *models.PaginatedResultSet, error) {
	req := GetTransactionDetailsHistoryRequest{
		LeagueID:          c.LeagueID,
		MaxResultsPerPage: fmt.Sprintf("%d", maxResults),
		ExecutedOnly:      executedOnly,
		IncludeDeleted:    false,
		View:              view,
		PageNumber:        fmt.Sprintf("%d", pageNumber),
	}

	// Get raw response
	rawResponse, err := c.GetTransactionDetailsHistoryFullRaw(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get transaction history page %d: %w", pageNumber, err)
	}

	// Parse the response
	historyResponse, err := parser.ParseTransactionHistoryResponse(rawResponse)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse transaction history response page %d: %w", pageNumber, err)
	}

	// Convert to simplified transactions
	transactions, err := parser.ParseTransactions(historyResponse)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse transactions page %d: %w", pageNumber, err)
	}

	// Get pagination info
	var pagination *models.PaginatedResultSet
	if len(historyResponse.Responses) > 0 {
		pagination = &historyResponse.Responses[0].Data.PaginatedResultSet
	}

	return transactions, pagination, nil
}
