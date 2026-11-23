-- +goose Up
-- "Ja sabia" (known): no primeiro contato com a palavra, a pessoa diz que ja a conhecia e ela volta so em ~3 semanas.
-- Expand: so alarga a lista de valores aceitos; nenhuma linha muda (o review_log e append-only).
ALTER TABLE review_log DROP CONSTRAINT IF EXISTS review_log_difficulty_chosen_check;
ALTER TABLE review_log ADD CONSTRAINT review_log_difficulty_chosen_check
    CHECK (difficulty_chosen IN ('hard', 'medium', 'easy', 'known'));

-- +goose Down
-- Sem desfazer: voltar a lista apertada falharia com linhas 'known' ja gravadas no log imutavel.
SELECT 1;
