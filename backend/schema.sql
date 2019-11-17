CREATE TABLE IF NOT EXISTS source(
    id SERIAL NOT NULL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS house_district(
    id SERIAL NOT NULL PRIMARY KEY,
    state VARCHAR(2) NOT NULL, /* Two-digit state FIPS code */
    district VARCHAR(2) NOT NULL, /* E.g., '02' for 2nd */
    congress INTEGER NOT NULL,   /* E.g., 115 for 115th */
    pop INTEGER,
    pop_moe INTEGER, /* Margin of error */
    cvap INTEGER,
    cvap_moe INTEGER, /* Margin of error */
    pop_source INTEGER REFERENCES source(id) ON DELETE CASCADE,
    cvap_source INTEGER REFERENCES source(id) ON DELETE CASCADE,

    CONSTRAINT state_district_congress_unique UNIQUE (state, district, congress)
);

CREATE TABLE IF NOT EXISTS house_district_turnout(
    id SERIAL NOT NULL PRIMARY KEY,
    house_district_id INTEGER REFERENCES house_district(id) ON DELETE CASCADE,
    num_votes INTEGER NOT NULL,
    month INTEGER NOT NULL,
    year INTEGER NOT NULL,
    source INTEGER REFERENCES source(id) ON DELETE RESTRICT
);