-- Merge duplicate professors that differ only by surrounding whitespace.
-- Cause: import-old-data copied names verbatim from the legacy PHP DB, which
-- held both " Name" and "Name" variants; each variant spawned its own
-- class_professors row, so discipline pages listed the professor twice.
--
-- Run inside a transaction against the uspavalia2 database AFTER taking a
-- backup, and BEFORE `./uspavalia migrate` applies the new unique index on
-- class_professors(class_id, professor_id):
--
--   mysqldump uspavalia2 professors class_professors votes comments > backup.sql
--   mysql uspavalia2 < scripts/dedupe_professors.sql

START TRANSACTION;

-- 1. Map every professor to the canonical row of its trimmed-name group
--    (lowest id wins).
CREATE TEMPORARY TABLE prof_map AS
SELECT p.id AS old_id, m.canon_id
FROM professors p
JOIN (
  SELECT TRIM(name) AS tname, MIN(id) AS canon_id
  FROM professors
  GROUP BY TRIM(name)
) m ON TRIM(p.name) = m.tname
WHERE p.id <> m.canon_id;

-- 2. Repoint class_professors at canonical professors.
UPDATE class_professors cp
JOIN prof_map pm ON cp.professor_id = pm.old_id
SET cp.professor_id = pm.canon_id;

-- 3. Collapse class_professors rows that now duplicate (class_id, professor_id);
--    keep the lowest id.
CREATE TEMPORARY TABLE cp_map AS
SELECT cp.id AS old_id, s.keep_id
FROM class_professors cp
JOIN (
  SELECT class_id, professor_id, MIN(id) AS keep_id
  FROM class_professors
  GROUP BY class_id, professor_id
  HAVING COUNT(*) > 1
) s ON cp.class_id = s.class_id AND cp.professor_id = s.professor_id
WHERE cp.id <> s.keep_id;

-- 4. Move votes and comments to the surviving class_professors rows.
UPDATE votes v    JOIN cp_map m ON v.class_professor_id = m.old_id SET v.class_professor_id = m.keep_id;
UPDATE comments c JOIN cp_map m ON c.class_professor_id = m.old_id SET c.class_professor_id = m.keep_id;

-- 5. Delete the collapsed class_professors rows and the orphaned professor
--    variants, then trim every remaining name.
DELETE cp FROM class_professors cp JOIN cp_map m ON cp.id = m.old_id;
DELETE p  FROM professors p       JOIN prof_map m ON p.id = m.old_id;
UPDATE professors SET name = TRIM(name) WHERE name <> TRIM(name);

-- Sanity: both counts must be zero before committing.
SELECT COUNT(*) AS remaining_untrimmed FROM professors WHERE name <> TRIM(name);
SELECT COUNT(*) AS remaining_dup_pairs FROM (
  SELECT 1 FROM class_professors GROUP BY class_id, professor_id HAVING COUNT(*) > 1
) x;

COMMIT;
