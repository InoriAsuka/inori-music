package historypg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"inori-music/services/api/internal/history"
)

// Repository implements history.Repository using a PostgreSQL connection pool.
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SavePlayEvent(ctx context.Context, e history.PlayEvent) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO play_events (id, user_id, track_id, played_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING`,
		e.ID, e.UserID, e.TrackID, e.PlayedAt.UTC(), e.CreatedAt.UTC(),
	)
	return err
}

func (r *Repository) GetPlayEventByID(ctx context.Context, id string) (history.PlayEvent, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, track_id, played_at, created_at
		FROM play_events
		WHERE id = $1`, id)
	var e history.PlayEvent
	if err := row.Scan(&e.ID, &e.UserID, &e.TrackID, &e.PlayedAt, &e.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return history.PlayEvent{}, history.ErrEventNotFound
		}
		return history.PlayEvent{}, err
	}
	return e, nil
}

func (r *Repository) UpdatePlayEventByID(ctx context.Context, id string, playedAt time.Time) (history.PlayEvent, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE play_events
		SET played_at = $2
		WHERE id = $1
		RETURNING id, user_id, track_id, played_at, created_at`,
		id, playedAt.UTC())
	var e history.PlayEvent
	if err := row.Scan(&e.ID, &e.UserID, &e.TrackID, &e.PlayedAt, &e.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return history.PlayEvent{}, history.ErrEventNotFound
		}
		return history.PlayEvent{}, err
	}
	return e, nil
}

func (r *Repository) DeletePlayEventByID(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM play_events WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return history.ErrEventNotFound
	}
	return nil
}

func (r *Repository) DeletePlayEventsByIDs(ctx context.Context, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	tag, err := r.pool.Exec(ctx, `DELETE FROM play_events WHERE id = ANY($1)`, ids)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (r *Repository) DeletePlayEventsByIDsForUser(ctx context.Context, userID string, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	tag, err := r.pool.Exec(ctx, `DELETE FROM play_events WHERE id = ANY($1) AND user_id = $2`, ids, userID)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (r *Repository) ListPlayEvents(ctx context.Context, f history.PlayEventFilter) ([]history.PlayEvent, int, error) {
	if f.UserID == "" {
		return nil, 0, fmt.Errorf("userID is required")
	}

	// Build WHERE: user_id required, then optional track_id, since, until.
	clauses := []string{"user_id = $3"}
	args := []any{f.Limit, f.Offset, f.UserID}
	if f.TrackID != "" {
		args = append(args, f.TrackID)
		clauses = append(clauses, fmt.Sprintf("track_id = $%d", len(args)))
	}
	if !f.Since.IsZero() {
		args = append(args, f.Since.UTC())
		clauses = append(clauses, fmt.Sprintf("played_at >= $%d", len(args)))
	}
	if !f.Until.IsZero() {
		args = append(args, f.Until.UTC())
		clauses = append(clauses, fmt.Sprintf("played_at < $%d", len(args)))
	}
	where := clauses[0]
	for _, c := range clauses[1:] {
		where += " AND " + c
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, track_id, played_at, created_at,
		       COUNT(*) OVER () AS total_count
		FROM play_events
		WHERE `+where+`
		ORDER BY `+eventOrder(f.Asc)+`
		LIMIT $1 OFFSET $2`, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []history.PlayEvent{}, 0, nil
		}
		return nil, 0, err
	}
	defer rows.Close()

	var events []history.PlayEvent
	total := 0
	for rows.Next() {
		var e history.PlayEvent
		if err := rows.Scan(&e.ID, &e.UserID, &e.TrackID, &e.PlayedAt, &e.CreatedAt, &total); err != nil {
			return nil, 0, err
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if events == nil {
		events = []history.PlayEvent{}
	}
	return events, total, nil
}

func (r *Repository) DeletePlayEventsByUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM play_events WHERE user_id = $1`, userID)
	return err
}

func (r *Repository) DeletePlayEventsByUserAdmin(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM play_events WHERE user_id = $1`, userID)
	return err
}

func (r *Repository) DeletePlayEventsByTrack(ctx context.Context, trackID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM play_events WHERE track_id = $1`, trackID)
	return err
}

func (r *Repository) DeletePlayEventsInWindow(ctx context.Context, f history.StatsFilter) error {
	where, args := statsWhere(f)
	_, err := r.pool.Exec(ctx, `DELETE FROM play_events`+where, args...)
	return err
}

func (r *Repository) ListPlayEventsByTrack(ctx context.Context, f history.AdminPlayEventFilter) ([]history.PlayEvent, int, error) {
	if f.TrackID == "" {
		return nil, 0, fmt.Errorf("trackID is required")
	}

	// Build WHERE: track_id required, then optional user_id, since, until.
	clauses := []string{"track_id = $3"}
	args := []any{f.Limit, f.Offset, f.TrackID}
	if f.UserID != "" {
		args = append(args, f.UserID)
		clauses = append(clauses, fmt.Sprintf("user_id = $%d", len(args)))
	}
	if !f.Since.IsZero() {
		args = append(args, f.Since.UTC())
		clauses = append(clauses, fmt.Sprintf("played_at >= $%d", len(args)))
	}
	if !f.Until.IsZero() {
		args = append(args, f.Until.UTC())
		clauses = append(clauses, fmt.Sprintf("played_at < $%d", len(args)))
	}
	where := clauses[0]
	for _, c := range clauses[1:] {
		where += " AND " + c
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, track_id, played_at, created_at,
		       COUNT(*) OVER () AS total_count
		FROM play_events
		WHERE `+where+`
		ORDER BY `+eventOrder(f.Asc)+`
		LIMIT $1 OFFSET $2`, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []history.PlayEvent{}, 0, nil
		}
		return nil, 0, err
	}
	defer rows.Close()

	var events []history.PlayEvent
	total := 0
	for rows.Next() {
		var e history.PlayEvent
		if err := rows.Scan(&e.ID, &e.UserID, &e.TrackID, &e.PlayedAt, &e.CreatedAt, &total); err != nil {
			return nil, 0, err
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if events == nil {
		events = []history.PlayEvent{}
	}
	return events, total, nil
}

func (r *Repository) HistoryStats(ctx context.Context, f history.StatsFilter) (history.HistoryStats, error) {
	where, args := statsWhere(f)
	row := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*)                    AS total_events,
			COUNT(DISTINCT user_id)     AS unique_users,
			COUNT(DISTINCT track_id)    AS unique_tracks
		FROM play_events`+where, args...)
	var s history.HistoryStats
	if err := row.Scan(&s.TotalEvents, &s.UniqueUsers, &s.UniqueTracks); err != nil {
		return history.HistoryStats{}, err
	}
	return s, nil
}

func (r *Repository) TopTracks(ctx context.Context, f history.StatsFilter, limit int) ([]history.TrackPlayCount, error) {
	if limit <= 0 {
		limit = 10
	}
	where, args := statsWhere(f)
	nextParam := fmt.Sprintf("$%d", len(args)+1)
	args = append(args, limit)
	rows, err := r.pool.Query(ctx, `
		SELECT track_id, COUNT(*) AS play_count
		FROM play_events`+where+`
		GROUP BY track_id
		ORDER BY play_count DESC, track_id ASC
		LIMIT `+nextParam, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []history.TrackPlayCount
	for rows.Next() {
		var item history.TrackPlayCount
		if err := rows.Scan(&item.TrackID, &item.PlayCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = []history.TrackPlayCount{}
	}
	return result, nil
}

func (r *Repository) TopUsers(ctx context.Context, f history.StatsFilter, limit int) ([]history.UserPlayCount, error) {
	if limit <= 0 {
		limit = 10
	}
	where, args := statsWhere(f)
	nextParam := fmt.Sprintf("$%d", len(args)+1)
	args = append(args, limit)
	rows, err := r.pool.Query(ctx, `
		SELECT user_id, COUNT(*) AS play_count
		FROM play_events`+where+`
		GROUP BY user_id
		ORDER BY play_count DESC, user_id ASC
		LIMIT `+nextParam, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []history.UserPlayCount
	for rows.Next() {
		var item history.UserPlayCount
		if err := rows.Scan(&item.UserID, &item.PlayCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = []history.UserPlayCount{}
	}
	return result, nil
}

func (r *Repository) ListAllPlayEvents(ctx context.Context, f history.GlobalPlayEventFilter) ([]history.PlayEvent, int, error) {
	// Build optional WHERE clauses and args starting after limit($1) and offset($2).
	var clauses []string
	var filterArgs []any

	if f.UserID != "" {
		filterArgs = append(filterArgs, f.UserID)
		clauses = append(clauses, fmt.Sprintf("user_id = $%d", len(filterArgs)+2))
	}
	if f.TrackID != "" {
		filterArgs = append(filterArgs, f.TrackID)
		clauses = append(clauses, fmt.Sprintf("track_id = $%d", len(filterArgs)+2))
	}
	if !f.Since.IsZero() {
		filterArgs = append(filterArgs, f.Since.UTC())
		clauses = append(clauses, fmt.Sprintf("played_at >= $%d", len(filterArgs)+2))
	}
	if !f.Until.IsZero() {
		filterArgs = append(filterArgs, f.Until.UTC())
		clauses = append(clauses, fmt.Sprintf("played_at < $%d", len(filterArgs)+2))
	}

	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + clauses[0]
		for _, c := range clauses[1:] {
			where += " AND " + c
		}
	}

	// $1 = limit, $2 = offset; filter args are $3...$N
	args := append([]any{f.Limit, f.Offset}, filterArgs...)

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, track_id, played_at, created_at,
		       COUNT(*) OVER () AS total_count
		FROM play_events`+where+`
		ORDER BY `+eventOrder(f.Asc)+`
		LIMIT $1 OFFSET $2`, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []history.PlayEvent{}, 0, nil
		}
		return nil, 0, err
	}
	defer rows.Close()

	var events []history.PlayEvent
	total := 0
	for rows.Next() {
		var e history.PlayEvent
		if err := rows.Scan(&e.ID, &e.UserID, &e.TrackID, &e.PlayedAt, &e.CreatedAt, &total); err != nil {
			return nil, 0, err
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if events == nil {
		events = []history.PlayEvent{}
	}
	return events, total, nil
}

// eventOrder returns the ORDER BY expression for played_at-sorted event queries.
// asc=true → oldest-first; asc=false (default) → newest-first.
func eventOrder(asc bool) string {
	if asc {
		return "played_at ASC, id ASC"
	}
	return "played_at DESC, id DESC"
}

// statsWhere builds the optional WHERE clause and args slice for StatsFilter.
func statsWhere(f history.StatsFilter) (string, []any) {
	var clauses []string
	var args []any
	if !f.Since.IsZero() {
		args = append(args, f.Since.UTC())
		clauses = append(clauses, fmt.Sprintf("played_at >= $%d", len(args)))
	}
	if !f.Until.IsZero() {
		args = append(args, f.Until.UTC())
		clauses = append(clauses, fmt.Sprintf("played_at < $%d", len(args)))
	}
	if len(clauses) == 0 {
		return "", args
	}
	where := " WHERE "
	for i, c := range clauses {
		if i > 0 {
			where += " AND "
		}
		where += c
	}
	return where, args
}

// userStatsWhere builds a mandatory user_id clause plus optional time bounds.
func userStatsWhere(f history.UserStatsFilter) (string, []any) {
	args := []any{f.UserID}
	clauses := []string{"user_id = $1"}
	if !f.Since.IsZero() {
		args = append(args, f.Since.UTC())
		clauses = append(clauses, fmt.Sprintf("played_at >= $%d", len(args)))
	}
	if !f.Until.IsZero() {
		args = append(args, f.Until.UTC())
		clauses = append(clauses, fmt.Sprintf("played_at < $%d", len(args)))
	}
	where := " WHERE " + clauses[0]
	for _, c := range clauses[1:] {
		where += " AND " + c
	}
	return where, args
}

func (r *Repository) UserTopTracks(ctx context.Context, f history.UserStatsFilter, limit int) ([]history.TrackPlayCount, error) {
	if limit <= 0 {
		limit = 10
	}
	where, args := userStatsWhere(f)
	nextParam := fmt.Sprintf("$%d", len(args)+1)
	args = append(args, limit)
	rows, err := r.pool.Query(ctx, `
		SELECT track_id, COUNT(*) AS play_count
		FROM play_events`+where+`
		GROUP BY track_id
		ORDER BY play_count DESC, track_id ASC
		LIMIT `+nextParam, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []history.TrackPlayCount
	for rows.Next() {
		var item history.TrackPlayCount
		if err := rows.Scan(&item.TrackID, &item.PlayCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = []history.TrackPlayCount{}
	}
	return result, nil
}

func (r *Repository) UserHistoryStats(ctx context.Context, f history.UserStatsFilter) (history.UserHistoryStats, error) {
	where, args := userStatsWhere(f)
	row := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*)                AS total_events,
			COUNT(DISTINCT track_id) AS unique_tracks
		FROM play_events`+where, args...)
	var s history.UserHistoryStats
	if err := row.Scan(&s.TotalEvents, &s.UniqueTracks); err != nil {
		return history.UserHistoryStats{}, err
	}
	return s, nil
}

func (r *Repository) UserTrackPlayStats(ctx context.Context, userID, trackID string) (history.UserTrackStats, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*)   AS total_plays,
			MIN(played_at) AS first_played_at,
			MAX(played_at) AS last_played_at
		FROM play_events
		WHERE user_id = $1 AND track_id = $2`, userID, trackID)
	stats := history.UserTrackStats{TrackID: trackID}
	var (
		totalPlays    int
		firstPlayedAt *time.Time
		lastPlayedAt  *time.Time
	)
	if err := row.Scan(&totalPlays, &firstPlayedAt, &lastPlayedAt); err != nil {
		return history.UserTrackStats{}, err
	}
	stats.TotalPlays = totalPlays
	if firstPlayedAt != nil {
		stats.FirstPlayedAt = firstPlayedAt.UTC()
	}
	if lastPlayedAt != nil {
		stats.LastPlayedAt = lastPlayedAt.UTC()
	}
	return stats, nil
}

// trackStatsWhere builds a mandatory track_id clause plus optional time bounds.
func trackStatsWhere(f history.TrackStatsFilter) (string, []any) {
	args := []any{f.TrackID}
	clauses := []string{"track_id = $1"}
	if !f.Since.IsZero() {
		args = append(args, f.Since.UTC())
		clauses = append(clauses, fmt.Sprintf("played_at >= $%d", len(args)))
	}
	if !f.Until.IsZero() {
		args = append(args, f.Until.UTC())
		clauses = append(clauses, fmt.Sprintf("played_at < $%d", len(args)))
	}
	where := " WHERE " + clauses[0]
	for _, c := range clauses[1:] {
		where += " AND " + c
	}
	return where, args
}

func (r *Repository) TrackHistoryStats(ctx context.Context, f history.TrackStatsFilter) (history.TrackHistoryStatsResult, error) {
	where, args := trackStatsWhere(f)
	row := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*)                AS total_events,
			COUNT(DISTINCT user_id) AS unique_listeners
		FROM play_events`+where, args...)
	var s history.TrackHistoryStatsResult
	if err := row.Scan(&s.TotalEvents, &s.UniqueListeners); err != nil {
		return history.TrackHistoryStatsResult{}, err
	}
	return s, nil
}

func (r *Repository) TrackTopListeners(ctx context.Context, f history.TrackStatsFilter, limit int) ([]history.UserPlayCount, error) {
	if limit <= 0 {
		limit = 10
	}
	where, args := trackStatsWhere(f)
	nextParam := fmt.Sprintf("$%d", len(args)+1)
	args = append(args, limit)
	rows, err := r.pool.Query(ctx, `
		SELECT user_id, COUNT(*) AS play_count
		FROM play_events`+where+`
		GROUP BY user_id
		ORDER BY play_count DESC, user_id ASC
		LIMIT `+nextParam, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []history.UserPlayCount
	for rows.Next() {
		var item history.UserPlayCount
		if err := rows.Scan(&item.UserID, &item.PlayCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = []history.UserPlayCount{}
	}
	return result, nil
}

// timelineWhere builds a WHERE clause for a TimelineFilter.
// Since and Until are expected to be non-zero (validated by the service layer).
func timelineWhere(f history.TimelineFilter) (string, []any) {
	args := []any{f.Since.UTC(), f.Until.UTC()}
	clauses := []string{"played_at >= $1", "played_at < $2"}
	if f.UserID != "" {
		args = append(args, f.UserID)
		clauses = append(clauses, fmt.Sprintf("user_id = $%d", len(args)))
	}
	if f.TrackID != "" {
		args = append(args, f.TrackID)
		clauses = append(clauses, fmt.Sprintf("track_id = $%d", len(args)))
	}
	where := " WHERE " + clauses[0]
	for _, c := range clauses[1:] {
		where += " AND " + c
	}
	return where, args
}

func (r *Repository) HistoryTimeline(ctx context.Context, f history.TimelineFilter) ([]history.TimelineBucket, error) {
	gran := string(f.Granularity)
	if gran == "" {
		gran = string(history.GranularityDay)
	}
	where, args := timelineWhere(f)
	// Pass the granularity as a plain string literal; the service already validated it.
	rows, err := r.pool.Query(ctx, `
		SELECT DATE_TRUNC('`+gran+`', played_at AT TIME ZONE 'UTC') AS bucket,
		       COUNT(*) AS event_count
		FROM play_events`+where+`
		GROUP BY bucket
		ORDER BY bucket ASC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []history.TimelineBucket
	for rows.Next() {
		var b history.TimelineBucket
		if err := rows.Scan(&b.BucketStart, &b.EventCount); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = []history.TimelineBucket{}
	}
	return result, nil
}
