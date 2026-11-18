-- +goose Up
-- Expand: o cartao produtivo ganha duas variantes de exercicio, lacuna (cloze) e imagem -> palavra (image).
-- So alarga o CHECK; nenhuma linha existente muda (o log e imutavel).
ALTER TABLE review_log DROP CONSTRAINT review_log_exercise_type_check;
ALTER TABLE review_log ADD CONSTRAINT review_log_exercise_type_check
    CHECK (exercise_type IN ('choose', 'type', 'cloze', 'image'));

-- +goose Down
ALTER TABLE review_log DROP CONSTRAINT review_log_exercise_type_check;
ALTER TABLE review_log ADD CONSTRAINT review_log_exercise_type_check
    CHECK (exercise_type IN ('choose', 'type'));
