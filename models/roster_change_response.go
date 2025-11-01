package models

// RosterChangeResponse represents the full API response from confirmOrExecuteTeamRosterChanges
type RosterChangeResponse struct {
	Data struct {
		SDate int64  `json:"sDate"`
		Adrt  int    `json:"adrt"`
		Up    string `json:"up"`
	} `json:"data"`
	Roles     []string `json:"roles"`
	Responses []struct {
		Data struct {
			FantasyResponse struct {
				MainMsg              string            `json:"mainMsg,omitempty"` // Error message if present
				MsgType              string            `json:"msgType"`
				LineupChanges        []interface{}     `json:"lineupChanges"`
				ShowConfirmWindow    bool              `json:"showConfirmWindow"`
				NavItems             []interface{}     `json:"navItems,omitempty"`
				ShowApplyToFuturePeriods bool          `json:"showApplyToFuturePeriods"`
				RemoveSubmitButton   bool              `json:"removeSubmitButton"`
				ApplyToFuturePeriods bool              `json:"applyToFuturePeriods"`
				ResourceMap          map[string]string `json:"resourceMap"`
			} `json:"fantasyResponse"`
			TextArray struct {
				Data  []interface{} `json:"data"`
				Model struct {
					RosterLimitPeriodDisplay        string `json:"rosterLimitPeriodDisplay"`
					RosterAdjustmentInfo            RosterAdjustmentInfo `json:"rosterAdjustmentInfo"`
					FirstIllegalRosterPeriodDisplay string `json:"firstIllegalRosterPeriodDisplay"`
					FirstIllegalRosterPeriod        int    `json:"firstIllegalRosterPeriod"`
					NumIllegalRosterMsgs            int    `json:"numIllegalRosterMsgs"`
					PlayerPickDeadlinePassed        bool   `json:"playerPickDeadlinePassed"`
					IllegalRosterMsgs               []string `json:"illegalRosterMsgs"`
					IllegalBefore                   bool   `json:"illegalBefore"`
					ChangeAllowed                   bool   `json:"changeAllowed"`
				} `json:"model"`
			} `json:"textArray"`
			Commissioner bool `json:"commissioner,omitempty"` // Present when adminMode was true
		} `json:"data"`
	} `json:"responses"`
}

// RosterAdjustmentInfo contains details about the roster changes and associated fees
type RosterAdjustmentInfo struct {
	LineupChanges        []string `json:"lineupChanges"`        // e.g., ["Active to Reserve", "Reserve to Active"]
	TotalFee             float64  `json:"totalFee"`
	TotalClaimFee        float64  `json:"totalClaimFee"`
	TotalLineupChangeFee float64  `json:"totalLineupChangeFee"`
	RosterLimitPeriod    int      `json:"rosterLimitPeriod"`
	TotalDropFee         float64  `json:"totalDropFee"`
}

// RosterChangeResult is a simplified representation of the roster change outcome
type RosterChangeResult struct {
	Success          bool     // True if the change was successful
	Changes          []string // List of changes made (e.g., "Active to Reserve")
	ErrorMessage     string   // Human-readable error message if failed
	Warnings         []string // Roster validation warnings (can exist even when successful)
	TotalFee         float64  // Total cost of the changes
	IsCommissioner   bool     // True if change was made in commissioner mode
}
