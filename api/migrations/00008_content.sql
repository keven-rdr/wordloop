-- +goose Up
CREATE TABLE source (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code               text NOT NULL UNIQUE,
    name               text NOT NULL,
    url                text,
    version            text,
    license_spdx       text NOT NULL,
    attribution_notice text NOT NULL,
    retrieved_at       date,
    citation           text,
    notes              text
);

CREATE TABLE subject (
    id       uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code     text NOT NULL UNIQUE,
    name_key text NOT NULL
);

CREATE TABLE course (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_id     uuid NOT NULL REFERENCES subject (id),
    slug           text NOT NULL UNIQUE,
    studied_locale text NOT NULL,
    help_locale    text NOT NULL,
    status         text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived'))
);

CREATE TABLE unit (
    id        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id uuid NOT NULL REFERENCES course (id),
    position  int  NOT NULL,
    theme_key text NOT NULL,
    UNIQUE (course_id, position)
);

CREATE TABLE lexeme (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id      uuid NOT NULL REFERENCES course (id),
    lemma          text NOT NULL,
    pos            text NOT NULL,
    freq_rank      int,
    freq_band      smallint,
    family_key     text,
    coverage_share real,
    UNIQUE (course_id, lemma, pos)
);

CREATE TABLE item (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id       uuid NOT NULL REFERENCES course (id),
    unit_id         uuid REFERENCES unit (id),
    lexeme_id       uuid REFERENCES lexeme (id),
    kind            text NOT NULL CHECK (kind IN
        ('word', 'expression', 'phrasal_verb', 'sentence', 'dialogue', 'paragraph', 'text')),
    natural_key     text NOT NULL,
    sense_no        smallint,
    cefr            text CHECK (cefr IN ('A1', 'A2', 'B1', 'B2', 'C1', 'C2')),
    concreteness    text CHECK (concreteness IN ('concrete', 'abstract', 'function')),
    status          text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'reviewed', 'published', 'retired')),
    content_version int  NOT NULL DEFAULT 1,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (course_id, natural_key)
);
CREATE INDEX item_unit_idx   ON item (unit_id);
CREATE INDEX item_lexeme_idx ON item (lexeme_id);
CREATE INDEX item_listing_idx ON item (course_id, kind, status);

CREATE TABLE item_text (
    item_id    uuid     NOT NULL REFERENCES item (id) ON DELETE CASCADE,
    role       text     NOT NULL CHECK (role IN
        ('term', 'gloss', 'definition', 'explanation', 'hint', 'cloze', 'example', 'example_gloss')),
    locale     text     NOT NULL,
    position   smallint NOT NULL DEFAULT 0,
    body       text     NOT NULL,
    source_id  uuid REFERENCES source (id),
    source_ref text,
    PRIMARY KEY (item_id, role, locale, position)
);

CREATE TABLE item_relation (
    from_item_id       uuid NOT NULL REFERENCES item (id) ON DELETE CASCADE,
    to_item_id         uuid NOT NULL REFERENCES item (id) ON DELETE CASCADE,
    type               text NOT NULL CHECK (type IN
        ('collocation', 'family', 'synonym', 'antonym', 'hypernym', 'contrast', 'related')),
    weight             real NOT NULL DEFAULT 0.5 CHECK (weight BETWEEN 0 AND 1),
    min_stability_days real NOT NULL DEFAULT 21,
    source_id          uuid REFERENCES source (id),
    PRIMARY KEY (from_item_id, to_item_id, type),
    CHECK (from_item_id <> to_item_id)
);

CREATE TABLE media (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    kind             text NOT NULL CHECK (kind IN ('image', 'audio')),
    storage_key      text NOT NULL UNIQUE,
    sha256           text NOT NULL,
    mime             text NOT NULL,
    width            int,
    height           int,
    duration_ms      int,
    source_id        uuid NOT NULL REFERENCES source (id),
    origin_url       text,
    creator          text,
    license_spdx     text NOT NULL,
    attribution_text text NOT NULL,
    status           text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected'))
);

CREATE TABLE item_media (
    item_id  uuid     NOT NULL REFERENCES item (id) ON DELETE CASCADE,
    media_id uuid     NOT NULL REFERENCES media (id),
    role     text     NOT NULL CHECK (role IN ('context', 'cue', 'pronunciation')),
    position smallint NOT NULL DEFAULT 0,
    PRIMARY KEY (item_id, media_id, role)
);

CREATE TABLE attribution (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type text NOT NULL CHECK (entity_type IN ('item', 'item_text', 'media')),
    entity_id   uuid NOT NULL,
    source_id   uuid NOT NULL REFERENCES source (id),
    creator     text,
    source_ref  text,
    notice      text
);
CREATE INDEX attribution_entity_idx ON attribution (entity_type, entity_id);

-- +goose Down
DROP TABLE attribution;
DROP TABLE item_media;
DROP TABLE media;
DROP TABLE item_relation;
DROP TABLE item_text;
DROP TABLE item;
DROP TABLE lexeme;
DROP TABLE unit;
DROP TABLE course;
DROP TABLE subject;
DROP TABLE source;
