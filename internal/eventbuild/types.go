package eventbuild

type Bundle struct {
	Metadata Metadata    `json:"metadata"`
	Roster   RosterFile  `json:"roster"`
	Events   EventsFile  `json:"events"`
	Players  PlayersFile `json:"players"`
	Teams    TeamsFile   `json:"teams"`
	Summary  SummaryFile `json:"summary"`

	Warnings []string `json:"-"`
}

type Metadata struct {
	EventName           string `json:"event_name,omitempty"`
	SourceADM           string `json:"source_adm"`
	WindowStart         string `json:"window_start"`
	WindowEnd           string `json:"window_end"`
	AssistWindowSeconds int    `json:"assist_window_seconds"`
	GeneratedAt         string `json:"generated_at"`
	PlayerCount         int    `json:"player_count"`
	TeamCount           int    `json:"team_count"`
}

type RosterFile struct {
	Players []RosterPlayer `json:"players"`
}

type RosterPlayer struct {
	PlayerID      string   `json:"player_id"`
	PreferredName string   `json:"preferred_name"`
	DisplayName   string   `json:"display_name"`
	TeamName      string   `json:"team_name,omitempty"`
	SeenNames     []string `json:"seen_names,omitempty"`
}

type EventsFile struct {
	Hits                []HitEvent                `json:"hits"`
	Kills               []KillEvent               `json:"kills"`
	EnvironmentalDeaths []EnvironmentalDeathEvent `json:"environmental_deaths"`
	Assists             []AssistEvent             `json:"assists"`
}

type HitEvent struct {
	Timestamp           string   `json:"timestamp"`
	LineNumber          int      `json:"line_number"`
	AttackerID          string   `json:"attacker_id"`
	AttackerDisplayName string   `json:"attacker_display_name"`
	VictimID            string   `json:"victim_id"`
	VictimDisplayName   string   `json:"victim_display_name"`
	BodyPart            string   `json:"body_part"`
	Damage              float64  `json:"damage"`
	AmmoType            string   `json:"ammo_type"`
	Weapon              string   `json:"weapon,omitempty"`
	RangeMeters         *float64 `json:"range_meters,omitempty"`
	Teamkill            bool     `json:"teamkill"`
}

type KillEvent struct {
	Timestamp         string   `json:"timestamp"`
	LineNumber        int      `json:"line_number"`
	KillerID          string   `json:"killer_id"`
	KillerDisplayName string   `json:"killer_display_name"`
	VictimID          string   `json:"victim_id"`
	VictimDisplayName string   `json:"victim_display_name"`
	Weapon            string   `json:"weapon,omitempty"`
	RangeMeters       *float64 `json:"range_meters,omitempty"`
	Teamkill          bool     `json:"teamkill"`
}

type EnvironmentalDeathEvent struct {
	Timestamp         string `json:"timestamp"`
	LineNumber        int    `json:"line_number"`
	VictimID          string `json:"victim_id"`
	VictimDisplayName string `json:"victim_display_name"`
	Cause             string `json:"cause"`
}

type AssistHitDetail struct {
	Timestamp   string   `json:"timestamp"`
	LineNumber  int      `json:"line_number"`
	Weapon      string   `json:"weapon,omitempty"`
	RangeMeters *float64 `json:"range_meters,omitempty"`
}

type AssistEvent struct {
	Timestamp           string            `json:"timestamp"`
	KillLineNumber      int               `json:"kill_line_number"`
	KillerID            string            `json:"killer_id"`
	KillerDisplayName   string            `json:"killer_display_name"`
	VictimID            string            `json:"victim_id"`
	VictimDisplayName   string            `json:"victim_display_name"`
	AssisterID          string            `json:"assister_id"`
	AssisterDisplayName string            `json:"assister_display_name"`
	Hits                []AssistHitDetail `json:"hits"`
}

type PlayersFile struct {
	Players []PlayerReport `json:"players"`
}

type PlayerReport struct {
	PlayerID            string                     `json:"player_id"`
	PreferredName       string                     `json:"preferred_name"`
	DisplayName         string                     `json:"display_name"`
	TeamName            string                     `json:"team_name,omitempty"`
	Stats               PlayerStats                `json:"stats"`
	Kills               []PlayerKillDetail         `json:"kills"`
	Deaths              []PlayerDeathDetail        `json:"deaths"`
	Assists             []PlayerAssistDetail       `json:"assists"`
	Teamkills           []PlayerKillDetail         `json:"teamkills"`
	EnvironmentalDeaths []PlayerEnvironmentalDeath `json:"environmental_deaths"`
	HitsDealt           HitBreakdown               `json:"hits_dealt"`
	HitsTaken           HitBreakdown               `json:"hits_taken"`
}

type PlayerStats struct {
	Kills               int      `json:"kills"`
	Deaths              int      `json:"deaths"`
	Assists             int      `json:"assists"`
	Teamkills           int      `json:"teamkills"`
	EnvironmentalDeaths int      `json:"environmental_deaths"`
	RatioValue          *float64 `json:"ratio_value,omitempty"`
	RatioDisplay        string   `json:"ratio_display"`
}

type PlayerKillDetail struct {
	Timestamp         string   `json:"timestamp"`
	LineNumber        int      `json:"line_number"`
	VictimID          string   `json:"victim_id"`
	VictimDisplayName string   `json:"victim_display_name"`
	VictimTeamName    string   `json:"victim_team_name,omitempty"`
	Weapon            string   `json:"weapon,omitempty"`
	RangeMeters       *float64 `json:"range_meters,omitempty"`
	Teamkill          bool     `json:"teamkill"`
}

type PlayerDeathDetail struct {
	Timestamp         string   `json:"timestamp"`
	LineNumber        int      `json:"line_number"`
	KillerID          string   `json:"killer_id"`
	KillerDisplayName string   `json:"killer_display_name"`
	KillerTeamName    string   `json:"killer_team_name,omitempty"`
	Weapon            string   `json:"weapon,omitempty"`
	RangeMeters       *float64 `json:"range_meters,omitempty"`
	Teamkill          bool     `json:"teamkill"`
	CountsTowardStats bool     `json:"counts_toward_stats"`
}

type PlayerAssistDetail struct {
	Timestamp         string            `json:"timestamp"`
	KillLineNumber    int               `json:"kill_line_number"`
	KillerID          string            `json:"killer_id"`
	KillerDisplayName string            `json:"killer_display_name"`
	VictimID          string            `json:"victim_id"`
	VictimDisplayName string            `json:"victim_display_name"`
	Hits              []AssistHitDetail `json:"hits"`
}

type PlayerEnvironmentalDeath struct {
	Timestamp  string `json:"timestamp"`
	LineNumber int    `json:"line_number"`
	Cause      string `json:"cause"`
}

type HitBreakdown struct {
	Total      int            `json:"total"`
	ByBodyPart map[string]int `json:"by_body_part"`
}

type TeamsFile struct {
	Teams []TeamReport `json:"teams"`
}

type TeamReport struct {
	Name      string               `json:"name"`
	Stats     TeamStats            `json:"stats"`
	Members   []TeamMemberSummary  `json:"members"`
	Matchups  []TeamMatchupSummary `json:"matchups"`
	Teamkills []PlayerKillDetail   `json:"teamkills"`
}

type TeamStats struct {
	Kills        int      `json:"kills"`
	Deaths       int      `json:"deaths"`
	Assists      int      `json:"assists"`
	Teamkills    int      `json:"teamkills"`
	RatioValue   *float64 `json:"ratio_value,omitempty"`
	RatioDisplay string   `json:"ratio_display"`
}

type TeamMemberSummary struct {
	PlayerID            string   `json:"player_id"`
	PreferredName       string   `json:"preferred_name"`
	DisplayName         string   `json:"display_name"`
	Kills               int      `json:"kills"`
	Deaths              int      `json:"deaths"`
	Assists             int      `json:"assists"`
	Teamkills           int      `json:"teamkills"`
	EnvironmentalDeaths int      `json:"environmental_deaths"`
	RatioValue          *float64 `json:"ratio_value,omitempty"`
	RatioDisplay        string   `json:"ratio_display"`
}

type TeamMatchupSummary struct {
	OpponentTeamName string   `json:"opponent_team_name"`
	Kills            int      `json:"kills"`
	Deaths           int      `json:"deaths"`
	RatioValue       *float64 `json:"ratio_value,omitempty"`
	RatioDisplay     string   `json:"ratio_display"`
}

type SummaryFile struct {
	PlayerTable  []SummaryPlayerRow `json:"player_table"`
	TeamTable    []SummaryTeamRow   `json:"team_table"`
	LongestHits  []ShotSummary      `json:"longest_hits"`
	LongestKills []ShotSummary      `json:"longest_kills"`
}

type SummaryPlayerRow struct {
	PlayerID            string   `json:"player_id"`
	PreferredName       string   `json:"preferred_name"`
	DisplayName         string   `json:"display_name"`
	TeamName            string   `json:"team_name,omitempty"`
	Kills               int      `json:"kills"`
	Deaths              int      `json:"deaths"`
	Assists             int      `json:"assists"`
	Teamkills           int      `json:"teamkills"`
	EnvironmentalDeaths int      `json:"environmental_deaths"`
	RatioValue          *float64 `json:"ratio_value,omitempty"`
	RatioDisplay        string   `json:"ratio_display"`
}

type SummaryTeamRow struct {
	Name         string   `json:"name"`
	Kills        int      `json:"kills"`
	Deaths       int      `json:"deaths"`
	Assists      int      `json:"assists"`
	Teamkills    int      `json:"teamkills"`
	RatioValue   *float64 `json:"ratio_value,omitempty"`
	RatioDisplay string   `json:"ratio_display"`
}

type ShotSummary struct {
	Timestamp           string  `json:"timestamp"`
	LineNumber          int     `json:"line_number"`
	AttackerID          string  `json:"attacker_id"`
	AttackerDisplayName string  `json:"attacker_display_name"`
	VictimID            string  `json:"victim_id"`
	VictimDisplayName   string  `json:"victim_display_name"`
	Weapon              string  `json:"weapon,omitempty"`
	RangeMeters         float64 `json:"range_meters"`
	BodyPart            string  `json:"body_part,omitempty"`
	IsKill              bool    `json:"is_kill"`
}
