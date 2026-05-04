// team_roster_stats demonstrates fetching the per-period YTD stats view of a
// team roster (instead of the default daily-lineup view) by passing
// WithScoringCategoryType and WithStatsType to GetTeamRosterInfoRaw.
//
// The Fantrax getTeamRosterInfo endpoint returns two different shapes
// depending on whether scoringCategoryType + statsType are set:
//
//	default:           current daily lineup with slot assignments
//	with stats opts:   per-period YTD stat columns (GS, fpts, gp, ...)
//
// Fetching the stats view is useful for tasks like enforcing weekly
// game-start caps (count YTD GS deltas across days) or backfilling daily
// fantasy-points totals from snapshots.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pmurley/go-fantrax/auth_client"
)

func main() {
	leagueID := os.Getenv("FANTRAX_LEAGUE_ID")
	if leagueID == "" {
		leagueID = "q8lydqf5m4u30rca"
	}
	teamID := os.Getenv("FANTRAX_TEAM_ID")
	if teamID == "" {
		teamID = "j12wv4h4m6iakb28"
	}
	period := os.Getenv("FANTRAX_PERIOD")
	if period == "" {
		period = "1"
	}

	client, err := auth_client.NewClient(leagueID, true)
	if err != nil {
		log.Fatalf("create auth client: %v", err)
	}

	fmt.Printf("=== Fetching YTD stats view for team %s, period %s ===\n", teamID, period)

	// "1" selects the standard scoring category; "2" selects the YTD season
	// stats view. The library is intentionally string-typed here so callers
	// can pass any value the upstream API accepts.
	resp, err := client.GetTeamRosterInfoRaw(period, teamID,
		auth_client.WithScoringCategoryType("1"),
		auth_client.WithStatsType("2"),
	)
	if err != nil {
		log.Fatalf("get team roster (stats view): %v", err)
	}

	if len(resp.Responses) == 0 {
		log.Fatal("empty response")
	}
	tables := resp.Responses[0].Data.Tables
	fmt.Printf("\n%d table(s) returned\n", len(tables))

	// Walk each table and print the GS / fpts columns when present so the
	// caller can see the YTD-stats columns the variant exposes.
	for ti, table := range tables {
		var gsIdx, fptsIdx = -1, -1
		for i, col := range table.Header.Cells {
			if col.ShortName == "GS" {
				gsIdx = i
			}
			if col.Key == "fpts" {
				fptsIdx = i
			}
		}
		fmt.Printf("\n--- Table %d (scGroup=%v) — GS col=%d, fpts col=%d ---\n",
			ti, table.SCGroup, gsIdx, fptsIdx)

		shown := 0
		for _, row := range table.Rows {
			if row.IsEmptyRosterSlot || row.Scorer.Name == "" {
				continue
			}
			gs, fpts := "-", "-"
			if gsIdx >= 0 && gsIdx < len(row.Cells) {
				gs = row.Cells[gsIdx].Content
			}
			if fptsIdx >= 0 && fptsIdx < len(row.Cells) {
				fpts = row.Cells[fptsIdx].Content
			}
			fmt.Printf("  %-25s GS=%-4s fpts=%s\n", row.Scorer.Name, gs, fpts)
			shown++
			if shown >= 10 {
				fmt.Printf("  ... and %d more\n", len(table.Rows)-shown)
				break
			}
		}
	}
}
