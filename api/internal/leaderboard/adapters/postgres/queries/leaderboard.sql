-- name: GetSettings :one
SELECT tz, day_cutoff_hour FROM user_settings WHERE user_id = @user_id;

-- name: Ranking :many
-- Pontos de cada pessoa no periodo: os das respostas + 50 por dia com a meta cumprida + 300 por semana de meta com os
-- 7 dias cumpridos (ADR 0032). A posicao conta so quem participa (apelido visivel); empate divide a posicao (RANK).
-- Devolve as row_limit primeiras posicoes e a linha de quem pergunta, mesmo sem participar (nickname vazio).
WITH days AS (
    SELECT user_id, goal_id, points, met
    FROM user_day_score
    WHERE local_date BETWEEN @from_date::date AND @to_date::date
),
full_weeks AS (
    SELECT user_id, count(*)::int AS weeks
    FROM (SELECT user_id, goal_id FROM days WHERE met GROUP BY user_id, goal_id HAVING count(*) >= 7) w
    GROUP BY user_id
),
scored AS (
    SELECT d.user_id,
           (sum(d.points) + 50 * count(*) FILTER (WHERE d.met) + 300 * COALESCE(max(fw.weeks), 0))::int AS points
    FROM days d
    LEFT JOIN full_weeks fw ON fw.user_id = d.user_id
    GROUP BY d.user_id
),
ranked AS (
    SELECT s.user_id, s.points, p.nickname, (RANK() OVER (ORDER BY s.points DESC))::int AS position
    FROM scored s
    JOIN leaderboard_profile p ON p.user_id = s.user_id AND p.visible
    WHERE s.points > 0
),
-- Quem pergunta sem participar: nao entra na contagem dos outros; a posicao dele e so para ele.
hidden_me AS (
    SELECT s.user_id, s.points, ''::text AS nickname,
           (1 + (SELECT count(*) FROM ranked r WHERE r.points > s.points))::int AS position
    FROM scored s
    WHERE s.user_id = @me::uuid AND s.points > 0 AND NOT EXISTS (SELECT 1 FROM ranked r WHERE r.user_id = s.user_id)
)
SELECT user_id, points, nickname, position FROM ranked WHERE position <= @row_limit::int OR user_id = @me::uuid
UNION ALL
SELECT user_id, points, nickname, position FROM hidden_me
ORDER BY position, nickname, user_id;

-- name: DaysMet :one
-- Dias do periodo em que o tempo ativo chegou ao da meta.
SELECT count(*)::int FROM user_day_score
WHERE user_id = @user_id AND met AND local_date BETWEEN @from_date::date AND @to_date::date;

-- name: GetProfile :one
SELECT nickname, visible FROM leaderboard_profile WHERE user_id = @user_id;

-- name: SaveProfile :one
-- O apelido e unico sem diferenciar maiusculas (indice em lower(nickname)); repetido vira erro de unicidade.
INSERT INTO leaderboard_profile (user_id, nickname, visible)
VALUES (@user_id, @nickname, @visible)
ON CONFLICT (user_id) DO UPDATE SET nickname = EXCLUDED.nickname, visible = EXCLUDED.visible, updated_at = now()
RETURNING nickname, visible;
