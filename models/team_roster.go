package models

// TeamRoster represents a simplified view of a team's roster
type TeamRoster struct {
	TeamInfo       TeamInfo
	ActiveRoster   []RosterPlayer // Status ID "1"
	ReserveRoster  []RosterPlayer // Status ID "2"
	InjuredReserve []RosterPlayer // Status ID "3"
	MinorsRoster   []RosterPlayer // Status ID "9"
	ClaimBudget    float64
	LeagueTeams    []FantasyTeam
}

// TeamInfo contains basic team information
type TeamInfo struct {
	TeamID    string
	OwnerName string
	Record    string
	Rank      string
	LogoURL   string
}

// RosterPlayer represents a player on the roster with essential information
type RosterPlayer struct {
	PlayerID        string
	Name            string
	ShortName       string
	Age             int
	TeamName        string
	TeamShortName   string
	TeamID          string
	Positions       []string
	PrimaryPosition string
	PosShortNames   string // HTML formatted position string (e.g., "<b>C</b>")
	HeadshotURL     string
	URLName         string
	Rookie          bool
	MinorsEligible  bool
	Status          string       // Active, Reserve, etc.
	RosterPosition  string       // The position they're rostered at
	Stats           *PlayerStats // Strongly-typed stats (batting or pitching)
	NextGame        *GameInfo
}

// GameInfo represents upcoming game information
type GameInfo struct {
	Opponent        string
	DateTime        string
	EventID         string
	ProbablePitcher *PitcherInfo
}

// PitcherInfo represents opposing pitcher information
type PitcherInfo struct {
	Name      string
	ShortName string
	Stats     map[string]string
}
