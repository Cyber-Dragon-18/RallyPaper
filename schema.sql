CREATE TABLE IF NOT EXISTS teams (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    color TEXT DEFAULT '#1a73e8',
    stop1 INTEGER DEFAULT 0,
    stop2 INTEGER DEFAULT 0,
    stop3 INTEGER DEFAULT 0,
    stop4 INTEGER DEFAULT 0,
    stop5 INTEGER DEFAULT 0,
    stop6 INTEGER DEFAULT 0,
    total_points INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS stages (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    location TEXT,
    clue TEXT,
    max_points INTEGER DEFAULT 100,
    stop_number INTEGER UNIQUE NOT NULL  -- 1 to 6
);