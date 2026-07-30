-- Merge duplicate units, disciplines, and professors left by the legacy
-- import and the fetch-courses NOME-column bug.
--
--  * professors: 2,800 whitespace-variant name groups (import copied names
--    verbatim; leading spaces are significant in MySQL equality)
--  * disciplines: 17 codes with two rows each (same cause); code is the
--    discipline identity
--  * units: ~6,430 rows for ~80 real names (fetch-courses queried the old
--    PHP column NOME, always "missed", and created a fresh unit per course
--    per run)
--
-- Run inside a transaction against the uspavalia2 database AFTER taking a
-- backup, and BEFORE `./uspavalia migrate` applies the unique indexes on
-- professors(name), disciplines(code), class_professors(class_id,
-- professor_id):
--
--   mysqldump uspavalia2 units professors disciplines class_professors \
--     class_offerings votes comments > backup.sql
--   mysql uspavalia2 < scripts/dedupe_professors.sql

START TRANSACTION;

-- ========== 1. Units: merge by trimmed name (lowest id wins) ==========
CREATE TEMPORARY TABLE unit_map AS
SELECT u.id AS old_id, m.canon_id
FROM units u
JOIN (
  SELECT TRIM(name) AS tname, MIN(id) AS canon_id
  FROM units GROUP BY TRIM(name)
) m ON TRIM(u.name) = m.tname
WHERE u.id <> m.canon_id;

UPDATE professors p  JOIN unit_map um ON p.unit_id = um.old_id SET p.unit_id = um.canon_id;
UPDATE disciplines d JOIN unit_map um ON d.unit_id = um.old_id SET d.unit_id = um.canon_id;
UPDATE courses c     JOIN unit_map um ON c.unit_id = um.old_id SET c.unit_id = um.canon_id;
DELETE u FROM units u JOIN unit_map um ON u.id = um.old_id;
UPDATE units SET name = TRIM(name) WHERE name <> TRIM(name);

-- ========== 2. Disciplines: merge by code (lowest id wins) ==========
CREATE TEMPORARY TABLE disc_map AS
SELECT d.id AS old_id, m.canon_id
FROM disciplines d
JOIN (
  SELECT code, MIN(id) AS canon_id
  FROM disciplines GROUP BY code
) m ON d.code = m.code
WHERE d.id <> m.canon_id;

UPDATE class_professors cp JOIN disc_map dm ON cp.class_id = dm.old_id
SET cp.class_id = dm.canon_id;
UPDATE class_offerings co JOIN disc_map dm ON co.discipline_id = dm.old_id
SET co.discipline_id = dm.canon_id;

-- Collapse class offerings that now duplicate (discipline_id, code);
-- keep the most recently updated row.
CREATE TEMPORARY TABLE co_map AS
SELECT co.id AS old_id, s.keep_id
FROM class_offerings co
JOIN (
  SELECT discipline_id, code,
         SUBSTRING_INDEX(GROUP_CONCAT(id ORDER BY updated_at DESC, id DESC), ',', 1) AS keep_id
  FROM class_offerings
  GROUP BY discipline_id, code HAVING COUNT(*) > 1
) s ON co.discipline_id = s.discipline_id AND co.code = s.code
WHERE co.id <> s.keep_id;

DELETE co FROM class_offerings co JOIN co_map m ON co.id = m.old_id;
DELETE d  FROM disciplines d      JOIN disc_map m ON d.id = m.old_id;
UPDATE disciplines SET name = TRIM(name) WHERE name <> TRIM(name);

-- ========== 3. Professors: merge by trimmed name (lowest id wins) ==========
CREATE TEMPORARY TABLE prof_map AS
SELECT p.id AS old_id, m.canon_id
FROM professors p
JOIN (
  SELECT TRIM(name) AS tname, MIN(id) AS canon_id
  FROM professors GROUP BY TRIM(name)
) m ON TRIM(p.name) = m.tname
WHERE p.id <> m.canon_id;

UPDATE class_professors cp JOIN prof_map pm ON cp.professor_id = pm.old_id
SET cp.professor_id = pm.canon_id;

-- ========== 4. Collapse class_professors duplicated by the merges ==========
CREATE TEMPORARY TABLE cp_map AS
SELECT cp.id AS old_id, s.keep_id
FROM class_professors cp
JOIN (
  SELECT class_id, professor_id, MIN(id) AS keep_id
  FROM class_professors
  GROUP BY class_id, professor_id HAVING COUNT(*) > 1
) s ON cp.class_id = s.class_id AND cp.professor_id = s.professor_id
WHERE cp.id <> s.keep_id;

UPDATE votes v    JOIN cp_map m ON v.class_professor_id = m.old_id SET v.class_professor_id = m.keep_id;
UPDATE comments c JOIN cp_map m ON c.class_professor_id = m.old_id SET c.class_professor_id = m.keep_id;

DELETE cp FROM class_professors cp JOIN cp_map m ON cp.id = m.old_id;
DELETE p  FROM professors p        JOIN prof_map m ON p.id = m.old_id;
UPDATE professors SET name = TRIM(name) WHERE name <> TRIM(name);

-- ========== Sanity: every count below must be zero before COMMIT ==========
SELECT COUNT(*) AS untrimmed_professors  FROM professors  WHERE name <> TRIM(name);
SELECT COUNT(*) AS untrimmed_disciplines FROM disciplines WHERE name <> TRIM(name);
SELECT COUNT(*) AS untrimmed_units       FROM units       WHERE name <> TRIM(name);
SELECT COUNT(*) AS dup_professor_names FROM (SELECT 1 FROM professors  GROUP BY TRIM(name) HAVING COUNT(*) > 1) x;
SELECT COUNT(*) AS dup_discipline_codes FROM (SELECT 1 FROM disciplines GROUP BY code HAVING COUNT(*) > 1) x;
SELECT COUNT(*) AS dup_unit_names FROM (SELECT 1 FROM units GROUP BY TRIM(name) HAVING COUNT(*) > 1) x;
SELECT COUNT(*) AS dup_cp_pairs FROM (SELECT 1 FROM class_professors GROUP BY class_id, professor_id HAVING COUNT(*) > 1) x;
SELECT COUNT(*) AS dup_offerings FROM (SELECT 1 FROM class_offerings GROUP BY discipline_id, code HAVING COUNT(*) > 1) x;

COMMIT;
