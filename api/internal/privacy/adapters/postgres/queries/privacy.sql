-- name: AccountEmail :one
SELECT COALESCE(email, '')::text AS email FROM app_user WHERE id = @user_id AND deleted_at IS NULL;

-- name: ExportData :one
-- Tudo da pessoa num JSON so, na mesma instantanea. As colunas sao SEMPRE listadas: nunca SELECT *, para que um
-- segredo (hash de senha, chave publica de passkey, hash de sessao, token) nao entre por acidente numa coluna nova.
SELECT jsonb_build_object(
    'profile', (SELECT to_jsonb(p) FROM (
        SELECT u.id, u.display_name, u.email, u.email_verified_at, u.role, u.mfa_required, u.created_at, u.last_seen_at,
               (u.password_hash IS NOT NULL) AS has_password
        FROM app_user AS u WHERE u.id = @user_id) p),
    'logins', (SELECT COALESCE(jsonb_agg(to_jsonb(l) ORDER BY l.created_at), '[]'::jsonb) FROM (
        SELECT provider, subject, created_at FROM auth_identity WHERE user_id = @user_id) l),
    'passkeys', (SELECT COALESCE(jsonb_agg(to_jsonb(k) ORDER BY k.created_at), '[]'::jsonb) FROM (
        SELECT name, created_at, last_used_at, transports, backup_eligible, backup_state
        FROM passkey_credential WHERE user_id = @user_id) k),
    'sessions', (SELECT COALESCE(jsonb_agg(to_jsonb(s) ORDER BY s.created_at), '[]'::jsonb) FROM (
        SELECT public_id AS id, created_at, authenticated_at, last_used_at, expires_at, user_agent
        FROM auth_session WHERE user_id = @user_id) s),
    'preferences', (SELECT to_jsonb(u) FROM (
        SELECT tz, day_cutoff_hour, scheduler, ui_locale, theme, tz_auto, updated_at
        FROM user_settings WHERE user_id = @user_id) u),
    'goals', (SELECT COALESCE(jsonb_agg(to_jsonb(g) ORDER BY g.period_start), '[]'::jsonb) FROM (
        SELECT id, period_start, period_end, preset, daily_minutes, days_target, weekdays, status, created_at
        FROM goal_cycle WHERE user_id = @user_id) g),
    'study_days', (SELECT COALESCE(jsonb_agg(to_jsonb(d) ORDER BY d.local_date), '[]'::jsonb) FROM (
        SELECT local_date, active_ms, reviews, correct, lapses, new_items, known_lexemes, points
        FROM study_day WHERE user_id = @user_id) d),
    'ranking', (SELECT to_jsonb(r) FROM (
        SELECT nickname, visible, updated_at FROM leaderboard_profile WHERE user_id = @user_id) r),
    'study_sessions', (SELECT COALESCE(jsonb_agg(to_jsonb(x) ORDER BY x.started_at), '[]'::jsonb) FROM (
        SELECT id, mode, planned_minutes, started_at, expires_at
        FROM study_session WHERE user_id = @user_id) x),
    'cards', (SELECT COALESCE(jsonb_agg(to_jsonb(c) ORDER BY c.introduced_at), '[]'::jsonb) FROM (
        SELECT c.id, i.natural_key AS item, term.body AS term, c.direction, c.state, c.stability, c.due_at,
               c.last_review_at, c.reps, c.lapses, c.suspended, c.introduced_at
        FROM card c
        JOIN item i ON i.id = c.item_id
        LEFT JOIN course co ON co.id = i.course_id
        LEFT JOIN item_text term ON term.item_id = i.id AND term.role = 'term' AND term.position = 0
                                AND term.locale = co.studied_locale
        WHERE c.user_id = @user_id) c),
    'reviews', (SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY r.answered_at, r.received_at), '[]'::jsonb) FROM (
        SELECT id, card_id, session_id, exercise_type, source, shown_at, answered_at, received_at, tz, day_cutoff_hour,
               local_date, response_ms, answer_raw, chosen_item_id, client_correct, is_correct, typo_flag, hint_used,
               explanation_viewed, image_before_answer, difficulty_suggested, difficulty_chosen, error_kind, grade,
               state_before, s_before, r_at_review, elapsed_days, state_after, s_after, scheduled_days, due_after,
               scheduler, scheduler_version
        FROM review_log WHERE user_id = @user_id) r)
)::jsonb AS data;

-- name: SetErasingUser :one
-- Declara, so nesta transacao, de quem o gatilho do review_log pode apagar as linhas.
SELECT set_config('wordloop.erase_user', @user_id::text, true)::text AS value;

-- name: EraseReviewVoids :exec
DELETE FROM review_void WHERE review_id IN (SELECT id FROM review_log WHERE user_id = @user_id);

-- name: EraseReviews :execrows
DELETE FROM review_log WHERE user_id = @user_id;

-- name: EraseAccount :execrows
-- O resto sai por ON DELETE CASCADE: vinculos, passkeys, sessoes, desafios, tokens, preferencias, cartoes,
-- sessoes de estudo, dias de estudo, metas e o perfil do ranking.
DELETE FROM app_user WHERE id = @user_id;

-- name: RecordErasure :exec
INSERT INTO account_erasure (review_count) VALUES (@review_count);
