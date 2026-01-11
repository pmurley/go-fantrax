package models

// ServiceTimeResponse represents the full API response for getTeamServiceTime
type ServiceTimeResponse struct {
	Data      ServiceTimeMetadata `json:"data"`
	Roles     []string            `json:"roles"`
	Responses []struct {
		Data ServiceTimeData `json:"data"`
	} `json:"responses"`
}

// ServiceTimeMetadata contains server metadata
type ServiceTimeMetadata struct {
	SDate int64  `json:"sDate"`
	Adrt  int    `json:"adrt"`
	Up    string `json:"up"`
}

// ServiceTimeData contains the main service time information
type ServiceTimeData struct {
	LatestPeriodAllowed int                      `json:"latestPeriodAllowed"`
	DisplayedSelections ServiceTimeSelections    `json:"displayedSelections"`
	DisplayedLists      ServiceTimeDisplayedList `json:"displayedLists"`
	ServiceTime         ServiceTime              `json:"serviceTime"`
}

// ServiceTimeSelections contains the selected team ID
type ServiceTimeSelections struct {
	TeamID string `json:"teamId"`
}

// ServiceTimeDisplayedList contains status definitions
type ServiceTimeDisplayedList struct {
	AllStatus []ServiceTimeStatus `json:"allStatus"`
}

// ServiceTimeStatus represents a roster status type
type ServiceTimeStatus struct {
	Code        string `json:"code"`
	SortOrder   int    `json:"sortOrder"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ID          string `json:"id"`
	ShortName   string `json:"shortName"`
}

// ServiceTime contains the main service time table data
type ServiceTime struct {
	Headers     []ServiceTimeHeader `json:"headers"`
	HelpText    string              `json:"helpText"`
	LeagueTitle string              `json:"leagueTitle"`
	Title       string              `json:"title"`
	Rows        []ServiceTimeRow    `json:"rows"`
}

// ServiceTimeHeader represents a column header
type ServiceTimeHeader struct {
	Name      string      `json:"name,omitempty"`
	Width     int         `json:"width,omitempty"`
	ID        string      `json:"id,omitempty"`
	ShortName interface{} `json:"shortName"` // Can be string or int (period number)
}

// ServiceTimeRow represents a player's service time row
type ServiceTimeRow struct {
	Cells  []ServiceTimeCell `json:"cells"`
	Scorer ServiceTimeScorer `json:"scorer"`
}

// ServiceTimeCell represents a cell in the service time table
type ServiceTimeCell struct {
	StatusID string `json:"statusId,omitempty"`
	Content  string `json:"content"`
}

// ServiceTimeScorer represents a player in the service time response
type ServiceTimeScorer struct {
	TeamName       string            `json:"teamName"`
	URLName        string            `json:"urlName"`
	ScorerID       string            `json:"scorerId"`
	PosShortNames  string            `json:"posShortNames"`
	Team           bool              `json:"team"`
	Icons          []ServiceTimeIcon `json:"icons,omitempty"`
	Rookie         bool              `json:"rookie"`
	MinorsEligible bool              `json:"minorsEligible"`
	PosIDs         []string          `json:"posIds"`
	TeamID         string            `json:"teamId"`
	Name           string            `json:"name"`
	TeamShortName  string            `json:"teamShortName"`
	ShortName      string            `json:"shortName"`
}

// ServiceTimeIcon represents an icon for a player
type ServiceTimeIcon struct {
	Tooltip string `json:"tooltip"`
	TypeID  string `json:"typeId"`
}

// --- Processed types for clean API ---

// TeamServiceTimeResult maps scorerId to player service time info
type TeamServiceTimeResult map[string]PlayerServiceTime

// PlayerServiceTime contains processed service time data for a player
type PlayerServiceTime struct {
	// Player info
	ScorerID         string
	Name             string
	ShortName        string
	TeamName         string
	TeamShortName    string
	Positions        string
	IsRookie         bool
	IsMinorsEligible bool

	// Totals
	DaysActive  int
	DaysReserve int
	DaysIR      int
	DaysMinors  int

	// Per-period history
	PeriodHistory map[int]PeriodStatus
}

// PeriodStatus represents a player's status for a specific period
type PeriodStatus struct {
	Status   RosterStatus
	Position string
}

// RosterStatus represents the roster status of a player
type RosterStatus string

const (
	StatusActive    RosterStatus = "ACTIVE"
	StatusReserve   RosterStatus = "RESERVE"
	StatusIR        RosterStatus = "IR"
	StatusMinors    RosterStatus = "MINORS"
	StatusNotOnTeam RosterStatus = "NOT_ON_TEAM"
)
