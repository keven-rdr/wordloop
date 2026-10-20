-- name: CreateSettings :exec
INSERT INTO user_settings (user_id) VALUES (@user_id) ON CONFLICT (user_id) DO NOTHING;

-- name: GetSettings :one
SELECT tz, day_cutoff_hour, scheduler FROM user_settings WHERE user_id = @user_id;

-- name: SaveTZ :exec
UPDATE user_settings SET tz = @tz, updated_at = now() WHERE user_id = @user_id;

-- name: DueCards :many
SELECT id, item_id, direction, state, stability, due_at, last_review_at, reps, lapses, since_lapse, scheduler
FROM card
WHERE user_id = @user_id AND state <> 'new' AND NOT suspended AND due_at <= @now::timestamptz
ORDER BY due_at, id
LIMIT @row_limit::int;

-- Proximos itens para estudar: sem cartao receptivo, ou com um cartao que nasceu num plano e nunca foi respondido.
-- name: NewItems :many
SELECT i.id
FROM item i
JOIN course c ON c.id = i.course_id AND (c.status = 'published' OR @include_drafts::boolean)
JOIN item_text term ON term.item_id = i.id AND term.role = 'term' AND term.position = 0 AND term.locale = c.studied_locale
LEFT JOIN unit u ON u.id = i.unit_id
LEFT JOIN card cd ON cd.user_id = @user_id AND cd.item_id = i.id AND cd.direction = 'receptive'
WHERE (i.status = 'published' OR @include_drafts::boolean)
  AND i.kind IN ('word', 'expression', 'phrasal_verb', 'sentence')
  AND (cd.id IS NULL OR cd.state = 'new')
ORDER BY c.slug, u.position NULLS LAST, i.cefr NULLS LAST, lower(term.body), i.id
LIMIT @row_limit::int;

-- O produtivo nasce quando o receptivo chega a S minimo, ha pelo menos um dia (nunca na mesma sessao).
-- name: ProductiveItems :many
SELECT r.item_id
FROM card r
LEFT JOIN card p ON p.user_id = r.user_id AND p.item_id = r.item_id AND p.direction = 'productive'
WHERE r.user_id = @user_id AND r.direction = 'receptive' AND r.state IN ('review', 'relearning') AND NOT r.suspended
  AND r.stability >= @min_stability::float8 AND r.last_review_at <= @before::timestamptz
  AND (p.id IS NULL OR p.state = 'new')
ORDER BY r.stability DESC, r.item_id
LIMIT @row_limit::int;

-- name: CreateCard :one
INSERT INTO card (user_id, item_id, direction, introduced_at)
VALUES (@user_id, @item_id, @direction, @now::timestamptz)
ON CONFLICT (user_id, item_id, direction) DO UPDATE SET introduced_at = card.introduced_at
RETURNING id, item_id, direction, state, stability, due_at, last_review_at, reps, lapses, since_lapse, scheduler;

-- name: CardContents :many
SELECT i.id AS item_id, term.body AS term,
       COALESCE(gloss.body, gloss_default.body, '')::text AS gloss,
       COALESCE(ex.body, '')::text AS example,
       COALESCE(exg.body, '')::text AS example_gloss,
       COALESCE((SELECT m.storage_key FROM item_media im JOIN media m ON m.id = im.media_id
                  WHERE im.item_id = i.id AND im.role = 'context' AND (m.status = 'approved' OR @include_drafts::boolean)
                  ORDER BY im.position LIMIT 1), '')::text AS image_key
FROM item i
JOIN course c ON c.id = i.course_id AND (c.status = 'published' OR @include_drafts::boolean)
JOIN item_text term ON term.item_id = i.id AND term.role = 'term' AND term.position = 0 AND term.locale = c.studied_locale
LEFT JOIN item_text gloss ON gloss.item_id = i.id AND gloss.role = 'gloss' AND gloss.position = 0 AND gloss.locale = @locale::text
LEFT JOIN item_text gloss_default ON gloss_default.item_id = i.id AND gloss_default.role = 'gloss' AND gloss_default.position = 0 AND gloss_default.locale = c.help_locale
LEFT JOIN item_text ex ON ex.item_id = i.id AND ex.role = 'example' AND ex.position = 0 AND ex.locale = c.studied_locale
LEFT JOIN item_text exg ON exg.item_id = i.id AND exg.role = 'example_gloss' AND exg.position = 0 AND exg.locale = @locale::text
WHERE i.id = ANY(@ids::uuid[]) AND (i.status = 'published' OR @include_drafts::boolean);

-- Distratores: mesmo curso e tipo, mesma unidade primeiro, sem ligacao semantica com o item e sem a mesma traducao
-- (senao haveria duas respostas certas). A ordem e fixa por par, sem acaso.
-- name: Distractors :many
SELECT o.id AS item_id, COALESCE(og.body, ogd.body, '')::text AS gloss
FROM item me
JOIN course c ON c.id = me.course_id
LEFT JOIN item_text mg ON mg.item_id = me.id AND mg.role = 'gloss' AND mg.position = 0 AND mg.locale = @locale::text
LEFT JOIN item_text mgd ON mgd.item_id = me.id AND mgd.role = 'gloss' AND mgd.position = 0 AND mgd.locale = c.help_locale
JOIN item o ON o.course_id = me.course_id AND o.id <> me.id AND o.kind = me.kind
           AND (o.status = 'published' OR @include_drafts::boolean)
LEFT JOIN item_text og ON og.item_id = o.id AND og.role = 'gloss' AND og.position = 0 AND og.locale = @locale::text
LEFT JOIN item_text ogd ON ogd.item_id = o.id AND ogd.role = 'gloss' AND ogd.position = 0 AND ogd.locale = c.help_locale
WHERE me.id = @item_id
  AND COALESCE(og.body, ogd.body, '') <> ''
  AND lower(COALESCE(og.body, ogd.body, '')) <> lower(COALESCE(mg.body, mgd.body, ''))
  AND NOT EXISTS (SELECT 1 FROM item_relation r
                   WHERE (r.from_item_id = me.id AND r.to_item_id = o.id) OR (r.from_item_id = o.id AND r.to_item_id = me.id))
ORDER BY (o.unit_id IS NOT DISTINCT FROM me.unit_id) DESC, hashtext(o.id::text || me.id::text), o.id
LIMIT @row_limit::int;

-- name: CreateSession :exec
INSERT INTO study_session (id, user_id, mode, planned_minutes, started_at, expires_at)
VALUES (@id, @user_id, @mode, @minutes::smallint, @now::timestamptz, @expires::timestamptz);

-- name: LockCard :one
SELECT id, item_id, direction, state, stability, due_at, last_review_at, reps, lapses, since_lapse, scheduler
FROM card WHERE id = @id AND user_id = @user_id FOR UPDATE;

-- name: ExpectedTerm :one
SELECT term.body
FROM item i
JOIN course c ON c.id = i.course_id
JOIN item_text term ON term.item_id = i.id AND term.role = 'term' AND term.position = 0 AND term.locale = c.studied_locale
WHERE i.id = @item_id;

-- name: SaveCard :exec
UPDATE card
SET state = @state, stability = @stability::float8, due_at = @due_at, last_review_at = @last_review_at,
    reps = @reps, lapses = @lapses, since_lapse = @since_lapse, scheduler = @scheduler
WHERE id = @id;

-- name: InsertReview :execrows
INSERT INTO review_log (
    id, user_id, card_id, session_id, device_id, exercise_type, source, shown_at, answered_at, client_seq,
    tz, day_cutoff_hour, local_date, response_ms, answer_raw, chosen_item_id, client_correct, is_correct, typo_flag,
    hint_used, explanation_viewed, image_before_answer, difficulty_suggested, difficulty_chosen, error_kind, grade,
    state_before, s_before, r_at_review, elapsed_days, state_after, s_after, scheduled_days, due_after,
    scheduler, scheduler_version)
VALUES (
    @id, @user_id, @card_id, sqlc.narg('session_id'), sqlc.narg('device_id'), @exercise_type, @source, @shown_at, @answered_at, sqlc.narg('client_seq'),
    @tz, @day_cutoff_hour, @local_date, sqlc.narg('response_ms'), sqlc.narg('answer_raw'), sqlc.narg('chosen_item_id'), sqlc.narg('client_correct'), @is_correct, @typo_flag,
    @hint_used, @explanation_viewed, @image_before_answer, sqlc.narg('difficulty_suggested'), sqlc.narg('difficulty_chosen'), sqlc.narg('error_kind'), @grade,
    @state_before, @s_before, @r_at_review, @elapsed_days, @state_after, @s_after, @scheduled_days, @due_after,
    @scheduler, @scheduler_version)
ON CONFLICT (id) DO NOTHING;

-- name: CardLog :many
SELECT answered_at, exercise_type, is_correct, typo_flag, hint_used, explanation_viewed, image_before_answer,
       difficulty_chosen, error_kind
FROM review_log
WHERE card_id = @card_id
ORDER BY answered_at, received_at;
