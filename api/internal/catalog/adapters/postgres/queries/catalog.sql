-- name: ListUnits :many
SELECT u.id, u.position, u.theme_key,
       (SELECT count(*) FROM item i
         WHERE i.unit_id = u.id AND (i.status = 'published' OR @include_drafts::boolean))::int AS item_count
FROM unit u
JOIN course c ON c.id = u.course_id AND (c.status = 'published' OR @include_drafts::boolean)
ORDER BY c.slug, u.position;

-- name: ListItems :many
SELECT i.id, i.kind, i.cefr, i.unit_id, i.status,
       term.body AS term,
       COALESCE(gloss.body, gloss_default.body, '')::text AS gloss,
       COALESCE((SELECT m.storage_key FROM item_media im JOIN media m ON m.id = im.media_id
                  WHERE im.item_id = i.id AND im.role = 'context' AND (m.status = 'approved' OR @include_drafts::boolean)
                  ORDER BY im.position LIMIT 1), '')::text AS image_key,
       (count(*) OVER ())::int AS total
FROM item i
JOIN course c ON c.id = i.course_id AND (c.status = 'published' OR @include_drafts::boolean)
JOIN item_text term ON term.item_id = i.id AND term.role = 'term' AND term.position = 0 AND term.locale = c.studied_locale
LEFT JOIN item_text gloss ON gloss.item_id = i.id AND gloss.role = 'gloss' AND gloss.position = 0 AND gloss.locale = @locale::text
LEFT JOIN item_text gloss_default ON gloss_default.item_id = i.id AND gloss_default.role = 'gloss' AND gloss_default.position = 0 AND gloss_default.locale = c.help_locale
WHERE (i.status = 'published' OR @include_drafts::boolean)
  AND i.kind IN ('word', 'expression', 'phrasal_verb', 'sentence')
  AND (sqlc.narg('id')::uuid IS NULL OR i.id = sqlc.narg('id')::uuid)
  AND (sqlc.narg('unit')::uuid IS NULL OR i.unit_id = sqlc.narg('unit')::uuid)
  AND (sqlc.narg('cefr')::text IS NULL OR i.cefr = sqlc.narg('cefr')::text)
  AND (sqlc.narg('kind')::text IS NULL OR i.kind = sqlc.narg('kind')::text)
  AND (sqlc.narg('pattern')::text IS NULL
       OR term.body ILIKE sqlc.narg('pattern')::text
       OR COALESCE(gloss.body, gloss_default.body, '') ILIKE sqlc.narg('pattern')::text)
ORDER BY
  CASE WHEN @sort::text = 'term,asc' THEN lower(term.body) END ASC,
  CASE WHEN @sort::text = 'term,desc' THEN lower(term.body) END DESC,
  CASE WHEN @sort::text = 'cefr,asc' THEN i.cefr END ASC,
  CASE WHEN @sort::text = 'cefr,desc' THEN i.cefr END DESC,
  lower(term.body), i.id
LIMIT @row_limit::int OFFSET @row_offset::int;

-- name: ListItemExamples :many
SELECT DISTINCT ON (ex.position) ex.body AS text, COALESCE(tr.body, '')::text AS translation
FROM item_text ex
LEFT JOIN item_text tr ON tr.item_id = ex.item_id AND tr.role = 'example_gloss' AND tr.position = ex.position
                      AND tr.locale IN (@locale::text, @fallback_locale::text)
WHERE ex.item_id = @item_id AND ex.role = 'example'
ORDER BY ex.position, (tr.locale = @locale::text) DESC NULLS LAST;

-- name: ListItemImages :many
SELECT m.storage_key, m.attribution_text
FROM item_media im
JOIN media m ON m.id = im.media_id
WHERE im.item_id = @item_id AND im.role = 'context' AND (m.status = 'approved' OR @include_drafts::boolean)
ORDER BY im.position;

-- name: ListRelatedItems :many
SELECT o.id AS item_id, term.body AS term, COALESCE(gloss.body, '')::text AS gloss, r.type
FROM (SELECT ir.to_item_id AS other_id, ir.type, ir.weight FROM item_relation ir WHERE ir.from_item_id = @item_id
      UNION ALL
      SELECT ir.from_item_id AS other_id, ir.type, ir.weight FROM item_relation ir WHERE ir.to_item_id = @item_id) r
JOIN item o ON o.id = r.other_id AND (o.status = 'published' OR @include_drafts::boolean)
JOIN course c ON c.id = o.course_id
JOIN item_text term ON term.item_id = o.id AND term.role = 'term' AND term.position = 0 AND term.locale = c.studied_locale
LEFT JOIN item_text gloss ON gloss.item_id = o.id AND gloss.role = 'gloss' AND gloss.position = 0 AND gloss.locale = @locale::text
ORDER BY r.weight DESC, lower(term.body);

-- name: ListCredits :many
SELECT code, name, COALESCE(url, '')::text AS url, license_spdx, attribution_notice
FROM source
ORDER BY name;
