package storage

import (
	"context"
	"path"
	"ufc_stats_api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetFighterIDByURL(pool *pgxpool.Pool, url string) (int, error) {
	ctx := context.Background()
	var id int
	hash := path.Base(url)
	err := pool.QueryRow(ctx, "SELECT id FROM fighters WHERE url like '%/' || $1", hash).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
func GetEventIDByURL(pool *pgxpool.Pool, url string) (int, error) {
	ctx := context.Background()
	var id int
	hash := path.Base(url)
	err := pool.QueryRow(ctx, "SELECT id FROM events WHERE url like '%/' || $1", hash).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func InsertFight(pool *pgxpool.Pool, f *models.Fight) error {
	ctx := context.Background()
	err := pool.QueryRow(ctx, "INSERT INTO fights(event_id, fighter1_id, fighter2_id, winner_id, weight_class, method, round, time, time_format, referee,details, is_title, url) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) ON CONFLICT (url) DO UPDATE SET event_id = EXCLUDED.event_id, fighter1_id = EXCLUDED.fighter1_id, fighter2_id = EXCLUDED.fighter2_id, winner_id = EXCLUDED.winner_id, weight_class = EXCLUDED.weight_class, method = EXCLUDED.method, round = EXCLUDED.round, time = EXCLUDED.time, time_format = EXCLUDED.time_format, referee = EXCLUDED.referee, details = EXCLUDED.details, is_title = EXCLUDED.is_title RETURNING id", f.EventID, f.Fighter1ID, f.Fighter2ID, f.WinnerID, f.WeightClass, f.Method, f.Round, f.Time, f.TimeFormat, f.Referee, f.Details, f.IsTitle, f.URL).Scan(&f.ID)
	if err != nil {
		return err
	}
	return nil
}

func GetFightsByEventID(pool *pgxpool.Pool, eventID int) ([]models.Fight, error) {
	ctx := context.Background()
	rows, err := pool.Query(ctx, "SELECT id, event_id, fighter1_id, fighter2_id, winner_id, weight_class, method, round, time, time_format, referee,details, is_title, url FROM fights WHERE event_id = $1", eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fights := []models.Fight{}
	for rows.Next() {
		var f models.Fight
		err := rows.Scan(&f.ID, &f.EventID, &f.Fighter1ID, &f.Fighter2ID, &f.WinnerID, &f.WeightClass, &f.Method, &f.Round, &f.Time, &f.TimeFormat, &f.Referee, &f.Details, &f.IsTitle, &f.URL)
		if err != nil {
			return nil, err
		}
		fights = append(fights, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return fights, nil
}

func GetFightsByFighterID(pool *pgxpool.Pool, fighterID int) ([]models.Fight, error) {
	ctx := context.Background()
	rows, err := pool.Query(ctx, "SELECT id, event_id, fighter1_id, fighter2_id, winner_id, weight_class, method, round, time, time_format, referee,details, is_title, url FROM fights WHERE fighter1_id = $1 OR fighter2_id = $1", fighterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fights := []models.Fight{}
	for rows.Next() {
		var f models.Fight
		err := rows.Scan(&f.ID, &f.EventID, &f.Fighter1ID, &f.Fighter2ID, &f.WinnerID, &f.WeightClass, &f.Method, &f.Round, &f.Time, &f.TimeFormat, &f.Referee, &f.Details, &f.IsTitle, &f.URL)
		if err != nil {
			return nil, err
		}
		fights = append(fights, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return fights, nil
}
func GetAllFightURLs(pool *pgxpool.Pool) ([]models.Fight, error) {
	ctx := context.Background()
	rows, err := pool.Query(ctx, "SELECT id,url FROM fights")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fights := []models.Fight{}
	for rows.Next() {
		var f models.Fight
		err := rows.Scan(&f.ID, &f.URL)
		if err != nil {
			return nil, err
		}
		fights = append(fights, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return fights, nil

}

func GetFightStatsByFightID(pool *pgxpool.Pool, fightID int) ([]models.FightStats, error) {
	ctx := context.Background()
	rows, err := pool.Query(ctx, `SELECT
		fight_id, fighter_id, round_number, knockdowns,
		sig_strikes_landed, sig_strikes_attempted,
		total_strikes_landed, total_strikes_attempted,
		total_takedowns_landed, total_takedowns_attempted,
		sub_attempts, reversals, control_time,
		significant_strikes_head_landed, significant_strikes_head_attempted,
		significant_strikes_body_landed, significant_strikes_body_attempted,
		significant_strikes_leg_landed, significant_strikes_leg_attempted,
		significant_strikes_distance_landed, significant_strikes_distance_attempted,
		significant_strikes_clinch_landed, significant_strikes_clinch_attempted,
		significant_strikes_ground_landed, significant_strikes_ground_attempted
	FROM fight_stats WHERE fight_id = $1 ORDER BY round_number, fighter_id`, fightID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stats := []models.FightStats{}
	for rows.Next() {
		var s models.FightStats
		err := rows.Scan(&s.FightID, &s.FighterID, &s.RoundNumber, &s.Knockdowns,
			&s.SigStrikesLanded, &s.SigStrikesAttempted,
			&s.TotalStrikesLanded, &s.TotalStrikesAttempted,
			&s.TotalTakedownsLanded, &s.TotalTakedownsAttempted,
			&s.SubAttempts, &s.Reversals, &s.ControlTime,
			&s.SignificantStrikesHeadLanded, &s.SignificantStrikesHeadAttempted,
			&s.SignificantStrikesBodyLanded, &s.SignificantStrikesBodyAttempted,
			&s.SignificantStrikesLegLanded, &s.SignificantStrikesLegAttempted,
			&s.SignificantStrikesDistanceLanded, &s.SignificantStrikesDistanceAttempted,
			&s.SignificantStrikesClinchLanded, &s.SignificantStrikesClinchAttempted,
			&s.SignificantStrikesGroundLanded, &s.SignificantStrikesGroundAttempted)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

func InsertFightStats(pool *pgxpool.Pool, s *models.FightStats) error {
	ctx := context.Background()
	_, err := pool.Exec(ctx, `INSERT INTO fight_stats (
		fight_id, fighter_id, round_number, knockdowns,
		sig_strikes_landed, sig_strikes_attempted,
		total_strikes_landed, total_strikes_attempted,
		total_takedowns_landed, total_takedowns_attempted,
		sub_attempts, reversals, control_time,
		significant_strikes_head_landed, significant_strikes_head_attempted,
		significant_strikes_body_landed, significant_strikes_body_attempted,
		significant_strikes_leg_landed, significant_strikes_leg_attempted,
		significant_strikes_distance_landed, significant_strikes_distance_attempted,
		significant_strikes_clinch_landed, significant_strikes_clinch_attempted,
		significant_strikes_ground_landed, significant_strikes_ground_attempted
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25)
	ON CONFLICT (fight_id, fighter_id, round_number) DO UPDATE SET
		knockdowns = EXCLUDED.knockdowns,
		sig_strikes_landed = EXCLUDED.sig_strikes_landed,
		sig_strikes_attempted = EXCLUDED.sig_strikes_attempted,
		total_strikes_landed = EXCLUDED.total_strikes_landed,
		total_strikes_attempted = EXCLUDED.total_strikes_attempted,
		total_takedowns_landed = EXCLUDED.total_takedowns_landed,
		total_takedowns_attempted = EXCLUDED.total_takedowns_attempted,
		sub_attempts = EXCLUDED.sub_attempts,
		reversals = EXCLUDED.reversals,
		control_time = EXCLUDED.control_time,
		significant_strikes_head_landed = EXCLUDED.significant_strikes_head_landed,
		significant_strikes_head_attempted = EXCLUDED.significant_strikes_head_attempted,
		significant_strikes_body_landed = EXCLUDED.significant_strikes_body_landed,
		significant_strikes_body_attempted = EXCLUDED.significant_strikes_body_attempted,
		significant_strikes_leg_landed = EXCLUDED.significant_strikes_leg_landed,
		significant_strikes_leg_attempted = EXCLUDED.significant_strikes_leg_attempted,
		significant_strikes_distance_landed = EXCLUDED.significant_strikes_distance_landed,
		significant_strikes_distance_attempted = EXCLUDED.significant_strikes_distance_attempted,
		significant_strikes_clinch_landed = EXCLUDED.significant_strikes_clinch_landed,
		significant_strikes_clinch_attempted = EXCLUDED.significant_strikes_clinch_attempted,
		significant_strikes_ground_landed = EXCLUDED.significant_strikes_ground_landed,
		significant_strikes_ground_attempted = EXCLUDED.significant_strikes_ground_attempted`,
		s.FightID, s.FighterID, s.RoundNumber, s.Knockdowns,
		s.SigStrikesLanded, s.SigStrikesAttempted,
		s.TotalStrikesLanded, s.TotalStrikesAttempted,
		s.TotalTakedownsLanded, s.TotalTakedownsAttempted,
		s.SubAttempts, s.Reversals, s.ControlTime,
		s.SignificantStrikesHeadLanded, s.SignificantStrikesHeadAttempted,
		s.SignificantStrikesBodyLanded, s.SignificantStrikesBodyAttempted,
		s.SignificantStrikesLegLanded, s.SignificantStrikesLegAttempted,
		s.SignificantStrikesDistanceLanded, s.SignificantStrikesDistanceAttempted,
		s.SignificantStrikesClinchLanded, s.SignificantStrikesClinchAttempted,
		s.SignificantStrikesGroundLanded, s.SignificantStrikesGroundAttempted)
	if err != nil {
		return err
	}
	return nil
}
