package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/pmurley/go-fantrax/auth_client"
	"github.com/pmurley/go-fantrax/auth_client/parser"
)

func main() {
	// Get league ID from environment or use a default
	leagueID := os.Getenv("FANTRAX_LEAGUE_ID")
	if leagueID == "" {
		leagueID = "q8lydqf5m4u30rca" // Using the league ID from the example
	}

	// Create an auth client (with caching disabled for fresh data)
	client, err := auth_client.NewClient(leagueID, false)
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}

	// Fetch transaction history
	fmt.Println("Fetching transaction history...")
	rawResponse, err := client.GetTransactionDetailsHistory()
	if err != nil {
		log.Fatalf("Failed to fetch transaction history: %v", err)
	}

	// Pretty print the JSON for better readability
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, rawResponse, "", "  ")
	if err != nil {
		log.Fatalf("Failed to format JSON: %v", err)
	}

	// Save raw response to file
	outputFile := "transaction_history_response.json"
	err = os.WriteFile(outputFile, prettyJSON.Bytes(), 0644)
	if err != nil {
		log.Fatalf("Failed to write response to file: %v", err)
	}

	fmt.Printf("Transaction history response saved to %s\n", outputFile)
	fmt.Printf("Response size: %d bytes\n", len(rawResponse))

	// Parse the response
	fmt.Println("\nParsing transaction data...")
	historyResponse, err := parser.ParseTransactionHistoryResponse(rawResponse)
	if err != nil {
		log.Fatalf("Failed to parse response: %v", err)
	}

	userTimezone := ""
	if client.UserInfo != nil {
		userTimezone = client.UserInfo.Timezone
		fmt.Printf("User timezone: %s (%s)\n", client.UserInfo.TimezoneDisplay, userTimezone)
	}

	transactions, err := parser.ParseTransactions(historyResponse, userTimezone)
	if err != nil {
		log.Fatalf("Failed to parse transactions: %v", err)
	}

	// Display transaction summary
	fmt.Printf("\n=== Transaction Summary ===\n")
	fmt.Printf("Total transactions: %d\n", len(transactions))

	// Group by type
	byType := parser.GroupTransactionsByType(transactions)
	fmt.Printf("\nBy Type:\n")
	for txType, txs := range byType {
		fmt.Printf("  %s: %d\n", txType, len(txs))
		// If CLAIM, break down by FA vs WW
		if txType == "CLAIM" {
			faCount := 0
			wwCount := 0
			for _, tx := range txs {
				switch tx.ClaimType {
				case "FA":
					faCount++
				case "WW":
					wwCount++
				}
			}
			if faCount > 0 || wwCount > 0 {
				fmt.Printf("    - Free Agent (FA): %d\n", faCount)
				fmt.Printf("    - Waiver Wire (WW): %d\n", wwCount)
			}
		}
	}

	// Group by team
	byTeam := parser.GroupTransactionsByTeam(transactions)
	fmt.Printf("\nBy Team (%d teams):\n", len(byTeam))

	// Show teams with most transactions
	type teamCount struct {
		name  string
		count int
	}
	var teams []teamCount
	for team, txs := range byTeam {
		teams = append(teams, teamCount{team, len(txs)})
	}

	// Simple sort - show top 5
	for i := 0; i < len(teams)-1; i++ {
		for j := i + 1; j < len(teams); j++ {
			if teams[j].count > teams[i].count {
				teams[i], teams[j] = teams[j], teams[i]
			}
		}
	}

	for i := 0; i < 5 && i < len(teams); i++ {
		fmt.Printf("  %s: %d\n", teams[i].name, teams[i].count)
	}

	// Show recent transactions
	fmt.Printf("\n=== Recent Transactions (last 10) ===\n")
	start := len(transactions) - 10
	if start < 0 {
		start = 0
	}
	for i := start; i < len(transactions); i++ {
		tx := transactions[i]
		fmt.Printf("%d. %s", i+1, tx.Type)
		if tx.Type == "CLAIM" && tx.ClaimType != "" {
			fmt.Printf(" (%s)", tx.ClaimType)
		}
		fmt.Printf(" %s %s (%s) - %s",
			tx.PlayerName,
			tx.PlayerPosition,
			tx.PlayerTeam,
			tx.TeamName)
		if tx.BidAmount != "" {
			fmt.Printf(" - Bid: $%s", tx.BidAmount)
		}
		if tx.ExecutedBy != "" {
			fmt.Printf(" [%s]", tx.ExecutedBy)
		}
		fmt.Println()
	}

	// Save parsed transactions to file
	parsedOutputFile := "parsed_transactions.json"
	jsonData, err := json.MarshalIndent(transactions, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal transactions: %v", err)
	}

	err = os.WriteFile(parsedOutputFile, jsonData, 0644)
	if err != nil {
		log.Fatalf("Failed to write parsed output file: %v", err)
	}

	fmt.Printf("\nParsed transactions saved to: %s\n", parsedOutputFile)
}
