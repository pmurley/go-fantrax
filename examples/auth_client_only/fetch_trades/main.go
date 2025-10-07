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

	// Fetch trade history
	fmt.Println("Fetching trade history...")
	rawResponse, err := client.GetAllTradesRaw()
	if err != nil {
		log.Fatalf("Failed to fetch trade history: %v", err)
	}

	// Pretty print the JSON for better readability
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, rawResponse, "", "  ")
	if err != nil {
		log.Fatalf("Failed to format JSON: %v", err)
	}

	// Save raw response to file
	outputFile := "trade_history_response.json"
	err = os.WriteFile(outputFile, prettyJSON.Bytes(), 0644)
	if err != nil {
		log.Fatalf("Failed to write response to file: %v", err)
	}

	fmt.Printf("Trade history response saved to %s\n", outputFile)
	fmt.Printf("Response size: %d bytes\n", len(rawResponse))

	// Parse the trades
	fmt.Println("\nParsing trade data...")
	trades, err := client.GetAllTrades()
	if err != nil {
		log.Fatalf("Failed to parse trades: %v", err)
	}

	// Display trade summary
	fmt.Printf("\n=== Trade Summary ===\n")
	fmt.Printf("Total trade transactions: %d\n", len(trades))

	// Group trades by trade ID
	groupedTrades := parser.GroupTradesByTradeID(trades)
	fmt.Printf("Total number of trades: %d\n", len(groupedTrades))

	// Show recent trades
	fmt.Printf("\n=== Recent Trades ===\n")
	count := 0
	for tradeID, players := range groupedTrades {
		if count >= 5 {
			break
		}

		fmt.Printf("\nTrade %s (Date: %s, Period: %d):\n",
			tradeID,
			players[0].ProcessedDate.Format("Jan 2, 2006"),
			players[0].Period)

		// Group players by from/to teams
		for _, player := range players {
			fmt.Printf("  %s (%s, %s) from %s to %s\n",
				player.PlayerName,
				player.PlayerPosition,
				player.PlayerTeam,
				player.FromTeamName,
				player.ToTeamName)
		}
		count++
	}

	// Save parsed trades to file
	parsedOutputFile := "parsed_trades.json"
	jsonData, err := json.MarshalIndent(trades, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal trades: %v", err)
	}

	err = os.WriteFile(parsedOutputFile, jsonData, 0644)
	if err != nil {
		log.Fatalf("Failed to write parsed output file: %v", err)
	}

	fmt.Printf("\nParsed trades saved to: %s\n", parsedOutputFile)
}
