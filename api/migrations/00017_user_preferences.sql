-- +goose Up
-- Preferencias da pessoa (W-01): idioma, tema e se o fuso acompanha o aparelho. So colunas novas com padrao (expand):
-- quem ja tem linha continua como estava (tz_auto = true e o comportamento de antes: o plano guarda o fuso do aparelho).
ALTER TABLE user_settings
    ADD COLUMN ui_locale text    NOT NULL DEFAULT 'system' CHECK (ui_locale IN ('system', 'pt-BR', 'en-US')),
    ADD COLUMN theme     text    NOT NULL DEFAULT 'system' CHECK (theme IN ('system', 'light', 'dark')),
    ADD COLUMN tz_auto   boolean NOT NULL DEFAULT true;

-- +goose Down
ALTER TABLE user_settings DROP COLUMN tz_auto, DROP COLUMN theme, DROP COLUMN ui_locale;
