-- +goose Up
-- Agregacao diaria (P-01): uma linha por pessoa e dia de estudo, atualizada na mesma transacao da revisao.
-- E um cache do review_log (o log continua sendo a verdade). So tabela e visao novas (expand).
CREATE TABLE study_day (
    user_id        uuid   NOT NULL REFERENCES app_user (id) ON DELETE CASCADE,
    local_date     date   NOT NULL,                       -- dia logico (fuso e corte da pessoa)
    active_ms      bigint NOT NULL DEFAULT 0,
    reviews        int    NOT NULL DEFAULT 0,
    correct        int    NOT NULL DEFAULT 0,
    lapses         int    NOT NULL DEFAULT 0,
    new_items      int    NOT NULL DEFAULT 0,             -- 1a revisao do cartao
    known_lexemes  int,                                   -- foto das palavras conhecidas depois da ultima resposta do dia
    PRIMARY KEY (user_id, local_date)
);

-- Preenchimento retroativo a partir do log (o lapso do log e a nota Again num cartao ja aprendido).
INSERT INTO study_day (user_id, local_date, active_ms, reviews, correct, lapses, new_items)
SELECT user_id, local_date,
       COALESCE(sum(response_ms), 0)::bigint,
       count(*)::int,
       (count(*) FILTER (WHERE is_correct))::int,
       (count(*) FILTER (WHERE state_before IN ('review', 'relearning') AND grade = 1))::int,
       (count(*) FILTER (WHERE state_before = 'new'))::int
FROM review_log
GROUP BY user_id, local_date
ON CONFLICT (user_id, local_date) DO NOTHING;

-- Palavra conhecida = lema com cartao receptivo em revisao e estabilidade de 7 dias ou mais.
CREATE VIEW user_lexeme_known AS
SELECT c.user_id, i.lexeme_id
FROM card c
JOIN item i ON i.id = c.item_id
WHERE c.direction = 'receptive' AND c.state = 'review' AND c.stability >= 7 AND i.lexeme_id IS NOT NULL
GROUP BY c.user_id, i.lexeme_id;

-- +goose Down
DROP VIEW user_lexeme_known;
DROP TABLE study_day;
