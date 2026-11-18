-- +goose Up
-- Meta por ciclo (P-03): a pessoa diz quanto quer estudar a cada 7 dias; o app pergunta de novo no fim.
-- So tabela nova (expand). O ciclo e contado em dias de estudo (fuso e corte da pessoa).
CREATE TABLE goal_cycle (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       uuid        NOT NULL REFERENCES app_user (id) ON DELETE CASCADE,
    period_start  date        NOT NULL,
    period_end    date        NOT NULL,
    preset        text        CHECK (preset IN ('light', 'steady', 'intense', 'custom')),
    daily_minutes smallint    NOT NULL CHECK (daily_minutes BETWEEN 1 AND 240),
    days_target   smallint    NOT NULL CHECK (days_target BETWEEN 1 AND 7),
    weekdays      smallint[]  NOT NULL DEFAULT '{1,2,3,4,5,6,7}',   -- 1 = segunda ... 7 = domingo
    status        text        NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'closed')),
    created_at    timestamptz NOT NULL DEFAULT now(),
    CHECK (period_end >= period_start),
    UNIQUE (user_id, period_start)
);
CREATE INDEX goal_cycle_user_idx ON goal_cycle (user_id, period_start DESC);

-- +goose Down
DROP TABLE goal_cycle;
