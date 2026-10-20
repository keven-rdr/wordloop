-- +goose Up
-- Log imutavel de revisoes: so INSERT (o gatilho recusa UPDATE e DELETE). Correcoes entram como evento em review_void.
CREATE TABLE review_log (
    id                   uuid        PRIMARY KEY,                       -- gerado no cliente: idempotencia
    user_id              uuid        NOT NULL REFERENCES app_user (id),
    card_id              uuid        NOT NULL REFERENCES card (id),
    session_id           uuid        REFERENCES study_session (id),
    device_id            uuid,
    exercise_type        text        NOT NULL CHECK (exercise_type IN ('choose', 'type')),
    source               text        NOT NULL CHECK (source IN ('session', 'notification', 'placement', 'reading')),
    shown_at             timestamptz NOT NULL,                          -- relogio do cliente
    answered_at          timestamptz NOT NULL,                          -- relogio do cliente
    received_at          timestamptz NOT NULL DEFAULT now(),            -- relogio do servidor
    client_seq           bigint,
    tz                   text        NOT NULL,
    day_cutoff_hour      smallint    NOT NULL,
    local_date           date        NOT NULL,                          -- dia logico calculado no servidor
    response_ms          int,
    answer_raw           text,
    chosen_item_id       uuid,
    client_correct       boolean,                                       -- o que o app mostrou (so registro)
    is_correct           boolean     NOT NULL,                          -- a correcao do servidor
    typo_flag            boolean     NOT NULL DEFAULT false,
    hint_used            boolean     NOT NULL DEFAULT false,
    explanation_viewed   boolean     NOT NULL DEFAULT false,
    image_before_answer  boolean     NOT NULL DEFAULT false,
    difficulty_suggested text        CHECK (difficulty_suggested IN ('hard', 'medium', 'easy')),
    difficulty_chosen    text        CHECK (difficulty_chosen IN ('hard', 'medium', 'easy')),
    error_kind           text        CHECK (error_kind IN ('no_idea', 'almost', 'slip')),
    grade                smallint    NOT NULL CHECK (grade BETWEEN 1 AND 4),
    state_before         text        NOT NULL,
    s_before double precision,
    r_at_review double precision,
    elapsed_days double precision,
    state_after          text        NOT NULL,
    s_after double precision,
    scheduled_days double precision,
    due_after            timestamptz,
    scheduler            text        NOT NULL,
    scheduler_version    text        NOT NULL
);
CREATE INDEX review_log_user_time_idx ON review_log (user_id, answered_at);
CREATE INDEX review_log_card_time_idx ON review_log (card_id, answered_at);
CREATE INDEX review_log_user_day_idx ON review_log (user_id, local_date);

-- +goose StatementBegin
CREATE FUNCTION review_log_immutable() RETURNS trigger LANGUAGE plpgsql AS
$$ BEGIN RAISE EXCEPTION 'review_log é append-only'; END $$;
-- +goose StatementEnd

CREATE TRIGGER review_log_no_update BEFORE UPDATE OR DELETE ON review_log
    FOR EACH ROW EXECUTE FUNCTION review_log_immutable();

CREATE TABLE review_void (                                              -- correcao = novo evento, nunca UPDATE
    id         uuid        PRIMARY KEY,
    review_id  uuid        NOT NULL REFERENCES review_log (id),
    reason     text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE review_void;
DROP TRIGGER review_log_no_update ON review_log;
DROP FUNCTION review_log_immutable();
DROP TABLE review_log;
