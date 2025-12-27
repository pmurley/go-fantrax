package models

// PlayerPoolResponse represents the full API response for getPlayerStats
type PlayerPoolResponse struct {
	Data      PlayerPoolData `json:"data"`
	Roles     []string       `json:"roles"`
	Responses []struct {
		Data PlayerPoolResponseData `json:"data"`
	} `json:"responses"`
}

// PlayerPoolData contains server metadata
type PlayerPoolData struct {
	SDate int64  `json:"sDate"`
	Adrt  int    `json:"adrt"`
	Up    string `json:"up"`
}

// PlayerPoolResponseData contains the main player pool information
type PlayerPoolResponseData struct {
	DisplayedStatusOrTeam string             `json:"displayedStatusOrTeam"`
	PaginatedResultSet    PaginatedResultSet `json:"paginatedResultSet"`
	StatsTable            []StatsTableEntry  `json:"statsTable"`
	TableHeader           TableHeader        `json:"tableHeader"`
}

// Note: PaginatedResultSet is defined in transaction.go

// StatsTableEntry represents a single player entry in the stats table
type StatsTableEntry struct {
	Scorer         PoolScorer         `json:"scorer"`
	MultiPositions string             `json:"multiPositions,omitempty"`
	Cells          []StatsTableCell   `json:"cells"`
	Actions        []StatsTableAction `json:"actions"`
}

// PoolScorer represents the player information in the player pool
type PoolScorer struct {
	ScorerID       string       `json:"scorerId"`
	Name           string       `json:"name"`
	ShortName      string       `json:"shortName"`
	URLName        string       `json:"urlName"`
	TeamName       string       `json:"teamName"`
	TeamShortName  string       `json:"teamShortName"`
	TeamID         string       `json:"teamId"`
	HeadshotURL    string       `json:"headshotUrl,omitempty"`
	Rank           int          `json:"rank"`
	PosIDs         []string     `json:"posIds"`
	PosIDsNoFlex   []string     `json:"posIdsNoFlex"`
	PrimaryPosID   string       `json:"primaryPosId"`
	DefaultPosID   string       `json:"defaultPosId"`
	PosShortNames  string       `json:"posShortNames"`
	StatusID       string       `json:"statusId"`
	Rookie         bool         `json:"rookie"`
	MinorsEligible bool         `json:"minorsEligible"`
	Team           bool         `json:"team"`
	Icons          []PlayerIcon `json:"icons,omitempty"`
}

// StatsTableCell represents a cell in the stats table
type StatsTableCell struct {
	Content   string `json:"content"`
	ToolTip   string `json:"toolTip,omitempty"`
	TeamID    string `json:"teamId,omitempty"` // Fantasy team ID for rostered players
	GainColor int    `json:"gainColor,omitempty"`
}

// StatsTableAction represents an available action for a player
type StatsTableAction struct {
	TypeID string `json:"typeId"`
}

// PoolPlayer represents a fully parsed player from the player pool
type PoolPlayer struct {
	// Core identification
	PlayerID  string // Fantrax scorer ID
	Name      string // Full player name
	ShortName string // Abbreviated name (e.g., "S. Ohtani")
	URLName   string // URL-friendly name (e.g., "shohei-ohtani")

	// MLB team info
	MLBTeamName      string // Full team name (e.g., "Los Angeles Dodgers")
	MLBTeamShortName string // Abbreviation (e.g., "LAD")
	MLBTeamID        string // Team ID (e.g., "10280")

	// Player attributes
	Age            int  // Player age
	Rookie         bool // Is rookie
	MinorsEligible bool // Is minors eligible

	// Position info
	Positions       []string // All eligible position IDs
	PositionsNoFlex []string // Position IDs without flex positions
	PrimaryPosID    string   // Primary position ID
	DefaultPosID    string   // Default position ID
	PosShortNames   string   // HTML formatted positions (e.g., "<b>UT</b>,SP,UT2")
	MultiPositions  string   // Comma-separated positions (e.g., "UT,SP,UT3,UT4")

	// Fantasy status
	FantasyStatus   string // "FA", "W", or fantasy team abbreviation
	FantasyTeamID   string // Fantasy team ID if rostered, empty if FA/waivers
	FantasyTeamName string // Fantasy team name if rostered

	// Rankings and stats
	Rank              int     // Overall fantasy points rank
	FantasyPoints     float64 // Total fantasy points
	FantasyPointsPerG float64 // Fantasy points per game
	PercentDrafted    float64 // % of leagues player was drafted in
	ADP               float64 // Average draft position
	PercentRostered   float64 // % of leagues rostering this player
	RosterChange      float64 // Change in roster % from previous week

	// Schedule
	NextOpponent string // Next opponent with date/time (may contain HTML)

	// Media
	HeadshotURL string // Player headshot image URL

	// Icons/badges
	Icons []PlayerIcon // News, injury, minors-eligible icons etc.

	// Available actions
	Actions []string // Action type IDs available for this player
}
