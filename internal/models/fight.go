package models

import "time"

type Event struct {
	ID        int       `json:"id"`
	EventName string    `json:"event_name"`
	Date      time.Time `json:"date"`
	Location  string    `json:"location"`
	URL       string    `json:"url"`
}
type Fight struct {
	ID          int     `json:"fight_id"`
	EventID     int     `json:"event_id"`
	Fighter1ID  int     `json:"fighter1_id"`
	Fighter2ID  int     `json:"fighter2_id"`
	WinnerID    *int    `json:"winner_id"`
	WeightClass string  `json:"weight_class"`
	Method      string  `json:"method"`
	Round       int     `json:"round"`
	Time        string  `json:"time"`
	TimeFormat  string  `json:"time_format"`
	Referee     string  `json:"referee"`
	Details     *string `json:"details"`
	IsTitle     bool    `json:"is_title"`
	URL         string  `json:"url"`
}

type FightStats struct {
	FightID                             int `json:"fight_id"`
	FighterID                           int `json:"fighter_id"`
	RoundNumber                         int `json:"round_number"`
	Knockdowns                          int `json:"knockdowns"`
	SigStrikesLanded                    int `json:"sig_strikes_landed"`
	SigStrikesAttempted                 int `json:"sig_strikes_attempted"`
	TotalStrikesLanded                  int `json:"total_strikes_landed"`
	TotalStrikesAttempted               int `json:"total_strikes_attempted"`
	TotalTakedownsLanded                int `json:"total_takedowns_landed"`
	TotalTakedownsAttempted             int `json:"total_takedowns_attempted"`
	SubAttempts                         int `json:"sub_attempts"`
	Reversals                           int `json:"reversals"`
	ControlTime                         int `json:"control_time"`
	SignificantStrikesHeadLanded        int `json:"significant_strikes_head_landed"`
	SignificantStrikesHeadAttempted     int `json:"significant_strikes_head_attempted"`
	SignificantStrikesBodyLanded        int `json:"significant_strikes_body_landed"`
	SignificantStrikesBodyAttempted     int `json:"significant_strikes_body_attempted"`
	SignificantStrikesLegLanded         int `json:"significant_strikes_leg_landed"`
	SignificantStrikesLegAttempted      int `json:"significant_strikes_leg_attempted"`
	SignificantStrikesDistanceLanded    int `json:"significant_strikes_distance_landed"`
	SignificantStrikesDistanceAttempted int `json:"significant_strikes_distance_attempted"`
	SignificantStrikesClinchLanded      int `json:"significant_strikes_clinch_landed"`
	SignificantStrikesClinchAttempted   int `json:"significant_strikes_clinch_attempted"`
	SignificantStrikesGroundLanded      int `json:"significant_strikes_ground_landed"`
	SignificantStrikesGroundAttempted   int `json:"significant_strikes_ground_attempted"`
}
