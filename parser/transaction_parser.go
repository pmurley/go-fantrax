package parser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pmurley/go-fantrax/models"
)

// ParseTransactionHistoryResponse parses the raw transaction history response into structured data
func ParseTransactionHistoryResponse(data []byte) (*models.TransactionHistoryResponse, error) {
	var response models.TransactionHistoryResponse
	err := json.Unmarshal(data, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction history response: %w", err)
	}
	return &response, nil
}

// ParseTransactions converts the raw transaction response into a simplified list of transactions
func ParseTransactions(response *models.TransactionHistoryResponse, userTimezoneOffset string) ([]models.Transaction, error) {
	if len(response.Responses) == 0 {
		return nil, fmt.Errorf("no responses found in transaction history")
	}

	transactionData := response.Responses[0].Data
	rows := transactionData.Table.Rows

	transactions := make([]models.Transaction, 0, len(rows))

	// Keep track of transactions with shared cells (rowspan > 1)
	groupedTransactionData := make(map[string]*groupData)

	for _, row := range rows {
		tx, err := parseTransactionRow(row, userTimezoneOffset)
		if err != nil {
			// Log error but continue processing other transactions
			continue
		}

		// First, check if this is the first transaction in a group (has rowspan > 1)
		hasRowspanCells := false
		for _, cell := range row.Cells {
			if cell.Rowspan > 1 {
				hasRowspanCells = true
				break
			}
		}

		// If this row has cells with rowspan > 1, it's the start of a group
		if hasRowspanCells {
			gd := &groupData{
				teamName:   "",
				teamID:     "",
				date:       time.Time{},
				executedBy: "",
			}

			// Extract data from cells with rowspan
			for _, cell := range row.Cells {
				if cell.Rowspan > 1 {
					switch cell.Key {
					case "team":
						gd.teamName = cell.Content
						gd.teamID = cell.TeamID
					case "date":
						gd.date, gd.executedBy = parseDateCell(cell, userTimezoneOffset)
					}
				}
			}

			groupedTransactionData[row.TxSetID] = gd
		}

		// Check if this transaction belongs to a group we've seen
		if gd, exists := groupedTransactionData[row.TxSetID]; exists {
			// Only apply team data if not already set (DROP transactions in a group might not have team cells)
			if tx.TeamName == "" && tx.TeamID == "" {
				tx.TeamName = gd.teamName
				tx.TeamID = gd.teamID
			}
			// Apply date if not set
			if tx.ProcessedDate.IsZero() && !gd.date.IsZero() {
				tx.ProcessedDate = gd.date
				tx.ExecutedBy = gd.executedBy
			}
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// groupData holds shared data for grouped transactions
type groupData struct {
	teamName   string
	teamID     string
	date       time.Time
	executedBy string
}

// parseTransactionRow converts a single transaction row into a Transaction
func parseTransactionRow(row models.TransactionRow, userTimezoneOffset string) (models.Transaction, error) {
	tx := models.Transaction{
		ID:             row.TxSetID,
		Type:           row.TransactionCode,
		PlayerName:     row.Scorer.Name,
		PlayerID:       row.Scorer.ScorerID,
		PlayerTeam:     row.Scorer.TeamShortName,
		PlayerPosition: stripHTMLTags(row.Scorer.PosShortNames),
		Executed:       row.Executed,
	}

	// Check if this is a trade by looking for from/to cells
	hasFromTo := false

	// Parse cells for additional information
	for _, cell := range row.Cells {
		switch cell.Key {
		case "team":
			// For CLAIM/DROP transactions
			tx.TeamName = cell.Content
			tx.TeamID = cell.TeamID
		case "from":
			// For TRADE transactions
			tx.FromTeamName = cell.Content
			tx.FromTeamID = cell.TeamID
			hasFromTo = true
		case "to":
			// For TRADE transactions
			tx.ToTeamName = cell.Content
			tx.ToTeamID = cell.TeamID
			hasFromTo = true
		case "bid":
			tx.BidAmount = cell.Content
		case "priority":
			tx.Priority = cell.Content
		case "date":
			date, executedBy := parseDateCell(cell, userTimezoneOffset)
			tx.ProcessedDate = date
			tx.ExecutedBy = executedBy
		case "week":
			if period, err := strconv.Atoi(cell.Content); err == nil {
				tx.Period = period
			}
		}
	}

	// If we found from/to fields, this is a trade
	if hasFromTo {
		tx.Type = "TRADE"
		tx.TradeGroupID = row.TxSetID
		tx.TradeGroupSize = row.NumInGroup
	}

	return tx, nil
}

// parseDateCell extracts the date and execution information from a date cell
func parseDateCell(cell models.TableCell, userTimezoneOffset string) (time.Time, string) {
	var executedBy string
	dateStr := cell.Content

	// Check if executed by commissioner
	if cell.Icon == "COMMISSIONER" {
		executedBy = "COMMISSIONER"
	}

	// Parse the date string (format: "Wed Jun 11, 2025, 2:37PM")
	date, err := parseFantraxDateWithTimezone(dateStr, userTimezoneOffset)
	if err != nil {
		// Try to parse from tooltip if main content fails
		if cell.ToolTip != "" {
			// Extract date from tooltip (format: "<b>Processed</b> Wed Jun 11, 2025, 2:37:00 PM")
			re := regexp.MustCompile(`<b>Processed</b>\s+(.+?)<br/>`)
			if matches := re.FindStringSubmatch(cell.ToolTip); len(matches) > 1 {
				date, _ = parseFantraxDateWithTimezone(matches[1], userTimezoneOffset)
			}
		}
	}

	return date, executedBy
}

// parseFantraxDate parses Fantrax date format
func parseFantraxDate(dateStr string) (time.Time, error) {
	// Remove day name if present (e.g., "Tue Jun 10, 2025, 8:07AM" -> "Jun 10, 2025, 8:07AM")
	dateStr = strings.TrimSpace(dateStr)
	parts := strings.Split(dateStr, " ")
	if len(parts) >= 4 {
		// Check if first part is a day name (3 letters)
		if len(parts[0]) == 3 && strings.Contains("MonTueWedThuFriSatSun", parts[0]) {
			// Remove the day name
			dateStr = strings.Join(parts[1:], " ")
		}
	}

	// Try various formats
	formats := []string{
		"Jan 2, 2006, 3:04PM",
		"Jan 2, 2006, 3:04AM",
		"Jan 2, 2006, 3:04:05 PM",
		"January 2, 2006, 3:04PM",
		"January 2, 2006, 3:04:05 PM",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// parseFantraxDateWithTimezone parses Fantrax date format and converts from user timezone to UTC
func parseFantraxDateWithTimezone(dateStr string, userTimezoneOffset string) (time.Time, error) {
	// First parse the date in a neutral way
	localTime, err := parseFantraxDate(dateStr)
	if err != nil {
		return time.Time{}, err
	}

	// If no timezone offset provided, return as-is
	if userTimezoneOffset == "" {
		return localTime, nil
	}

	// Parse the timezone offset (e.g., "-0500")
	if len(userTimezoneOffset) != 5 || (userTimezoneOffset[0] != '+' && userTimezoneOffset[0] != '-') {
		return localTime, nil // Invalid format, return as-is
	}

	sign := 1
	if userTimezoneOffset[0] == '-' {
		sign = -1
	}

	// Extract hours and minutes
	hoursStr := userTimezoneOffset[1:3]
	minutesStr := userTimezoneOffset[3:5]

	hours, err := strconv.Atoi(hoursStr)
	if err != nil {
		return localTime, nil // Invalid format, return as-is
	}

	minutes, err := strconv.Atoi(minutesStr)
	if err != nil {
		return localTime, nil // Invalid format, return as-is
	}

	// Calculate total offset in seconds
	offsetSeconds := sign * (hours*3600 + minutes*60)

	// Create a fixed timezone with the user's offset
	userLocation := time.FixedZone("UserTimezone", offsetSeconds)

	// Interpret the parsed time as being in the user's timezone
	timeInUserTimezone := time.Date(
		localTime.Year(),
		localTime.Month(),
		localTime.Day(),
		localTime.Hour(),
		localTime.Minute(),
		localTime.Second(),
		localTime.Nanosecond(),
		userLocation,
	)

	// Convert to UTC
	return timeInUserTimezone.UTC(), nil
}

// stripHTMLTags removes HTML tags from a string
func stripHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	return re.ReplaceAllString(s, "")
}

// GroupTransactionsByType groups transactions by their type
func GroupTransactionsByType(transactions []models.Transaction) map[string][]models.Transaction {
	grouped := make(map[string][]models.Transaction)

	for _, tx := range transactions {
		grouped[tx.Type] = append(grouped[tx.Type], tx)
	}

	return grouped
}

// GroupTransactionsByTeam groups transactions by team
func GroupTransactionsByTeam(transactions []models.Transaction) map[string][]models.Transaction {
	grouped := make(map[string][]models.Transaction)

	for _, tx := range transactions {
		grouped[tx.TeamName] = append(grouped[tx.TeamName], tx)
	}

	return grouped
}

// GroupTransactionsByPeriod groups transactions by period
func GroupTransactionsByPeriod(transactions []models.Transaction) map[int][]models.Transaction {
	grouped := make(map[int][]models.Transaction)

	for _, tx := range transactions {
		grouped[tx.Period] = append(grouped[tx.Period], tx)
	}

	return grouped
}

// GroupTradesByTradeID groups trade transactions by their trade group ID
func GroupTradesByTradeID(transactions []models.Transaction) map[string][]models.Transaction {
	grouped := make(map[string][]models.Transaction)

	for _, tx := range transactions {
		if tx.Type == "TRADE" && tx.TradeGroupID != "" {
			grouped[tx.TradeGroupID] = append(grouped[tx.TradeGroupID], tx)
		}
	}

	return grouped
}
