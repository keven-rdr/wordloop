-- Esquema de CONTEÚDO (Fase 2). PostgreSQL. Rascunho para revisão: não é a migração final.
-- Genérico: subject → course → unit → item. Inglês é só o primeiro curso.
-- Enums viram CHECK/tabelas de apoio (mais fácil de migrar que ENUM nativo).

CREATE TABLE source (                       -- de onde veio o conteúdo (licença do CONTEÚDO, não do código)
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code                text NOT NULL UNIQUE,  -- 'ngsl-1.2', 'tatoeba', 'openmoji', 'wiktionary', 'storyweaver'
  name                text NOT NULL,
  url                 text,
  version             text,
  license_spdx        text NOT NULL,         -- 'CC-BY-SA-4.0', 'CC-BY-2.0-FR', 'CC0-1.0', 'LicenseRef-...'
  attribution_notice  text NOT NULL,         -- texto exato exigido pela licença
  retrieved_at        date,
  citation            text,
  notes               text
);

CREATE TABLE subject (
  id        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code      text NOT NULL UNIQUE,            -- 'english'
  name_key  text NOT NULL                    -- chave i18n da UI
);

CREATE TABLE course (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  subject_id     uuid NOT NULL REFERENCES subject(id),
  slug           text NOT NULL UNIQUE,       -- 'en-for-pt-br'
  studied_locale text NOT NULL,              -- 'en-US'  (o que se estuda)
  help_locale    text NOT NULL,              -- 'pt-BR'  (idioma das explicações)
  status         text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','published','archived'))
);

CREATE TABLE unit (                          -- agrupamento TEMÁTICO (nunca por conjunto semântico)
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  course_id  uuid NOT NULL REFERENCES course(id),
  position   int  NOT NULL,
  theme_key  text NOT NULL,                  -- 'daily-routine', 'food', 'travel'
  UNIQUE (course_id, position)
);

CREATE TABLE lexeme (                        -- lema + classe gramatical; base da cobertura de texto
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  course_id   uuid NOT NULL REFERENCES course(id),
  lemma       text NOT NULL,                 -- normalizado (NFKC, minúsculas)
  pos         text NOT NULL,                 -- 'noun','verb','adj','adv','prep',...
  freq_rank   int,
  freq_band   smallint,                      -- faixa de 500 em 500 (nivelamento e trilha)
  family_key  text,                          -- família de palavras (derivações)
  coverage_share real,                       -- fração dos tokens de textos gerais que este lema cobre (NGSL); base da "cobertura estimada" do painel
  UNIQUE (course_id, lemma, pos)
);

CREATE TABLE item (                          -- unidade de ensino: 1 sentido, 1 pergunta, 1 resposta
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  course_id       uuid NOT NULL REFERENCES course(id),
  unit_id         uuid REFERENCES unit(id),
  lexeme_id       uuid REFERENCES lexeme(id),            -- só para kind in (word, phrasal_verb)
  kind            text NOT NULL CHECK (kind IN
                    ('word','expression','phrasal_verb','sentence','dialogue','paragraph','text')),
  natural_key     text NOT NULL,             -- 'word|bank|noun|1' (chave natural normalizada: idempotência do seed)
  sense_no        smallint,                  -- 1 = sentido principal
  cefr            text CHECK (cefr IN ('A1','A2','B1','B2','C1','C2')),
  concreteness    text CHECK (concreteness IN ('concrete','abstract','function')),
  status          text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','reviewed','published','retired')),
  content_version int  NOT NULL DEFAULT 1,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now(),
  UNIQUE (course_id, natural_key)
);
CREATE INDEX item_unit_idx    ON item (unit_id) WHERE status = 'published';
CREATE INDEX item_lexeme_idx  ON item (lexeme_id);

CREATE TABLE item_text (                     -- todo texto traduzível do conteúdo, por locale
  item_id    uuid NOT NULL REFERENCES item(id) ON DELETE CASCADE,
  role       text NOT NULL CHECK (role IN
               ('term','gloss','definition','explanation','hint','cloze','example','example_gloss')),
  locale     text NOT NULL,
  position   smallint NOT NULL DEFAULT 0,
  body       text NOT NULL,                  -- cloze usa marcação: 'I live in a {{hot}} country'
  source_id  uuid REFERENCES source(id),
  source_ref text,                           -- ex.: id da sentença no Tatoeba
  PRIMARY KEY (item_id, role, locale, position)
);

CREATE TABLE item_relation (                 -- grafo de conexões (tipo + peso)
  from_item_id uuid NOT NULL REFERENCES item(id) ON DELETE CASCADE,
  to_item_id   uuid NOT NULL REFERENCES item(id) ON DELETE CASCADE,
  type         text NOT NULL CHECK (type IN
                 ('collocation','family','synonym','antonym','hypernym','contrast','related')),
  weight       real NOT NULL DEFAULT 0.5 CHECK (weight BETWEEN 0 AND 1),
  min_stability_days real NOT NULL DEFAULT 21,   -- só liberar quando AMBOS tiverem S >= isto (interferência)
  source_id    uuid REFERENCES source(id),
  PRIMARY KEY (from_item_id, to_item_id, type),
  CHECK (from_item_id <> to_item_id)
);

CREATE TABLE media (                         -- arquivos ficam FORA do repositório (volume/objeto); aqui só o manifesto
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  kind             text NOT NULL CHECK (kind IN ('image','audio')),
  storage_key      text NOT NULL UNIQUE,
  sha256           text NOT NULL,
  mime             text NOT NULL,
  width            int, height int, duration_ms int,
  source_id        uuid NOT NULL REFERENCES source(id),
  origin_url       text,
  creator          text,
  license_spdx     text NOT NULL,
  attribution_text text NOT NULL,
  status           text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected'))
);

CREATE TABLE item_media (
  item_id   uuid NOT NULL REFERENCES item(id) ON DELETE CASCADE,
  media_id  uuid NOT NULL REFERENCES media(id),
  role      text NOT NULL CHECK (role IN ('context','cue','pronunciation')),
  position  smallint NOT NULL DEFAULT 0,
  PRIMARY KEY (item_id, media_id, role)
);

CREATE TABLE attribution (                   -- crédito por objeto, quando a licença exige (ex.: autor da sentença)
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  entity_type  text NOT NULL CHECK (entity_type IN ('item','item_text','media')),
  entity_id    uuid NOT NULL,
  source_id    uuid NOT NULL REFERENCES source(id),
  creator      text,
  source_ref   text,
  notice       text
);
CREATE INDEX attribution_entity_idx ON attribution (entity_type, entity_id);

-- Leitura: cobertura = quanto do texto o usuário já conhece (calculada com o estado SRS do usuário, em outra fase)
CREATE TABLE reading_text (
  item_id    uuid PRIMARY KEY REFERENCES item(id) ON DELETE CASCADE,   -- item.kind in (paragraph, text)
  word_count int NOT NULL,
  origin     text NOT NULL CHECK (origin IN ('curated','generated','adapted'))
);
CREATE TABLE reading_text_lexeme (
  text_item_id uuid NOT NULL REFERENCES reading_text(item_id) ON DELETE CASCADE,
  lexeme_id    uuid NOT NULL REFERENCES lexeme(id),
  occurrences  int  NOT NULL CHECK (occurrences > 0),
  PRIMARY KEY (text_item_id, lexeme_id)
);
