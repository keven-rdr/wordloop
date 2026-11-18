-- +goose Up
-- LGPD, eliminacao (I-05): o review_log e imutavel por gatilho, e a exclusao da conta precisa apagar o log da pessoa.
-- A excecao e estreita: so um DELETE, so das linhas de quem a transacao declara (SET LOCAL wordloop.erase_user = <id>).
-- UPDATE continua proibido, e qualquer outro DELETE tambem. Sem travar a tabela (nada de DISABLE TRIGGER).
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION review_log_immutable() RETURNS trigger LANGUAGE plpgsql AS
$$
BEGIN
    IF TG_OP = 'DELETE' AND OLD.user_id::text = current_setting('wordloop.erase_user', true) THEN
        RETURN OLD;
    END IF;
    RAISE EXCEPTION 'review_log é append-only';
END
$$;
-- +goose StatementEnd

-- Prestacao de contas: so registra QUE houve exclusao, sem id nem e-mail (a pessoa nao pode ficar identificavel).
CREATE TABLE account_erasure (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    erased_at    timestamptz NOT NULL DEFAULT now(),
    review_count int         NOT NULL
);

-- +goose Down
DROP TABLE account_erasure;
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION review_log_immutable() RETURNS trigger LANGUAGE plpgsql AS
$$ BEGIN RAISE EXCEPTION 'review_log é append-only'; END $$;
-- +goose StatementEnd
