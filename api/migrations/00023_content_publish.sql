-- GERADO por content/tools/gen_v1.py a partir de content/seed/v1/*.json. NÃO EDITE À MÃO.
-- +goose Up
-- O seed v0 nasceu como rascunho e nenhum caminho o publicava: sem isto, quem não é admin não vê nem estuda nada.
-- Os 125 itens passaram na conferência do verify-v0.sql; as glosas continuam marcadas como assistidas por IA nos créditos.
UPDATE item SET status = 'published', updated_at = now() WHERE course_id = 'f50c7b5b-ad2b-568e-9d46-4e8c0ebd0e27' AND status = 'draft';

-- +goose Down
-- Sem desfazer: o conteúdo é só acrescentado (expand); retirar conteúdo é decisão de contract futura.
SELECT 1;
