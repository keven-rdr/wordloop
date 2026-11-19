-- +goose Up
-- Expand: o cartao produtivo de uma frase ganha o exercicio de montar a frase com pecas (assemble).
-- So alarga o CHECK; nenhuma linha existente muda (o log e imutavel).
ALTER TABLE review_log DROP CONSTRAINT review_log_exercise_type_check;
ALTER TABLE review_log ADD CONSTRAINT review_log_exercise_type_check
    CHECK (exercise_type IN ('choose', 'type', 'cloze', 'image', 'assemble'));

-- +goose Down
ALTER TABLE review_log DROP CONSTRAINT review_log_exercise_type_check;
ALTER TABLE review_log ADD CONSTRAINT review_log_exercise_type_check
    CHECK (exercise_type IN ('choose', 'type', 'cloze', 'image'));
