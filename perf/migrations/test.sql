-- This is a useful file for playing around with SQL queries against a database
-- populated with Perf data. You can use this file by running:
--
--    sqlite3 test.db < test.sql
--
-- Where test.db is an empty sqlite3 database. You should be able to run this
-- file against the same sqlite3 database more than once w/o error.
CREATE TABLE IF NOT EXISTS TraceIDs  (
    trace_id INTEGER PRIMARY KEY,
    trace_name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS Postings  (
    tile_number INTEGER,
    key_value text NOT NULL,
    trace_id INTEGER,
    PRIMARY KEY (tile_number, key_value, trace_id)
);

CREATE TABLE IF NOT EXISTS SourceFiles (
    source_file_id INTEGER PRIMARY KEY,
    source_file TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS TraceValues (
    trace_id INTEGER,
    commit_number INTEGER,
    val REAL,
    source_file_id INTEGER,
    PRIMARY KEY (trace_id, commit_number)
);

CREATE TABLE IF NOT EXISTS Commits (
  commit_number INTEGER PRIMARY KEY,
  git_hash TEXT UNIQUE NOT NULL,
  commit_time INTEGER, -- And not author_time.
  author TEXT,  -- Name <email>
  subject TEXT
);

INSERT OR IGNORE INTO Commits (commit_number, git_hash, commit_time, author, subject)
VALUES
  (0, "586101c79b0490b50623e76c71a5fd67d8d92b08", 1158764756, "unknown@example.com", "initial directory structure"),
  (1, "0f87cd842dd46205d5252c35da6d2c869f3d2e98", 1158767262, "unknown@example.com", "initial code checkin"),
  (2, "48ede9b432a3c3d62835a1400a9ed347b4a93024", 1163013888, "unknown@example.org", "Add LICENSE");

-- Get most recent git hash.
SELECT git_hash FROM Commits
ORDER BY commit_number DESC
LIMIT 1;

-- Get commit_number from git hash.
SELECT commit_number FROM Commits
WHERE git_hash='0f87cd842dd46205d5252c35da6d2c869f3d2e98';

-- Get commit_number from time.
SELECT commit_number FROM Commits
WHERE commit_time <= 1163013888
ORDER BY commit_number DESC
LIMIT 1;

INSERT OR IGNORE INTO SourceFiles (source_file)
VALUES
  ("gs://perf-bucket/2020/02/08/11/testdata.json"),
  ("gs://perf-bucket/2020/02/08/12/testdata.json"),
  ("gs://perf-bucket/2020/02/08/13/testdata.json"),
  ("gs://perf-bucket/2020/02/08/14/testdata.json");

INSERT OR IGNORE INTO TraceIDs (trace_name)
VALUES
 (",arch=x86,config=8888,"),
 (",arch=x86,config=565,"),
 (",arch=arm,config=8888,"),
 (",arch=arm,config=565,");

 SELECT trace_id, trace_name FROM TraceIDs;

 INSERT OR REPLACE INTO TraceValues (trace_id, commit_number, val, source_file_id)
 VALUES
   (1, 1,   1.2, 1),
   (1, 2,   1.3, 2),
   (1, 3,   1.4, 3),
   (1, 256, 1.1, 4),
   (2, 1,   2.2, 1),
   (2, 2,   2.3, 2),
   (2, 3,   2.4, 3),
   (2, 256, 2.1, 4);

INSERT OR REPLACE INTO Postings (tile_number, key_value, trace_id)
VALUES
   (2, "config=565", 4),
   (0, "arch=x86", 1),
   (0, "arch=x86", 2),
   (0, "arch=arm", 3),
   (0, "arch=arm", 4),
   (0, "config=8888", 1),
   (0, "config=8888", 3),
   (0, "config=565", 2),
   (0, "config=565", 4);

-- All trace_ids that match a particular key=value.
SELECT tile_number, key_value, trace_id FROM Postings
WHERE tile_number=0 AND key_value="arch=x86"
ORDER BY trace_id;

-- Retrieve matching values. Note that sqlite querys are limited to 1MB,
-- so we might need to break up the trace_ids if the query is too long.
SELECT trace_id, commit_number, val FROM TraceValues
WHERE commit_number>=0 AND commit_number<255 AND trace_id IN (1,2);

-- Build traces using a JOIN.
SELECT TraceIDs.trace_name, TraceValues.commit_number, TraceValues.val FROM TraceIDs
INNER JOIN TraceValues ON TraceValues.trace_id = TraceIDs.trace_id
WHERE TraceIDs.trace_name=",arch=x86,config=8888," OR TraceIDs.trace_name=",arch=x86,config=565,";

-- Retrieve source file.
SELECT source_file_id, source_file from SourceFiles
WHERE source_file_id=1;

SELECT DISTINCT key_value FROM Postings
WHERE tile_number=0;

-- Most recent tile.
SELECT tile_number FROM Postings ORDER BY tile_number DESC LIMIT 1;

-- Count indices for Tile.
SELECT COUNT(*) FROM Postings WHERE tile_number=0;

-- GetSource by trace name.
SELECT SourceFiles.source_file FROM TraceIDs
INNER JOIN TraceValues ON TraceValues.trace_id = TraceIDs.trace_id
INNER JOIN SourceFiles ON SourceFiles.source_file_id = TraceValues.source_file_id
WHERE TraceIDs.trace_name=",arch=x86,config=8888," AND TraceValues.commit_number=256;

-- Fully query traces from tile based on query plan.
SELECT TraceIDs.trace_name, TraceValues.commit_number, TraceValues.val FROM TraceIDs
INNER JOIN TraceValues ON TraceValues.trace_id = TraceIDs.trace_id
WHERE
  TraceValues.trace_id IN (
    SELECT trace_id FROM Postings WHERE key_value IN ("arch=x86", "arch=arm")
    AND tile_number=0
  )
  AND TraceValues.trace_id IN (
    SELECT trace_id FROM Postings WHERE key_value IN ("config=8888")
    AND tile_number=0
  );

-- Count the number traces that are in a single tile.
SELECT COUNT(DISTINCT trace_id) FROM TraceValues
WHERE
  commit_number > 0
  AND commit_number < 8;
