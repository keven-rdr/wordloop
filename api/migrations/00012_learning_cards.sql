-- +goose Up
-- Estado de estudo de cada pessoa (modulo learning): ajustes, cartoes e sessoes. So tabelas novas (expand).
CREATE TABLE user_settings (
    user_id         uuid        PRIMARY KEY REFERENCES app_user (id) ON DELETE CASCADE,
    tz              text        NOT NULL DEFAULT 'America/Sao_Paulo',           -- IANA
    day_cutoff_hour smallint    NOT NULL DEFAULT 4 CHECK (day_cutoff_hour BETWEEN 0 AND 8),
    scheduler       text        NOT NULL DEFAULT 'simple_v1' CHECK (scheduler IN ('simple_v1', 'fsrs6')),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- Um cartao por (pessoa, item, direcao): a unidade agendada. O estado e um cache do review_log.
CREATE TABLE card (
    id             uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        uuid        NOT NULL REFERENCES app_user (id) ON DELETE CASCADE,
    item_id        uuid        NOT NULL REFERENCES item (id),
    direction      text        NOT NULL CHECK (direction IN ('receptive', 'productive')),
    state          text        NOT NULL DEFAULT 'new' CHECK (state IN ('new', 'learning', 'review', 'relearning')),
    stability      double precision NOT NULL DEFAULT 0,                              -- S, em dias
    due_at         timestamptz,
    last_review_at timestamptz,
    reps           int         NOT NULL DEFAULT 0,
    lapses         int         NOT NULL DEFAULT 0,
    since_lapse    int         NOT NULL DEFAULT 0,                              -- revisoes desde o ultimo lapso ("dominado")
    scheduler      text        NOT NULL DEFAULT 'simple_v1',
    suspended      boolean     NOT NULL DEFAULT false,
    introduced_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, item_id, direction)
);
CREATE INDEX card_due_idx ON card (user_id, due_at) WHERE state <> 'new' AND NOT suspended;

CREATE TABLE study_session (
    id              uuid        PRIMARY KEY,
    user_id         uuid        NOT NULL REFERENCES app_user (id) ON DELETE CASCADE,
    mode            text        NOT NULL CHECK (mode IN ('review', 'learn', 'mixed')),
    planned_minutes smallint    NOT NULL,
    started_at      timestamptz NOT NULL DEFAULT now(),
    expires_at      timestamptz NOT NULL
);
CREATE INDEX study_session_user_idx ON study_session (user_id, started_at);

-- +goose Down
DROP TABLE study_session;
DROP TABLE card;
DROP TABLE user_settings;
