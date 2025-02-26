CREATE TABLE IF NOT EXISTS inspec_profiles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(255) NOT NULL,
    description TEXT,
    stars INTEGER,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
