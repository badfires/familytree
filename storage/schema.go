package storage

const Schema = `
CREATE TABLE IF NOT EXISTS people (
 id TEXT PRIMARY KEY,
 name TEXT NOT NULL,
 gender TEXT,
 birth_date TEXT,
 birth_place TEXT,
 death_date TEXT,
 burial_place TEXT,
 bio TEXT,
 father_id TEXT,
 mother_id TEXT,
 note TEXT,
 created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
 updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS marriages (
 id TEXT PRIMARY KEY,
 husband_id TEXT,
 wife_id TEXT,
 marriage_date TEXT,
 note TEXT,
 created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
 updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS marriage_children (
 marriage_id TEXT NOT NULL,
 child_id TEXT NOT NULL,
 PRIMARY KEY (marriage_id, child_id)
);

CREATE TABLE IF NOT EXISTS adoptions (
 id TEXT PRIMARY KEY,
 person_id TEXT NOT NULL UNIQUE,
 from_father_id TEXT,
 from_mother_id TEXT,
 to_father_id TEXT,
 to_mother_id TEXT,
 note TEXT,
 created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
 updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 新增：各类型自增序列
CREATE TABLE IF NOT EXISTS id_sequences (
 seq_type TEXT PRIMARY KEY,
 current_value INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_people_name ON people(name);
CREATE INDEX IF NOT EXISTS idx_people_father_id ON people(father_id);
CREATE INDEX IF NOT EXISTS idx_people_mother_id ON people(mother_id);
CREATE INDEX IF NOT EXISTS idx_marriages_husband_id ON marriages(husband_id);
CREATE INDEX IF NOT EXISTS idx_marriages_wife_id ON marriages(wife_id);
CREATE INDEX IF NOT EXISTS idx_marriage_children_marriage_id ON marriage_children(marriage_id);
CREATE INDEX IF NOT EXISTS idx_marriage_children_child_id ON marriage_children(child_id);
CREATE INDEX IF NOT EXISTS idx_adoptions_person_id ON adoptions(person_id);
`