-- +goose Up
-- Gamificacao (ADR 0032). So acrescenta (expand): uma coluna, uma visao e uma tabela.

-- Pontos do dia, somados na mesma transacao de cada resposta (como as demais colunas do study_day). A tabela de pontos e
-- a do learning/domain.ReviewPoints: acerto 10 (reconhecimento) ou 15 (producao), erro 2; acerto x nivel do item
-- (A1 1; A2 1,2; B1 1,5; B2 ou mais 2); teto de 3.000 por dia.
ALTER TABLE study_day ADD COLUMN points int NOT NULL DEFAULT 0 CHECK (points >= 0);

-- Os dias que ja existem ganham os pontos recalculados a partir do log (a verdade).
UPDATE study_day d SET points = x.points
FROM (
    SELECT r.user_id, r.local_date, LEAST(3000, sum(
        CASE WHEN NOT r.is_correct THEN 2
             ELSE round(
                 (CASE WHEN r.exercise_type IN ('type', 'cloze', 'image', 'assemble') THEN 15 ELSE 10 END)
                 * (CASE WHEN i.cefr IS NULL OR i.cefr = 'A1' THEN 1.0 WHEN i.cefr = 'A2' THEN 1.2
                         WHEN i.cefr = 'B1' THEN 1.5 ELSE 2.0 END))
        END))::int AS points
    FROM review_log r
    JOIN card c ON c.id = r.card_id
    JOIN item i ON i.id = c.item_id
    WHERE NOT EXISTS (SELECT 1 FROM review_void v WHERE v.review_id = r.id) -- respostas anuladas nao pontuam
    GROUP BY r.user_id, r.local_date
) x
WHERE d.user_id = x.user_id AND d.local_date = x.local_date;

-- Pontos de cada dia com a meta: met = o tempo ativo do dia chegou ao tempo da meta que valia naquele dia.
-- goal_id identifica a semana da meta (para o bonus da semana inteira).
CREATE VIEW user_day_score AS
SELECT d.user_id, d.local_date, d.points, g.id AS goal_id,
       (g.id IS NOT NULL AND d.active_ms >= g.daily_minutes::bigint * 60000) AS met
FROM study_day d
LEFT JOIN LATERAL (
    SELECT gc.id, gc.daily_minutes
    FROM goal_cycle gc
    WHERE gc.user_id = d.user_id AND d.local_date BETWEEN gc.period_start AND gc.period_end
    ORDER BY gc.period_start DESC
    LIMIT 1
) g ON true;

-- Perfil do ranking: so aparece para os outros quem escolhe um apelido e liga a participacao.
CREATE TABLE leaderboard_profile (
    user_id    uuid        PRIMARY KEY REFERENCES app_user (id) ON DELETE CASCADE,
    nickname   text        NOT NULL CHECK (nickname ~ '^[A-Za-z0-9_]{3,20}$'),
    visible    boolean     NOT NULL DEFAULT false,
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX leaderboard_profile_nickname_idx ON leaderboard_profile (lower(nickname));

-- +goose Down
DROP TABLE leaderboard_profile;
DROP VIEW user_day_score;
ALTER TABLE study_day DROP COLUMN points;
