-- GERADO por gen_sql.py. Reconcilia o banco com a fonte.
-- Bloco 1: contagens (status deve ser 'ok'). Bloco 2: violações (devem retornar 0 linhas).
SELECT 'item.word' AS check_name, COUNT(*) AS found, 31 AS expected, CASE WHEN COUNT(*) = 31 THEN 'ok' ELSE 'FALHA' END AS status FROM item WHERE course_id = 'd888eee8-9eb3-515b-a2f0-c255d13be4f7' AND kind = 'word';
SELECT 'item.expression' AS check_name, COUNT(*) AS found, 7 AS expected, CASE WHEN COUNT(*) = 7 THEN 'ok' ELSE 'FALHA' END AS status FROM item WHERE course_id = 'd888eee8-9eb3-515b-a2f0-c255d13be4f7' AND kind = 'expression';
SELECT 'item.phrasal_verb' AS check_name, COUNT(*) AS found, 3 AS expected, CASE WHEN COUNT(*) = 3 THEN 'ok' ELSE 'FALHA' END AS status FROM item WHERE course_id = 'd888eee8-9eb3-515b-a2f0-c255d13be4f7' AND kind = 'phrasal_verb';
SELECT 'item.sentence' AS check_name, COUNT(*) AS found, 10 AS expected, CASE WHEN COUNT(*) = 10 THEN 'ok' ELSE 'FALHA' END AS status FROM item WHERE course_id = 'd888eee8-9eb3-515b-a2f0-c255d13be4f7' AND kind = 'sentence';
SELECT 'item.total' AS check_name, COUNT(*) AS found, 51 AS expected, CASE WHEN COUNT(*) = 51 THEN 'ok' ELSE 'FALHA' END AS status FROM item WHERE course_id = 'd888eee8-9eb3-515b-a2f0-c255d13be4f7';
SELECT 'item_text.total' AS check_name, COUNT(*) AS found, 194 AS expected, CASE WHEN COUNT(*) = 194 THEN 'ok' ELSE 'FALHA' END AS status FROM item_text WHERE item_id IN (SELECT id FROM item WHERE course_id = 'd888eee8-9eb3-515b-a2f0-c255d13be4f7');
SELECT 'item_relation.total' AS check_name, COUNT(*) AS found, 8 AS expected, CASE WHEN COUNT(*) = 8 THEN 'ok' ELSE 'FALHA' END AS status FROM item_relation WHERE from_item_id IN (SELECT id FROM item WHERE course_id = 'd888eee8-9eb3-515b-a2f0-c255d13be4f7');
SELECT 'media.total' AS check_name, COUNT(*) AS found, 10 AS expected, CASE WHEN COUNT(*) = 10 THEN 'ok' ELSE 'FALHA' END AS status FROM media WHERE source_id IN (SELECT id FROM source WHERE code = 'openmoji');
-- violação: item sem gloss pt-BR
SELECT 'item sem gloss pt-BR' AS violation, x.* FROM (SELECT i.natural_key FROM item i WHERE i.course_id = 'd888eee8-9eb3-515b-a2f0-c255d13be4f7' AND NOT EXISTS (SELECT 1 FROM item_text t WHERE t.item_id = i.id AND t.role = 'gloss' AND t.locale = 'pt-BR')) AS x;
-- violação: item publicado sem revisão (todo item da amostra deve ser draft)
SELECT 'item publicado sem revisão (todo item da amostra deve ser draft)' AS violation, x.* FROM (SELECT natural_key FROM item WHERE course_id = 'd888eee8-9eb3-515b-a2f0-c255d13be4f7' AND status <> 'draft') AS x;
-- violação: cloze sem exatamente uma lacuna
SELECT 'cloze sem exatamente uma lacuna' AS violation, x.* FROM (SELECT item_id FROM item_text WHERE role = 'cloze' AND (LENGTH(body) - LENGTH(REPLACE(body, '{{', ''))) <> 2) AS x;
-- violação: item_text sem fonte
SELECT 'item_text sem fonte' AS violation, x.* FROM (SELECT item_id FROM item_text WHERE source_id IS NULL AND item_id IN (SELECT id FROM item WHERE course_id = 'd888eee8-9eb3-515b-a2f0-c255d13be4f7')) AS x;
-- violação: palavra sem sentido 1 no seu lema
SELECT 'palavra sem sentido 1 no seu lema' AS violation, x.* FROM (SELECT l.lemma, l.pos FROM lexeme l WHERE l.course_id = 'd888eee8-9eb3-515b-a2f0-c255d13be4f7' AND EXISTS (SELECT 1 FROM item i WHERE i.lexeme_id = l.id AND i.kind = 'word') AND NOT EXISTS (SELECT 1 FROM item i WHERE i.lexeme_id = l.id AND i.kind = 'word' AND i.sense_no = 1)) AS x;
-- violação: mídia sem licença ou crédito
SELECT 'mídia sem licença ou crédito' AS violation, x.* FROM (SELECT storage_key FROM media WHERE license_spdx IS NULL OR attribution_text IS NULL OR attribution_text = '') AS x;
-- violação: mídia aprovada sem hash real (deve estar pending até a ingestão)
SELECT 'mídia aprovada sem hash real (deve estar pending até a ingestão)' AS violation, x.* FROM (SELECT storage_key FROM media WHERE status = 'approved' AND sha256 LIKE 'pending:%') AS x;
-- violação: relação consigo mesma ou fora do curso
SELECT 'relação consigo mesma ou fora do curso' AS violation, x.* FROM (SELECT r.from_item_id FROM item_relation r WHERE r.from_item_id = r.to_item_id OR r.to_item_id NOT IN (SELECT id FROM item WHERE course_id = 'd888eee8-9eb3-515b-a2f0-c255d13be4f7')) AS x;
-- violação: item de palavra com mais de um primeiro sentido para o mesmo lema
SELECT 'item de palavra com mais de um primeiro sentido para o mesmo lema' AS violation, x.* FROM (SELECT lexeme_id FROM item WHERE course_id = 'd888eee8-9eb3-515b-a2f0-c255d13be4f7' AND kind = 'word' AND sense_no = 1 GROUP BY lexeme_id HAVING COUNT(*) > 1) AS x;
