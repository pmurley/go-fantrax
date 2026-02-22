// upload_schedule reads a league schedule from a CSV (exported from Google Sheets)
// and uploads it to Fantrax period-by-period using SetPeriodMatchups.
//
// Usage:
//
//	FANTRAX_LEAGUE_ID=xxx go run ./examples/auth_client_only/upload_schedule/ [--dry-run] [--periods=1-142]
//
// The CSV is expected at schedule.csv in the repo root.
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pmurley/go-fantrax/auth_client"
	"github.com/pmurley/go-fantrax/models"
)

// spreadsheetNameOverrides maps team names that differ between the spreadsheet
// and Fantrax. Key is spreadsheet name, value is Fantrax name.
var spreadsheetNameOverrides = map[string]string{
	"Kansas City Monarchs": "Warwick Wombats",
	"Seattle Weiners":      "Seattle Wieners",
}

func main() {
	leagueID := os.Getenv("FANTRAX_LEAGUE_ID")
	if leagueID == "" {
		log.Fatal("Please set FANTRAX_LEAGUE_ID environment variable")
	}

	// Parse CLI flags
	dryRun := false
	periodStart, periodEnd := 1, 142
	for _, arg := range os.Args[1:] {
		if arg == "--dry-run" {
			dryRun = true
		} else if strings.HasPrefix(arg, "--periods=") {
			rangeStr := strings.TrimPrefix(arg, "--periods=")
			parts := strings.SplitN(rangeStr, "-", 2)
			var err error
			periodStart, err = strconv.Atoi(parts[0])
			if err != nil {
				log.Fatalf("Invalid period start: %v", err)
			}
			if len(parts) == 2 {
				periodEnd, err = strconv.Atoi(parts[1])
				if err != nil {
					log.Fatalf("Invalid period end: %v", err)
				}
			} else {
				periodEnd = periodStart
			}
		}
	}

	// ── Step 1: Parse the CSV schedule ──────────────────────────────────
	fmt.Println("=== Parsing schedule CSV ===")
	csvSchedule, periodColumns, err := parseScheduleCSV("schedule.csv")
	if err != nil {
		log.Fatalf("Failed to parse CSV: %v", err)
	}
	fmt.Printf("Parsed %d teams, %d periods from CSV\n", len(csvSchedule), len(periodColumns))

	// ── Step 2: Fetch current Fantrax setup ─────────────────────────────
	fmt.Println("\n=== Fetching Fantrax league setup ===")
	client, err := auth_client.NewClient(leagueID, false)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	fmt.Printf("Logged in as: %s\n", client.UserInfo.Username)

	setup, err := client.GetLeagueSetupMatchups()
	if err != nil {
		log.Fatalf("Failed to get league setup: %v", err)
	}
	fmt.Printf("Fantrax has %d teams, %d periods\n", len(setup.Teams), len(setup.Matchups))

	// ── Step 3: Build team name -> ID mapping ───────────────────────────
	nameToID := buildNameToIDMap(setup)
	fmt.Printf("Built name->ID map with %d entries\n", len(nameToID))

	// Validate all CSV team names can be resolved
	var unmapped []string
	for teamName := range csvSchedule {
		resolvedName := resolveTeamName(teamName)
		if _, ok := nameToID[resolvedName]; !ok {
			unmapped = append(unmapped, fmt.Sprintf("%q (resolved: %q)", teamName, resolvedName))
		}
	}
	if len(unmapped) > 0 {
		log.Fatalf("Cannot map these CSV team names to Fantrax IDs: %s", strings.Join(unmapped, ", "))
	}
	fmt.Println("All CSV team names resolved to Fantrax IDs")

	// Also validate all opponent names in cells
	for teamName, schedule := range csvSchedule {
		for period, cell := range schedule {
			if cell.OpponentName == "" {
				continue
			}
			resolvedOpp := resolveTeamName(cell.OpponentName)
			if _, ok := nameToID[resolvedOpp]; !ok {
				log.Fatalf("Cannot map opponent %q (in %s period %d) to Fantrax ID", cell.OpponentName, teamName, period)
			}
		}
	}
	fmt.Println("All opponent names resolved to Fantrax IDs")

	// ── Step 4: Build matchup pairs per period from CSV ─────────────────
	// Find the "Agents" team ID for the bye
	agentsID := ""
	for _, team := range setup.Teams {
		if team.Name == "Agents" {
			agentsID = team.TeamID
			break
		}
	}
	if agentsID == "" {
		log.Fatal("Could not find 'Agents' team in Fantrax setup")
	}

	newMatchups := buildMatchupsFromCSV(csvSchedule, periodColumns, nameToID, agentsID)
	fmt.Printf("Built matchups for %d periods\n", len(newMatchups))

	// ── Step 5: Show what would change for the requested period range ───
	fmt.Printf("\n=== Period range: %d to %d ===\n", periodStart, periodEnd)
	if dryRun {
		fmt.Println("DRY RUN — will not POST any changes")
	}

	changedPeriods := 0
	unchangedPeriods := 0
	for p := periodStart; p <= periodEnd; p++ {
		newPairs, ok := newMatchups[p]
		if !ok {
			fmt.Printf("Period %d: not in CSV, skipping\n", p)
			continue
		}

		currentPairs := setup.Matchups[p]
		if matchupsEqual(currentPairs, newPairs) {
			unchangedPeriods++
			continue
		}
		changedPeriods++

		if dryRun || p <= periodStart+1 {
			// Print detail for first two periods or dry run
			fmt.Printf("\nPeriod %d — CHANGED:\n", p)
			fmt.Printf("  Current:\n")
			printMatchups(setup, currentPairs)
			fmt.Printf("  New:\n")
			printMatchupsWithIDs(setup, newPairs)
		}
	}

	fmt.Printf("\nSummary: %d periods changed, %d unchanged\n", changedPeriods, unchangedPeriods)

	if dryRun {
		fmt.Println("\nDry run complete. Use without --dry-run to upload.")
		return
	}

	if changedPeriods == 0 {
		fmt.Println("Nothing to upload!")
		return
	}

	// ── Step 6: Upload period by period ─────────────────────────────────
	fmt.Printf("\n=== Uploading %d periods ===\n", changedPeriods)
	uploaded := 0
	for p := periodStart; p <= periodEnd; p++ {
		newPairs, ok := newMatchups[p]
		if !ok {
			continue
		}

		currentPairs := setup.Matchups[p]
		if matchupsEqual(currentPairs, newPairs) {
			continue
		}

		fmt.Printf("Uploading period %d...", p)
		err := client.SetPeriodMatchups(setup, p, newPairs)
		if err != nil {
			fmt.Printf(" FAILED: %v\n", err)
			log.Fatalf("Aborting after failure on period %d", p)
		}
		fmt.Printf(" OK\n")
		uploaded++

		// Rate limit: 1 second between requests
		if p < periodEnd {
			time.Sleep(1 * time.Second)
		}
	}
	fmt.Printf("\nUploaded %d periods successfully\n", uploaded)
}

// scheduleCell represents one cell in the CSV: an opponent and whether the
// row's team is home or away.
type scheduleCell struct {
	OpponentName string
	IsHome       bool // true if the row's team is Home
}

// parseScheduleCSV reads the schedule CSV and returns:
//   - A map of team name -> (period number -> scheduleCell)
//   - A slice of period numbers in column order
func parseScheduleCSV(path string) (map[string]map[int]scheduleCell, []int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open CSV: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("read CSV: %w", err)
	}
	if len(records) < 2 {
		return nil, nil, fmt.Errorf("CSV has fewer than 2 rows")
	}

	// Parse header to map column index -> period number
	header := records[0]
	colToPeriod := make(map[int]int) // column index -> period number
	var periodOrder []int
	for col := 3; col < len(header); col++ {
		cell := strings.TrimSpace(header[col])
		if cell == "" || cell == "All Star Break" {
			continue
		}
		period, err := strconv.Atoi(cell)
		if err != nil {
			continue // skip non-numeric headers
		}
		colToPeriod[col] = period
		periodOrder = append(periodOrder, period)
	}

	// Parse data rows (stop at first blank team name, which separates the two copies)
	result := make(map[string]map[int]scheduleCell)
	for row := 1; row < len(records); row++ {
		rec := records[row]
		if len(rec) < 4 {
			continue
		}
		teamName := strings.TrimSpace(rec[2])
		if teamName == "" || teamName == "team" {
			continue
		}
		// Skip duplicate rows (second copy of the schedule)
		if _, exists := result[teamName]; exists {
			continue
		}

		schedule := make(map[int]scheduleCell)
		for col, period := range colToPeriod {
			if col >= len(rec) {
				continue
			}
			cell := strings.TrimSpace(rec[col])
			if cell == "" {
				continue
			}
			opponent, isHome, err := parseCellValue(cell)
			if err != nil {
				return nil, nil, fmt.Errorf("row %d (team %s), col %d (period %d): %w", row, teamName, col, period, err)
			}
			schedule[period] = scheduleCell{
				OpponentName: opponent,
				IsHome:       isHome,
			}
		}
		result[teamName] = schedule
	}

	return result, periodOrder, nil
}

// parseCellValue parses "Opponent Name (H)" or "Opponent Name (A)" and returns
// the opponent name and whether the row's team is home.
func parseCellValue(cell string) (opponentName string, isHome bool, err error) {
	if strings.HasSuffix(cell, " (H)") {
		return strings.TrimSuffix(cell, " (H)"), true, nil
	}
	if strings.HasSuffix(cell, " (A)") {
		return strings.TrimSuffix(cell, " (A)"), false, nil
	}
	return "", false, fmt.Errorf("cell %q does not end with (H) or (A)", cell)
}

// resolveTeamName applies name overrides for spreadsheet -> Fantrax mapping.
func resolveTeamName(name string) string {
	if override, ok := spreadsheetNameOverrides[name]; ok {
		return override
	}
	return name
}

// buildNameToIDMap creates a map from Fantrax team name -> team ID.
func buildNameToIDMap(setup *models.LeagueSetupMatchups) map[string]string {
	m := make(map[string]string, len(setup.Teams))
	for _, team := range setup.Teams {
		m[team.Name] = team.TeamID
	}
	return m
}

// buildMatchupsFromCSV converts the parsed CSV schedule into MatchupPair slices
// per period, ready for SetPeriodMatchups.
func buildMatchupsFromCSV(
	csvSchedule map[string]map[int]scheduleCell,
	periods []int,
	nameToID map[string]string,
	agentsID string,
) map[int][]models.MatchupPair {
	result := make(map[int][]models.MatchupPair, len(periods))

	for _, period := range periods {
		seen := make(map[string]bool) // track teams already paired (by team ID)
		var pairs []models.MatchupPair

		for teamName, schedule := range csvSchedule {
			cell, ok := schedule[period]
			if !ok {
				continue
			}

			teamID := nameToID[resolveTeamName(teamName)]
			oppID := nameToID[resolveTeamName(cell.OpponentName)]

			// Skip if we already paired this matchup (from the other team's row)
			pairKey := teamID + "_" + oppID
			reversePairKey := oppID + "_" + teamID
			if seen[pairKey] || seen[reversePairKey] {
				continue
			}

			var pair models.MatchupPair
			if cell.IsHome {
				pair = models.MatchupPair{AwayTeamID: oppID, HomeTeamID: teamID}
			} else {
				pair = models.MatchupPair{AwayTeamID: teamID, HomeTeamID: oppID}
			}
			pairs = append(pairs, pair)
			seen[pairKey] = true
			seen[reversePairKey] = true
		}

		// Add the Agents bye matchup
		if !seen[agentsID] {
			pairs = append(pairs, models.MatchupPair{AwayTeamID: agentsID, HomeTeamID: "-1"})
		}

		result[period] = pairs
	}

	return result
}

func matchupsEqual(a, b []models.MatchupPair) bool {
	if len(a) != len(b) {
		return false
	}
	// Build sets for order-independent comparison
	setA := make(map[string]bool, len(a))
	for _, p := range a {
		setA[p.AwayTeamID+"_"+p.HomeTeamID] = true
	}
	for _, p := range b {
		if !setA[p.AwayTeamID+"_"+p.HomeTeamID] {
			return false
		}
	}
	return true
}

func printMatchups(setup *models.LeagueSetupMatchups, pairs []models.MatchupPair) {
	for _, pair := range pairs {
		away := teamShortName(setup, pair.AwayTeamID)
		home := teamShortName(setup, pair.HomeTeamID)
		if pair.HomeTeamID == "-1" {
			fmt.Printf("    %-6s BYE\n", away)
		} else {
			fmt.Printf("    %-6s vs %-6s\n", away, home)
		}
	}
}

func printMatchupsWithIDs(setup *models.LeagueSetupMatchups, pairs []models.MatchupPair) {
	for _, pair := range pairs {
		away := teamShortName(setup, pair.AwayTeamID)
		home := teamShortName(setup, pair.HomeTeamID)
		if pair.HomeTeamID == "-1" {
			fmt.Printf("    %-6s BYE\n", away)
		} else {
			fmt.Printf("    %-6s vs %-6s\n", away, home)
		}
	}
}

func teamShortName(setup *models.LeagueSetupMatchups, teamID string) string {
	if teamID == "-1" {
		return "BYE"
	}
	team := auth_client.GetTeamByID(setup, teamID)
	if team != nil {
		return team.ShortName
	}
	return teamID[:8] + "..."
}
