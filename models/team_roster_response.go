package models

// TeamRosterResponse represents the full API response for getTeamRosterInfo
type TeamRosterResponse struct {
	Data      TeamRosterData `json:"data"`
	Roles     []string       `json:"roles"`
	Responses []struct {
		Data TeamRosterResponseData `json:"data"`
	} `json:"responses"`
}

// TeamRosterData contains server metadata
type TeamRosterData struct {
	SDate int64  `json:"sDate"`
	Adrt  int    `json:"adrt"`
	Up    string `json:"up"`
}

// TeamRosterResponseData contains the main roster information
type TeamRosterResponseData struct {
	Settings                TeamSettings           `json:"settings"`
	ScoringCategoryTypes    []CategoryType         `json:"scoringCategoryTypes"`
	TeamHeadingInfo         TeamHeadingInfo        `json:"teamHeadingInfo"`
	PeriodOpponentTeamIDs   []string               `json:"periodOppnentTeamIds"`
	Tabs                    []Tab                  `json:"tabs"`
	MiscData                MiscData               `json:"miscData"`
	Tables                  []RosterTable          `json:"tables"`
	FantasyTeams            []FantasyTeam          `json:"fantasyTeams"`
	MyTeamIDs               []string               `json:"myTeamIds"`
	AvailableActiveViewType string                 `json:"availableActiveViewType"`
	DisplayedLists          map[string]interface{} `json:"displayedLists"`
	DisplayedSelections     map[string]interface{} `json:"displayedSelections"`
	DataLists               map[string]interface{} `json:"dataLists"`
	LeagueNotices           []interface{}          `json:"leagueNotices"`
	RosterDisplayMap        []interface{}          `json:"rosterDisplayMap"`
	GoBackDays              []int                  `json:"goBackDays"`
	HideRowsLineupChange    bool                   `json:"hideRowsLineupChange"`
}

// TeamSettings contains league settings
type TeamSettings struct {
	LogoUploaded bool   `json:"logoUploaded"`
	LogoURL      string `json:"logoUrl"`
}

// CategoryType represents a scoring category type
type CategoryType struct {
	Value string `json:"value"`
	Key   string `json:"key"`
}

// TeamHeadingInfo contains team header information
type TeamHeadingInfo struct {
	H2HRecord struct {
		Name      string `json:"name"`
		ShortName string `json:"shortName"`
		Value     string `json:"value"`
	} `json:"h2hRecord"`
	Rank struct {
		Name      string `json:"name"`
		ShortName string `json:"shortName"`
		Value     string `json:"value"`
	} `json:"rank"`
	Owners struct {
		Owners    string `json:"owners"`
		ShortName string `json:"shortName"`
		Value     string `json:"value"`
	} `json:"owners"`
}

// Tab represents a view tab
type Tab struct {
	ViewType string `json:"viewType"`
	Text     string `json:"text"`
	Code     string `json:"code"`
}

// MiscData contains miscellaneous roster data
type MiscData struct {
	MaxActions int `json:"maxActions"`
	SalaryInfo struct {
		Title string `json:"title"`
		Info  []struct {
			TradeName string `json:"tradeName"`
			Display   string `json:"display"`
			Name      string `json:"name"`
			Tradeable bool   `json:"tradeable"`
			Value     string `json:"value"`
			Key       string `json:"key"`
		} `json:"info"`
	} `json:"salaryInfo"`
}

// RosterTable represents a table of players (active roster or reserves)
type RosterTable struct {
	Header              TableHeader   `json:"header"`
	Rows                []PlayerRow   `json:"rows"`
	StatusTotals        []StatusTotal `json:"statusTotals"`
	SCGroup             interface{}   `json:"scGroup"`
	SCGroupScorerHeader interface{}   `json:"scGroupScorerHeader"`
}

// TableHeader contains column definitions
type TableHeader struct {
	Cells []Column `json:"cells"`
}

// Column represents a table column
type Column struct {
	IsStat        bool    `json:"isStat"`
	SortDirection int     `json:"sortDirection"`
	SortKey       string  `json:"sortKey"`
	SCIPId        string  `json:"scipId"`
	SortType      string  `json:"sortType"`
	Name          string  `json:"name"`
	Width         float64 `json:"width"`
	ShortName     string  `json:"shortName"`
	Key           string  `json:"key"`
	MaxWidth      float64 `json:"maxWidth"`
}

// PlayerRow represents a row in the roster table
type PlayerRow struct {
	Scorer            Player   `json:"scorer"`
	EligibleStatusIDs []string `json:"eligibleStatusIds"`
	StatusID          string   `json:"statusId"`
	PosID             string   `json:"posId"`
	Cells             []Cell   `json:"cells"`
	TeamID            string   `json:"teamId,omitempty"`
	IsEmptyRosterSlot bool     `json:"isEmptyRosterSlot,omitempty"`
}

// Player represents a player's information
type Player struct {
	TeamName              string       `json:"teamName"`
	URLName               string       `json:"urlName"`
	HeadshotURL           string       `json:"headshotUrl"`
	ScorerID              string       `json:"scorerId"`
	UpcomingEventStatusID string       `json:"upcomingEventStatusId,omitempty"`
	PosIDsNoFlex          []string     `json:"posIdsNoFlex"`
	DefaultPosID          string       `json:"defaultPosId"`
	PosShortNames         string       `json:"posShortNames"`
	Team                  bool         `json:"team"`
	Icons                 []PlayerIcon `json:"icons"`
	PrimaryPosID          string       `json:"primaryPosId"`
	Rookie                bool         `json:"rookie"`
	MinorsEligible        bool         `json:"minorsEligible"`
	PosIDs                []string     `json:"posIds"`
	TeamID                string       `json:"teamId"`
	Name                  string       `json:"name"`
	TeamShortName         string       `json:"teamShortName"`
	ShortName             string       `json:"shortName"`
}

// PlayerIcon represents an icon shown for a player
type PlayerIcon struct {
	Tooltip string `json:"tooltip"`
	TypeID  string `json:"typeId"`
}

// Cell represents a data cell in the roster table
type Cell struct {
	Content string   `json:"content"`
	EventID string   `json:"eventId,omitempty"`
	PopOver *PopOver `json:"popOver,omitempty"`
}

// PopOver contains hover information for a cell
type PopOver struct {
	Scorer  Player `json:"scorer"`
	Header  string `json:"header"`
	Content string `json:"content"`
}

// StatusTotal represents roster status totals
type StatusTotal struct {
	StatusID string `json:"statusId"`
	Total    int    `json:"total"`
}

// FantasyTeam represents a team in the fantasy league
type FantasyTeam struct {
	LogoURL256   string `json:"logoUrl256"`
	Name         string `json:"name"`
	ID           string `json:"id"`
	LogoURL128   string `json:"logoUrl128"`
	ShortName    string `json:"shortName"`
	Commissioner bool   `json:"commissioner"`
	LogoID       string `json:"logoId"`
}
