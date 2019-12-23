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
    district INTEGER NOT NULL, /* 0 = at-large */
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
    num_votes INTEGER NOT NULL,
    source_id INTEGER NOT NULL REFERENCES source(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS representative_term(
    id SERIAL NOT NULL PRIMARY KEY,
    house_district_id INTEGER REFERENCES house_district(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    bioguide_id VARCHAR(16) NOT NULL,
    state VARCHAR(2) NOT NULL, /* Two-digit state FIPS code */
    congress_nbr INTEGER NOT NULL REFERENCES congress(nbr) ON DELETE RESTRICT,

    CONSTRAINT rep_term_dates CHECK (start_date <= end_date)
);

CREATE VIEW state_with_overlapping_terms AS
SELECT DISTINCT t1.state, t1.congress_nbr
/* Collect all pairs of rep terms for the same state, district, and congress */
FROM representative_term t1 JOIN representative_term t2
    ON (t1.house_district_id = t2.house_district_id AND t1.id != t2.id
        AND t1.house_district_id IS NOT NULL)
/* Find ones that overlap in time */
WHERE t1.start_date <= t2.start_date AND t1.end_date > t2.start_date;

CREATE VIEW state_with_unknown_district AS
SELECT DISTINCT state, congress_nbr
FROM representative_term
WHERE house_district_id IS NULL;

CREATE VIEW state_with_atlarge_district AS
SELECT DISTINCT term.state, term.congress_nbr
FROM representative_term term JOIN house_district dist
    ON (term.house_district_id = dist.id)
WHERE dist.district = 0; /* 0 = at-large */

CREATE VIEW state_with_nonatlarge_district AS
SELECT DISTINCT term.state, term.congress_nbr
FROM representative_term term JOIN house_district dist
    ON (term.house_district_id = dist.id)
WHERE dist.district > 0;

CREATE VIEW state_with_atlarge_and_nonatlarge_districts AS
(SELECT state, congress_nbr FROM state_with_atlarge_district)
INTERSECT
(SELECT state, congress_nbr FROM state_with_nonatlarge_district);

CREATE MATERIALIZED VIEW irregular_state AS
(SELECT state, congress_nbr FROM state_with_atlarge_and_nonatlarge_districts)
UNION
(SELECT state, congress_nbr FROM state_with_unknown_district)
UNION
(SELECT state, congress_nbr FROM state_with_overlapping_terms);;