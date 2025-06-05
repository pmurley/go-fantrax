package models

// BattingStats represents Category 5 "Tracked" batting statistics
type BattingStats struct {
	FantasyPointsPerGame  *float64 `json:"fpg,omitempty"`  // FP/G
	AtBats                *int     `json:"ab,omitempty"`   // AB
	Hits                  *int     `json:"h,omitempty"`    // H
	Runs                  *int     `json:"r,omitempty"`    // R
	Doubles               *int     `json:"2b,omitempty"`   // 2B
	Triples               *int     `json:"3b,omitempty"`   // 3B
	HomeRuns              *int     `json:"hr,omitempty"`   // HR
	RBI                   *int     `json:"rbi,omitempty"`  // RBI
	Walks                 *int     `json:"bb,omitempty"`   // BB
	Strikeouts            *int     `json:"so,omitempty"`   // SO
	StolenBases           *int     `json:"sb,omitempty"`   // SB
	CaughtStealing        *int     `json:"cs,omitempty"`   // CS
	HitByPitch            *int     `json:"hbp,omitempty"`  // HBP
	GIDP                  *int     `json:"gidp,omitempty"` // GIDP
	Errors                *int     `json:"e,omitempty"`    // E
	CaughtStealingAgainst *int     `json:"csa,omitempty"`  // CSA
	DoublePlays           *int     `json:"dp,omitempty"`   // DP
	Assists               *int     `json:"a,omitempty"`    // A
	AssistsOutfield       *int     `json:"aof,omitempty"`  // AOF
	Putouts               *int     `json:"po,omitempty"`   // PO
	PutoutsOutfield       *int     `json:"poof,omitempty"` // POOF
	StolenBasesAgainst    *int     `json:"sba,omitempty"`  // SBA
	PassedBalls           *int     `json:"pb,omitempty"`   // PB
	GamesPlayed           *int     `json:"gp,omitempty"`   // GP
}

// PitchingStats represents Category 5 "Tracked" pitching statistics
type PitchingStats struct {
	FantasyPointsPerGame *float64 `json:"fpg,omitempty"` // FP/G
	InningsPitched       *float64 `json:"ip,omitempty"`  // IP
	QualityStarts        *int     `json:"qs,omitempty"`  // QS
	Saves                *int     `json:"sv,omitempty"`  // SV
	BlownSaves           *int     `json:"bs,omitempty"`  // BS
	Holds                *int     `json:"hld,omitempty"` // HLD
	CompleteGames        *int     `json:"cg,omitempty"`  // CG
	HitsAllowed          *int     `json:"h,omitempty"`   // H
	EarnedRuns           *int     `json:"er,omitempty"`  // ER
	WalksAllowed         *int     `json:"bb,omitempty"`  // BB
	Strikeouts           *int     `json:"k,omitempty"`   // K
	ERA                  *float64 `json:"era,omitempty"` // ERA
	Balks                *int     `json:"bk,omitempty"`  // BK
	WildPitches          *int     `json:"wp,omitempty"`  // WP
	HitBatsmen           *int     `json:"hb,omitempty"`  // HB
	Shutouts             *int     `json:"sho,omitempty"` // SHO
	Pickoffs             *int     `json:"pko,omitempty"` // PKO
	GamesPlayed          *int     `json:"gp,omitempty"`  // GP
}

// PlayerStats represents a player's statistics (either batting or pitching)
type PlayerStats struct {
	Batting  *BattingStats  `json:"batting,omitempty"`
	Pitching *PitchingStats `json:"pitching,omitempty"`
}

// StatCategory represents the type of stats being returned
type StatCategory string

const (
	StatCategoryTracked     StatCategory = "5" // Category 5: "Tracked"
	StatCategoryStandard    StatCategory = "1" // Category 1: "Standard"
	StatCategorySabermetric StatCategory = "3" // Category 3: "Sabermetric"
)
