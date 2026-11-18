-- name: GetSettings :one
SELECT tz, day_cutoff_hour FROM user_settings WHERE user_id = @user_id;

-- name: StudyDays :many
SELECT local_date, reviews, correct, new_items, active_ms, COALESCE(known_lexemes, 0)::int AS known_lexemes
FROM study_day
WHERE user_id = @user_id AND local_date BETWEEN @from_date AND @to_date AND reviews > 0
ORDER BY local_date;

-- name: ForecastCounts :many
-- Cartoes ja iniciados por dia de vencimento (no fuso e corte da pessoa); atrasadas contam em hoje.
SELECT GREATEST(x.due_date - @today::date, 0)::int AS day_offset, count(*)::int AS due
FROM (
    SELECT ((due_at AT TIME ZONE @tz::text) - make_interval(hours => @cutoff_hour::int))::date AS due_date
    FROM card
    WHERE user_id = @user_id AND state <> 'new' AND NOT suspended
) x
WHERE x.due_date - @today::date <= @last_offset::int
GROUP BY 1
ORDER BY 1;

-- name: KnownWords :one
SELECT count(*)::int FROM user_lexeme_known WHERE user_id = @user_id;

-- name: MostMissed :many
-- Erros dos ultimos dias por palavra; erro de digitacao de 1 letra (typo) nao conta.
SELECT c.item_id, term.body AS term, COALESCE(gloss.body, gloss_default.body, '')::text AS gloss, count(*)::int AS misses
FROM review_log r
JOIN card c ON c.id = r.card_id
JOIN item i ON i.id = c.item_id
JOIN course co ON co.id = i.course_id
JOIN item_text term ON term.item_id = i.id AND term.role = 'term' AND term.position = 0 AND term.locale = co.studied_locale
LEFT JOIN item_text gloss ON gloss.item_id = i.id AND gloss.role = 'gloss' AND gloss.position = 0 AND gloss.locale = @locale::text
LEFT JOIN item_text gloss_default ON gloss_default.item_id = i.id AND gloss_default.role = 'gloss' AND gloss_default.position = 0 AND gloss_default.locale = co.help_locale
WHERE r.user_id = @user_id AND NOT r.is_correct AND NOT r.typo_flag AND r.local_date >= @since
GROUP BY c.item_id, term.body, gloss.body, gloss_default.body
ORDER BY count(*) DESC, max(r.answered_at) DESC
LIMIT @max_rows;
