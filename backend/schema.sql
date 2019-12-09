CREATE TABLE IF NOT EXISTS source(
    id SERIAL NOT NULL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS congress(
    nbr INTEGER NOT NULL PRIMARY KEY, /* E.g., 115 for 115th */
    start_year INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS house_district(
    id SERIAL NOT NULL PRIMARY KEY,
    state VARCHAR(2) NOT NULL, /* Two-digit state FIPS code */
    district INTEGER NOT NULL,
    congress_nbr INTEGER NOT NULL REFERENCES congress(nbr) ON DELETE RESTRICT,

    CONSTRAINT state_district_congress_unique UNIQUE (state, district, congress_nbr)
);

CREATE TABLE IF NOT EXISTS house_district_pop(
    house_district_id INTEGER NOT NULL REFERENCES house_district(id) ON DELETE CASCADE,
    type VARCHAR(16) NOT NULL, /* 'all', 'adults', 'citizens', or 'cvap' */
    value INTEGER NOT NULL,
    margin_of_error INTEGER NOT NULL,
    source_id INTEGER NOT NULL REFERENCES source(id) ON DELETE RESTRICT,

    CONSTRAINT house_district_pop_unique UNIQUE (house_district_id, type)
);

CREATE TABLE IF NOT EXISTS house_district_turnout(
    house_district_id INTEGER NOT NULL REFERENCES house_district(id) ON DELETE CASCADE,
    year INTEGER NOT NULL,
    num_votes INTEGER NOT NULL,
    source_id INTEGER NOT NULL REFERENCES source(id) ON DELETE RESTRICT,

    CONSTRAINT house_district_turnout_unique UNIQUE (house_district_id, year)
);