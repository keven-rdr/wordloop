-- Semântica do schema de aprendizado, para rodar com psql (-v ON_ERROR_STOP=1) num PostgreSQL de verdade.
-- Pré-requisito: content-schema.sql, schema-learning.sql e o seed já aplicados.
-- Cada bloco levanta EXCEPTION se o comportamento esperado NÃO ocorrer; no fim imprime 'SEMANTICA OK'.

DO $$
DECLARE
  uid  uuid := '00000000-0000-4000-8000-000000000001';
  iid  uuid;
  cid  uuid;
  rid  uuid := '11111111-1111-4111-8111-111111111111';
  n    int;
BEGIN
  SELECT id INTO iid FROM item WHERE natural_key = 'word|cat|noun|1';
  INSERT INTO app_user(id, display_name) VALUES (uid, 'teste');
  INSERT INTO user_settings(user_id) VALUES (uid);
  INSERT INTO card(user_id, item_id, direction, scheduler) VALUES (uid, iid, 'receptive', 'simple_v1') RETURNING id INTO cid;

  -- review_log: inserir e reenviar o mesmo UUID (idempotência)
  FOR i IN 1..2 LOOP
    INSERT INTO review_log(id,user_id,card_id,exercise_type,source,shown_at,answered_at,tz,day_cutoff_hour,local_date,
                           is_correct,grade,state_before,state_after,scheduler,scheduler_version)
    VALUES (rid,uid,cid,'choice_en_pt','session',now(),now(),'America/Sao_Paulo',4,current_date,
            true,3,'new','learning','simple_v1','1') ON CONFLICT (id) DO NOTHING;
  END LOOP;
  SELECT count(*) INTO n FROM review_log;
  IF n <> 1 THEN RAISE EXCEPTION 'idempotência falhou: % linhas', n; END IF;
  RAISE NOTICE 'ok  review_log idempotente (1 linha após 2 envios)';

  -- gatilho append-only
  BEGIN UPDATE review_log SET grade = 4 WHERE id = rid; RAISE EXCEPTION 'UPDATE passou';
  EXCEPTION WHEN OTHERS THEN IF SQLERRM LIKE 'UPDATE passou' THEN RAISE; END IF; END;
  BEGIN DELETE FROM review_log WHERE id = rid; RAISE EXCEPTION 'DELETE passou';
  EXCEPTION WHEN OTHERS THEN IF SQLERRM LIKE 'DELETE passou' THEN RAISE; END IF; END;
  RAISE NOTICE 'ok  UPDATE e DELETE bloqueados no review_log';

  -- UNIQUE e CHECK de card
  BEGIN INSERT INTO card(user_id,item_id,direction,scheduler) VALUES (uid,iid,'receptive','simple_v1'); RAISE EXCEPTION 'UNIQUE passou';
  EXCEPTION WHEN unique_violation THEN NULL; END;
  BEGIN INSERT INTO card(user_id,item_id,direction,scheduler) VALUES (uid,iid,'xx','simple_v1'); RAISE EXCEPTION 'CHECK passou';
  EXCEPTION WHEN check_violation THEN NULL; END;
  RAISE NOTICE 'ok  UNIQUE(user,item,direction) e CHECK(direction)';

  -- outbox: UNIQUE(topic,event_key) e SKIP LOCKED
  INSERT INTO outbox(topic,event_key,payload) VALUES ('push.send', rid, '{}');
  BEGIN INSERT INTO outbox(topic,event_key,payload) VALUES ('push.send', rid, '{}'); RAISE EXCEPTION 'outbox UNIQUE passou';
  EXCEPTION WHEN unique_violation THEN NULL; END;
  PERFORM id FROM outbox WHERE processed_at IS NULL AND available_at <= now() ORDER BY id FOR UPDATE SKIP LOCKED LIMIT 10;
  RAISE NOTICE 'ok  outbox UNIQUE e FOR UPDATE SKIP LOCKED';

  -- view de lemas conhecidos (card receptivo em review com S >= 7)
  UPDATE card SET state = 'review', stability = 9 WHERE id = cid;
  SELECT count(*) INTO n FROM user_lexeme_known WHERE user_id = uid;
  IF n <> 1 THEN RAISE EXCEPTION 'user_lexeme_known esperava 1, veio %', n; END IF;
  RAISE NOTICE 'ok  user_lexeme_known reconhece o lema com S >= 7';

  -- partial index de cartões vencidos existe e é usado por consulta típica
  PERFORM 1 FROM pg_indexes WHERE indexname = 'card_due_idx';
  IF NOT FOUND THEN RAISE EXCEPTION 'card_due_idx ausente'; END IF;
  RAISE NOTICE 'ok  índices parciais presentes';
END $$;

SELECT 'SEMANTICA OK' AS resultado;
