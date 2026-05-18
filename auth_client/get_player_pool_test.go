package auth_client

import (
	"testing"

	"github.com/pmurley/go-fantrax/models"
)

// header8 is the real 8-column player-pool layout Fantrax returns today
// (no %Drafted / ADP columns), captured live from the API.
func header8() models.TableHeader {
	return models.TableHeader{Cells: []models.Column{
		{ShortName: "Rk", Key: "rankOv"},
		{ShortName: "Sta", SortType: "STATUS", Key: "status"},
		{ShortName: "Age", SortType: "AGE", Key: "age"},
		{ShortName: "Opp", Key: "opponent"},
		{ShortName: "FPts", SortType: "SCORE", Key: "fpts"},
		{ShortName: "FP/G", SortType: "FPTS_PER_GAME", Key: "fptsPerGame"},
		{ShortName: "Ros", SortType: "OVERVIEW_PERCENT_OWNED_2"},
		{ShortName: "+/-", SortType: "OVERVIEW_PLUS_MINUS_PERCENT_OWNED_2"},
	}}
}

// TestParseStatsTableEntry_EightColumnLayout reproduces the bug: the live
// player pool returns 8 columns, Age in cell[2]. The old code only parsed
// cells when len(cells) >= 10, so Age (and Status) were silently dropped.
func TestParseStatsTableEntry_EightColumnLayout(t *testing.T) {
	cols := buildColumnIndex(header8())
	entry := models.StatsTableEntry{
		Scorer: models.PoolScorer{
			ScorerID: "075zj",
			Name:     "Augusto Mendieta",
			Rank:     5914,
		},
		Cells: []models.StatsTableCell{
			{Content: "5914"},
			{Content: "FA", ToolTip: "Free Agent"},
			{Content: "21"},
			{Content: "BOS<br/>Mon 7:10PM"},
			{Content: "0"},
			{Content: "0"},
			{Content: "0%"},
			{Content: "0%"},
		},
	}

	player, err := parseStatsTableEntry(entry, cols)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if player.Age != 21 {
		t.Errorf("Age = %d, want 21", player.Age)
	}
	if player.FantasyStatus != "FA" {
		t.Errorf("FantasyStatus = %q, want \"FA\"", player.FantasyStatus)
	}
	if player.Rank != 5914 {
		t.Errorf("Rank = %d, want 5914", player.Rank)
	}
}

// TestParseStatsTableEntry_TenColumnLayout is a regression guard: when
// Fantrax includes %Drafted and ADP (10 columns), Age and %Rostered must
// still resolve correctly via the header rather than fixed positions.
func TestParseStatsTableEntry_TenColumnLayout(t *testing.T) {
	header := models.TableHeader{Cells: []models.Column{
		{ShortName: "Rk", Key: "rankOv"},
		{ShortName: "Sta", SortType: "STATUS", Key: "status"},
		{ShortName: "Age", SortType: "AGE", Key: "age"},
		{ShortName: "Opp", Key: "opponent"},
		{ShortName: "FPts", SortType: "SCORE", Key: "fpts"},
		{ShortName: "FP/G", SortType: "FPTS_PER_GAME", Key: "fptsPerGame"},
		{ShortName: "%D", SortType: "PERCENT_DRAFTED"},
		{ShortName: "ADP", SortType: "ADP"},
		{ShortName: "Ros", SortType: "OVERVIEW_PERCENT_OWNED_2"},
		{ShortName: "+/-", SortType: "OVERVIEW_PLUS_MINUS_PERCENT_OWNED_2"},
	}}
	cols := buildColumnIndex(header)
	entry := models.StatsTableEntry{
		Scorer: models.PoolScorer{ScorerID: "abc", Name: "Test Player", Rank: 12},
		Cells: []models.StatsTableCell{
			{Content: "12"}, {Content: "NYY", ToolTip: "New York Yankees"},
			{Content: "30"}, {Content: "@BOS"}, {Content: "100"},
			{Content: "5.5"}, {Content: "80%"}, {Content: "15"},
			{Content: "97%"}, {Content: "+2%"},
		},
	}

	player, err := parseStatsTableEntry(entry, cols)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if player.Age != 30 {
		t.Errorf("Age = %d, want 30", player.Age)
	}
	if player.PercentRostered != 97 {
		t.Errorf("PercentRostered = %v, want 97", player.PercentRostered)
	}
}
