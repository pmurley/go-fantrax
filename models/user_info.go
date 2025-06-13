package models

// UserInfo contains detailed user information including timezone data
type UserInfo struct {
	LName                       string      `json:"lName"`
	Country                     string      `json:"country"`
	Timezone                    string      `json:"timezone"` // Current timezone offset (e.g., "-0500")
	ChatEnabled                 bool        `json:"chatEnabled"`
	LocationDisplay             string      `json:"locationDisplay"`
	TimezoneNoDST               string      `json:"timezoneNoDST"` // Standard time offset (e.g., "-0600")
	FName                       string      `json:"fName"`
	LookAndFeel                 LookAndFeel `json:"lookAndFeel"`
	TimezoneDST                 string      `json:"timezoneDST"` // Daylight saving time offset (e.g., "-0500")
	Logo                        string      `json:"logo"`
	PushIds                     []string    `json:"pushIds"`
	State                       string      `json:"state"`
	Email                       string      `json:"email"`
	TimezoneCode                string      `json:"timezoneCode"` // Timezone name (e.g., "US/Central")
	S1                          string      `json:"s1"`
	S2                          string      `json:"s2"`
	S3                          string      `json:"s3"`
	VerificationStatus          string      `json:"verificationStatus"`
	ReceiveEmail                bool        `json:"receiveEmail"`
	VerificationRequired        bool        `json:"verificationRequired"`
	UserID                      string      `json:"userId"`
	KillOldUI                   bool        `json:"killOldUi"`
	UBA                         bool        `json:"uba"`
	TimezoneDisplay             string      `json:"timezoneDisplay"` // Display name (e.g., "CDT")
	ChatNotificationTypeDefault string      `json:"chatNotificationTypeDefault"`
	Username                    string      `json:"username"`
}

// LookAndFeel contains UI preferences
type LookAndFeel struct {
	UIScalingContent       float64 `json:"uiScalingContent"`
	UIScalingSpacing       float64 `json:"uiScalingSpacing"`
	HideHeaderImage        bool    `json:"hideHeaderImage"`
	OmitOldUI              bool    `json:"omitOldUI"`
	MultiPagePlayerProfile bool    `json:"multiPagePlayerProfile"`
	PlayerProfileHover     bool    `json:"playerProfileHover"`
}
