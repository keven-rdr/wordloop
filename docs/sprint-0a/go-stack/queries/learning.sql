-- name: ListDueCards :many
SELECT id, item_id, direction, state, stability, due_at
FROM card
WHERE user_id = $1 AND state <> 'new' AND NOT suspended AND due_at <= $2
ORDER BY due_at
LIMIT $3;

-- name: InsertReview :execrows
INSERT INTO review_log (id, user_id, card_id, exercise_type, source, shown_at, answered_at, tz, day_cutoff_hour,
                        local_date, is_correct, grade, state_before, state_after, scheduler, scheduler_version)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
ON CONFLICT (id) DO NOTHING;

-- name: GetItemByNaturalKey :one
SELECT id, kind, natural_key, cefr, status FROM item WHERE course_id = $1 AND natural_key = $2;

-- name: CountKnownLexemes :one
SELECT count(*) FROM user_lexeme_known WHERE user_id = $1;
