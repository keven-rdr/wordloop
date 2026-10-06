-- Esquema de ESTADO DO USUÁRIO (Fase 3). PostgreSQL. Complementa docs/fase-2/content-schema.sql.
-- Rascunho para revisão: não é a migração final (migrações reais: goose, uma por mudança).
-- Identidade (e-mail, senha, vínculos Google/GitHub) fica no Keycloak; aqui só o perfil do app.

-- ───────── identidade e sessão (módulo identity) ─────────
CREATE TABLE app_user (
  id            uuid PRIMARY KEY,                 -- = claim "sub" do Keycloak (provisionado no 1º login)
  display_name  text,
  created_at    timestamptz NOT NULL DEFAULT now(),
  last_seen_at  timestamptz,
  deleted_at    timestamptz                       -- exclusão LGPD: marca e depois apaga dados (ver docs)
);

CREATE TABLE user_settings (
  user_id            uuid PRIMARY KEY REFERENCES app_user(id) ON DELETE CASCADE,
  ui_locale          text NOT NULL DEFAULT 'pt-BR',
  tz                 text NOT NULL DEFAULT 'America/Sao_Paulo',      -- IANA
  day_cutoff_hour    smallint NOT NULL DEFAULT 4 CHECK (day_cutoff_hour BETWEEN 0 AND 8),
  scheduler          text NOT NULL DEFAULT 'simple_v1',               -- 'simple_v1' | 'fsrs6'
  desired_retention  real NOT NULL DEFAULT 0.90 CHECK (desired_retention BETWEEN 0.70 AND 0.97),  -- avançado
  theme              text NOT NULL DEFAULT 'system' CHECK (theme IN ('light','dark','system')),
  sounds_enabled     boolean NOT NULL DEFAULT false,
  updated_at         timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE auth_session (                         -- BFF: sessão no servidor; o navegador só tem cookie httpOnly
  id_hash            bytea PRIMARY KEY,              -- SHA-256 do id do cookie (nunca o id em claro)
  user_id            uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  refresh_token_enc  bytea NOT NULL,                 -- cifrado (AES-GCM, chave em segredo do ambiente)
  created_at         timestamptz NOT NULL DEFAULT now(),
  last_used_at       timestamptz NOT NULL DEFAULT now(),
  expires_at         timestamptz NOT NULL,
  user_agent         text
);
CREATE INDEX auth_session_user_idx ON auth_session (user_id);

CREATE TABLE device (
  id            uuid PRIMARY KEY,                    -- gerado no cliente, guardado no IndexedDB
  user_id       uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  label         text,
  platform      text CHECK (platform IN ('android','ios','desktop','other')),
  app_version   text,
  created_at    timestamptz NOT NULL DEFAULT now(),
  last_seen_at  timestamptz
);

-- ───────── aprendizado (módulo learning) ─────────
CREATE TABLE card (                                 -- 1 por (usuário, item, direção): unidade agendada (Fase 1)
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id          uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  item_id          uuid NOT NULL REFERENCES item(id),
  direction        text NOT NULL CHECK (direction IN ('receptive','productive')),
  state            text NOT NULL DEFAULT 'new' CHECK (state IN ('new','learning','review','relearning')),
  stability        real,                              -- S (dias)
  difficulty       real,                              -- D (1..10), só FSRS
  due_at           timestamptz,
  last_review_at   timestamptz,
  reps             int NOT NULL DEFAULT 0,
  lapses           int NOT NULL DEFAULT 0,
  scheduler        text NOT NULL,                     -- estratégia que produziu o estado
  memory           jsonb NOT NULL DEFAULT '{}',       -- campos específicos da estratégia
  leech            boolean NOT NULL DEFAULT false,
  suspended        boolean NOT NULL DEFAULT false,
  introduced_at    timestamptz,
  UNIQUE (user_id, item_id, direction)
);
CREATE INDEX card_due_idx ON card (user_id, due_at) WHERE state <> 'new' AND NOT suspended;
CREATE INDEX card_lapses_idx ON card (user_id, lapses DESC) WHERE lapses > 0;   -- "mais erradas"

CREATE TABLE study_session (
  id               uuid PRIMARY KEY,                  -- gerado no cliente
  user_id          uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  device_id        uuid REFERENCES device(id),
  mode             text NOT NULL CHECK (mode IN ('review','learn','challenge','placement','reading')),
  planned_minutes  smallint,                          -- "quanto tempo hoje?" define o tamanho da sessão
  started_at       timestamptz NOT NULL,
  ended_at         timestamptz
);

CREATE TABLE review_log (                           -- APPEND-ONLY (ver trigger abaixo). Campos: Fase 1, seção 8
  id                   uuid PRIMARY KEY,              -- gerado no cliente: idempotência
  user_id              uuid NOT NULL REFERENCES app_user(id),
  card_id              uuid NOT NULL REFERENCES card(id),
  session_id           uuid REFERENCES study_session(id),
  device_id            uuid REFERENCES device(id),
  exercise_type        text NOT NULL,
  source               text NOT NULL CHECK (source IN ('session','notification','placement','reading')),
  shown_at             timestamptz NOT NULL,          -- relógio do cliente
  answered_at          timestamptz NOT NULL,          -- relógio do cliente
  received_at          timestamptz NOT NULL DEFAULT now(),   -- relógio do servidor
  client_seq           bigint,
  tz                   text NOT NULL,
  day_cutoff_hour      smallint NOT NULL,
  local_date           date NOT NULL,                 -- dia lógico calculado no servidor (4h, fuso do usuário)
  response_ms          int,
  is_correct           boolean NOT NULL,
  answer_raw           text,
  typo_flag            boolean NOT NULL DEFAULT false,
  hint_used            boolean NOT NULL DEFAULT false,
  explanation_viewed   boolean NOT NULL DEFAULT false,
  image_before_answer  boolean NOT NULL DEFAULT false,
  difficulty_suggested text CHECK (difficulty_suggested IN ('hard','medium','easy')),
  difficulty_chosen    text CHECK (difficulty_chosen IN ('hard','medium','easy')),
  error_kind           text CHECK (error_kind IN ('no_idea','almost','slip')),
  grade                smallint NOT NULL CHECK (grade BETWEEN 1 AND 4),
  state_before         text NOT NULL,  s_before real, d_before real, r_at_review real, elapsed_days real,
  state_after          text NOT NULL,  s_after  real, d_after  real, scheduled_days real, due_after timestamptz,
  scheduler            text NOT NULL,
  scheduler_version    text NOT NULL,
  params_id            text,
  app_version          text
);
CREATE INDEX review_log_user_time_idx ON review_log (user_id, answered_at);
CREATE INDEX review_log_card_time_idx ON review_log (card_id, answered_at);
CREATE INDEX review_log_user_day_idx  ON review_log (user_id, local_date);

CREATE FUNCTION review_log_immutable() RETURNS trigger LANGUAGE plpgsql AS
$$ BEGIN RAISE EXCEPTION 'review_log é append-only'; END $$;
CREATE TRIGGER review_log_no_update BEFORE UPDATE OR DELETE ON review_log
  FOR EACH ROW EXECUTE FUNCTION review_log_immutable();
-- Exclusão de conta (LGPD): rotina própria desabilita o gatilho numa transação controlada e é registrada.

CREATE TABLE review_void (                          -- correção = novo evento, nunca UPDATE
  id         uuid PRIMARY KEY,
  review_id  uuid NOT NULL REFERENCES review_log(id),
  reason     text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- ───────── progresso (módulo progress) ─────────
CREATE TABLE goal_cycle (                           -- meta DINÂMICA: o app pergunta a cada ciclo (semana)
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id        uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  period_start   date NOT NULL,                      -- dia lógico
  period_end     date NOT NULL,
  preset         text CHECK (preset IN ('light','steady','intense','custom')),
  daily_minutes  smallint NOT NULL CHECK (daily_minutes BETWEEN 1 AND 240),
  days_target    smallint NOT NULL CHECK (days_target BETWEEN 1 AND 7),
  weekdays       smallint[] NOT NULL DEFAULT '{1,2,3,4,5,6,7}',
  status         text NOT NULL DEFAULT 'active' CHECK (status IN ('active','closed')),
  UNIQUE (user_id, period_start)
);

CREATE TABLE study_day (                            -- agregação diária pré-calculada (heatmap, acertos, streak)
  user_id           uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  local_date        date NOT NULL,
  active_ms         bigint NOT NULL DEFAULT 0,
  reviews           int NOT NULL DEFAULT 0,
  correct           int NOT NULL DEFAULT 0,
  lapses            int NOT NULL DEFAULT 0,
  new_items         int NOT NULL DEFAULT 0,
  known_lexemes     int,                               -- foto no fim do dia (cobertura / "melhora")
  mastered_items    int,
  PRIMARY KEY (user_id, local_date)
);
-- "Estudou no dia" (streak, heatmap) = reviews >= 1: o que importa é entrar e exercitar, não minutos.

CREATE VIEW user_lexeme_known AS                    -- lema conhecido = receptivo em review com S >= 7 d (Fase 2)
  SELECT c.user_id, i.lexeme_id
  FROM card c JOIN item i ON i.id = c.item_id
  WHERE c.direction = 'receptive' AND c.state = 'review' AND c.stability >= 7 AND i.lexeme_id IS NOT NULL
  GROUP BY c.user_id, i.lexeme_id;

-- ───────── lembretes e push (módulo reminders) ─────────
CREATE TABLE notification_pref (
  user_id        uuid PRIMARY KEY REFERENCES app_user(id) ON DELETE CASCADE,
  enabled        boolean NOT NULL DEFAULT false,
  window_start   time NOT NULL DEFAULT '09:00',
  window_end     time NOT NULL DEFAULT '21:00',
  quiet_start    time,
  quiet_end      time,
  weekdays       smallint[] NOT NULL DEFAULT '{1,2,3,4,5,6,7}',
  max_per_day    smallint NOT NULL DEFAULT 3 CHECK (max_per_day BETWEEN 1 AND 10),
  hide_content   boolean NOT NULL DEFAULT false       -- não mostrar a pergunta na tela de bloqueio
);

CREATE TABLE push_subscription (                    -- 1 por dispositivo/navegador
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id          uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  device_id        uuid REFERENCES device(id) ON DELETE SET NULL,
  endpoint         text NOT NULL UNIQUE,
  p256dh           text NOT NULL,
  auth             text NOT NULL,
  status           text NOT NULL DEFAULT 'active' CHECK (status IN ('active','expired','revoked')),
  failure_count    smallint NOT NULL DEFAULT 0,
  created_at       timestamptz NOT NULL DEFAULT now(),
  last_success_at  timestamptz
);

CREATE TABLE reminder_slot (                        -- horários sorteados por dia dentro da janela (determinístico, testável)
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  local_date  date NOT NULL,
  fire_at     timestamptz NOT NULL,
  status      text NOT NULL DEFAULT 'planned' CHECK (status IN ('planned','sent','skipped')),
  UNIQUE (user_id, fire_at)
);
CREATE INDEX reminder_slot_due_idx ON reminder_slot (fire_at) WHERE status = 'planned';

CREATE TABLE notification (                         -- o desafio enviado (a pergunta NÃO traz a resposta)
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  kind         text NOT NULL DEFAULT 'challenge',
  card_id      uuid REFERENCES card(id),
  payload      jsonb NOT NULL,
  created_at   timestamptz NOT NULL DEFAULT now(),
  expires_at   timestamptz NOT NULL,
  opened_at    timestamptz,
  review_id    uuid REFERENCES review_log(id)
);

CREATE TABLE notification_delivery (                -- log de falhas e tentativas
  id               bigserial PRIMARY KEY,
  notification_id  uuid NOT NULL REFERENCES notification(id) ON DELETE CASCADE,
  subscription_id  uuid NOT NULL REFERENCES push_subscription(id) ON DELETE CASCADE,
  status           text NOT NULL CHECK (status IN ('sent','failed','gone')),   -- 'gone' = 404/410 → expira a assinatura
  http_status      int,
  error            text,
  attempted_at     timestamptz NOT NULL DEFAULT now()
);

-- ───────── plataforma: outbox e idempotência ─────────
CREATE TABLE outbox (                               -- fila em Postgres; worker usa FOR UPDATE SKIP LOCKED
  id            bigserial PRIMARY KEY,
  topic         text NOT NULL,                        -- 'push.send'
  event_key     uuid NOT NULL,
  payload       jsonb NOT NULL,
  created_at    timestamptz NOT NULL DEFAULT now(),
  available_at  timestamptz NOT NULL DEFAULT now(),   -- backoff das retentativas
  attempts      smallint NOT NULL DEFAULT 0,
  last_error    text,
  processed_at  timestamptz,
  UNIQUE (topic, event_key)
);
CREATE INDEX outbox_ready_idx ON outbox (available_at) WHERE processed_at IS NULL;

CREATE TABLE processed_event (                      -- idempotência por consumidor (modelo do Contratos)
  consumer      text NOT NULL,
  event_id      uuid NOT NULL,
  processed_at  timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (consumer, event_id)
);
