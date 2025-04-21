package live_scoring

type LiveScoringResponse struct {
	Data      ResponseData `json:"data"`
	Roles     []string     `json:"roles"`
	Responses []Response   `json:"responses"`
}

type ResponseData struct {
	SDate int64  `json:"sDate"`
	Adrt  int    `json:"adrt"`
	Up    string `json:"up"`
}

type Response struct {
	Data ResponseDetails `json:"data"`
}

type ResponseDetails struct {
	ClientID            string                 `json:"clientId"`
	Teams               []Team                 `json:"teams"`
	Live2               bool                   `json:"live2"`
	V                   int                    `json:"v"`
	DisplayedSelections map[string]interface{} `json:"displayedSelections"`
	MiscData            map[string]interface{} `json:"miscData"`
	DisplayedLists      map[string]interface{} `json:"displayedLists"`
	AllPlayerStats      bool                   `json:"allPlayerStats"`
	ResourceMap         map[string]Resource    `json:"resourceMap"`
	AllEventsFinished   bool                   `json:"allEventsFinished"`
	Live                bool                   `json:"live"`
	StatsPerTeam        StatsPerTeam           `json:"statsPerTeam"`
}

type Team struct {
	Description string `json:"description"`
	ID          string `json:"id"`
	ShortName   string `json:"shortName"`
	LogoIconURL string `json:"logoIconUrl"`
}

type Resource struct {
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
}

type StatsPerTeam struct {
	AllTeamsStats map[string]TeamStats `json:"allTeamsStats"`
	StatsMap      map[string]TeamStats `json:"statsMap"`
}

type TeamStats struct {
	ACTIVE TeamStatsActive `json:"ACTIVE"`
}

type TeamStatsActive struct {
	ProjectedTotalsMap2   map[string]interface{} `json:"projectedTotalsMap2"`
	StatsMap              map[string]StatObject  `json:"statsMap"`
	ProjectedTotalsMap    map[string]float64     `json:"projectedTotalsMap"`
	PlayerGameInfo        []float64              `json:"playerGameInfo"`
	GameStatusMap         map[string]string      `json:"gameStatusMap"`
	TotalFpts2            float64                `json:"totalFpts2"`
	StatsMap2             map[string]StatObject  `json:"statsMap2"`
	TotalFpts             float64                `json:"totalFpts"`
	RemainingEventPercent map[string]float64     `json:"remainingEventPercent"`
}

type StatObject struct {
	Object1 float64       `json:"object1"`
	Object2 []StatDetails `json:"object2"`
}

type StatDetails struct {
	ScipID string  `json:"scipId"`
	Sv     string  `json:"sv"`
	Av     float64 `json:"av"`
	Fpts   float64 `json:"fpts"`
}

// Division represents a division in the league
type Division struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Teams []string `json:"teams"` // Team IDs in this division
}

// Standings represents the overall league standings
type Standings struct {
	Teams     []Team     `json:"teams"`
	Divisions []Division `json:"divisions"`
}
