import { PGlite } from "@electric-sql/pglite";
import { readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const R = resolve(dirname(fileURLToPath(import.meta.url)), "../..");   // docs/
const db = new PGlite();
const ver = (await db.query("select version() as v")).rows[0].v;
console.log("Engine:", ver.slice(0, 40));

const run = async (label, file) => {
  try {
    await db.exec(readFileSync(`${R}/${file}`, "utf8"));
    console.log(`OK   ${label}`);
  } catch (e) {
    console.log(`FAIL ${label}: ${e.message}`);
    process.exitCode = 1;
  }
};

await run("content-schema.sql", "fase-2/content-schema.sql");
await run("schema-learning.sql", "fase-3/schema-learning.sql");
await run("sample-seed.sql (1ª)", "fase-6/seed/out/sample-seed.sql");
const c1 = (await db.query("select (select count(*) from item) i, (select count(*) from item_text) t, (select count(*) from lexeme) l, (select count(*) from media) m")).rows[0];
await run("sample-seed.sql (2ª, idempotência)", "fase-6/seed/out/sample-seed.sql");
const c2 = (await db.query("select (select count(*) from item) i, (select count(*) from item_text) t, (select count(*) from lexeme) l, (select count(*) from media) m")).rows[0];
console.log("contagens 1ª:", JSON.stringify(c1), "2ª:", JSON.stringify(c2), JSON.stringify(c1) === JSON.stringify(c2) ? "→ idempotente" : "→ NÃO idempotente");

// conferência (sample-verify.sql): cada statement separado
const verify = readFileSync(`${R}/fase-6/seed/out/sample-verify.sql`, "utf8").split(String.fromCharCode(13)).join("");
let bad = 0;
for (const stmt of verify.split(";\n")) {
  const sql = stmt.split("\n").filter((l) => !l.trim().startsWith("--")).join("\n").trim();
  if (!sql.toUpperCase().startsWith("SELECT")) continue;
  const rows = (await db.query(sql)).rows;
  if (sql.includes("AS violation")) { if (rows.length) { bad++; console.log("VIOLAÇÃO", rows.slice(0, 3)); } }
  else { const r = rows[0]; console.log(`${r.status === "ok" ? "ok  " : "FALHA"} ${r.check_name} ${r.found}/${r.expected}`); if (r.status !== "ok") bad++; }
}
console.log(bad ? `conferência: ${bad} problema(s)` : "conferência: sem violações");

// Semântica do schema de aprendizado
const t = async (name, fn, expectError) => {
  try { await fn(); console.log(expectError ? `FAIL ${name}: deveria falhar` : `OK   ${name}`); if (expectError) process.exitCode = 1; }
  catch (e) { console.log(expectError ? `OK   ${name} (erro esperado: ${e.message.slice(0, 60)})` : `FAIL ${name}: ${e.message}`); if (!expectError) process.exitCode = 1; }
};
const uid = "00000000-0000-4000-8000-000000000001";
const itemId = (await db.query("select id from item where natural_key='word|cat|noun|1'")).rows[0].id;
await db.exec(`insert into app_user(id, display_name) values ('${uid}','teste')`);
await db.exec(`insert into user_settings(user_id) values ('${uid}')`);
const cardId = (await db.query(`insert into card(user_id,item_id,direction,scheduler) values ('${uid}','${itemId}','receptive','simple_v1') returning id`)).rows[0].id;
const rev = (id) => `insert into review_log(id,user_id,card_id,exercise_type,source,shown_at,answered_at,tz,day_cutoff_hour,local_date,is_correct,grade,state_before,state_after,scheduler,scheduler_version)
  values ('${id}','${uid}','${cardId}','choice_en_pt','session',now(),now(),'America/Sao_Paulo',4,current_date,true,3,'new','learning','simple_v1','1') on conflict (id) do nothing`;
const rid = "11111111-1111-4111-8111-111111111111";
await t("review_log: inserir", () => db.exec(rev(rid)), false);
await t("review_log: reenvio do mesmo UUID (ON CONFLICT DO NOTHING)", () => db.exec(rev(rid)), false);
console.log("   linhas no log:", (await db.query("select count(*) c from review_log")).rows[0].c, "(esperado 1)");
await t("review_log: UPDATE bloqueado pelo gatilho", () => db.exec(`update review_log set grade=4 where id='${rid}'`), true);
await t("review_log: DELETE bloqueado pelo gatilho", () => db.exec(`delete from review_log where id='${rid}'`), true);
await t("card: UNIQUE (user,item,direction)", () => db.exec(`insert into card(user_id,item_id,direction,scheduler) values ('${uid}','${itemId}','receptive','simple_v1')`), true);
await t("card: direction inválida (CHECK)", () => db.exec(`insert into card(user_id,item_id,direction,scheduler) values ('${uid}','${itemId}','xx','simple_v1')`), true);
await t("view user_lexeme_known consulta", () => db.query("select * from user_lexeme_known"), false);
await t("outbox: UNIQUE(topic,event_key)", async () => {
  await db.exec(`insert into outbox(topic,event_key,payload) values ('push.send','${rid}','{}')`);
  await db.exec(`insert into outbox(topic,event_key,payload) values ('push.send','${rid}','{}')`);
}, true);
await t("SKIP LOCKED na outbox", () => db.exec("select id from outbox where processed_at is null and available_at <= now() order by id for update skip locked limit 10"), false);
console.log(process.exitCode ? "\nRESULTADO: houve falhas" : "\nRESULTADO: tudo ok no PostgreSQL (PGlite)");
