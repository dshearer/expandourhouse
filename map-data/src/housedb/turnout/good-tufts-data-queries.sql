CREATE VIEW IF NOT EXISTS states(usps, name) AS VALUES
('AL', 'Alabama'),
('AK', 'Alaska'),
('AZ', 'Arizona'),
('AR', 'Arkansas'),
('CA', 'California'),
('CO', 'Colorado'),
('CT', 'Connecticut'),
('DE', 'Delaware'),
('DC', 'District of Columbia'),
('FL', 'Florida'),
('GA', 'Georgia'),
('HI', 'Hawaii'),
('ID', 'Idaho'),
('IL', 'Illinois'),
('IN', 'Indiana'),
('IA', 'Iowa'),
('KS', 'Kansas'),
('KY', 'Kentucky'),
('LA', 'Louisiana'),
('ME', 'Maine'),
('MD', 'Maryland'),
('MA', 'Massachusetts'),
('MI', 'Michigan'),
('MN', 'Minnesota'),
('MS', 'Mississippi'),
('MO', 'Missouri'),
('MT', 'Montana'),
('NE', 'Nebraska'),
('NV', 'Nevada'),
('NH', 'New Hampshire'),
('NJ', 'New Jersey'),
('NM', 'New Mexico'),
('NY', 'New York'),
('NC', 'North Carolina'),
('ND', 'North Dakota'),
('OH', 'Ohio'),
('OK', 'Oklahoma'),
('OR', 'Oregon'),
('PA', 'Pennsylvania'),
('RI', 'Rhode Island'),
('SC', 'South Carolina'),
('SD', 'South Dakota'),
('TN', 'Tennessee'),
('TX', 'Texas'),
('UT', 'Utah'),
('VT', 'Vermont'),
('VA', 'Virginia'),
('WA', 'Washington'),
('WV', 'West Virginia'),
('WI', 'Wisconsin'),
('WY', 'Wyoming'),
('GU', 'Guam'),
('MP', 'Northern Mariana Islands'),
('AS', 'American Samoa'),
('PR', 'Puerto Rico'),
('VI', 'Virgin Islands');

CREATE VIEW IF NOT EXISTS numbers(name, value) AS VALUES 
('One', 1),
('Two', 2),
('Three', 3),
('Four', 4),
('Five', 5),
('Six', 6),
('Seven', 7),
('Eight', 8),
('Nine', 9),
('Ten', 10),
('Eleven', 11),
('Twelve', 12),
('Thirteen', 13),
('Fourteen', 14),
('Fifteen', 15),
('Sixteen', 16),
('Seventeen', 17),
('Eighteen', 18),
('Nineteen', 19),
('Twenty', 20),
('Twnety One', 21),
('Twenty Two', 22),
('Twenty Three', 23),
('Twenty Four', 24),
('Twenty Five', 25),
('Twenty Six', 26),
('Twenty Seven', 27),
('Twenty Eight', 28),
('Twenty Nine', 29),
('Thirty', 30),
('Twenty-One', 21);

CREATE VIEW IF NOT EXISTS clean_raw_tufts AS
SELECT id,
	SUBSTR(date, 1, 4) AS year,
	CASE SUBSTR(date, 1, 4) % 2 = 0 
		WHEN TRUE THEN (SUBSTR(date, 1, 4)+1-1789)/2+1
		ELSE (SUBSTR(date, 1, 4)-1789)/2+1
	END AS congress_nbr,
	n.value AS district,
	states.usps AS state,
	vote,
	name_id
FROM raw_tufts JOIN numbers n ON (raw_tufts.district = n.name)
JOIN states ON (raw_tufts.state = states.name)
WHERE type = 'General'
AND territory = '' AND city = '' AND county = '' AND town = '' AND township = '' AND ward = '' AND parish = '' AND pop_place = '' AND hundred = '' AND borough = ''
AND district <> '' AND iteration = 'First Ballot';

CREATE VIEW IF NOT EXISTS multi_names AS 
SELECT id, name_id, COUNT(*) AS cnt
FROM clean_raw_tufts
GROUP BY id, name_id;

CREATE VIEW IF NOT EXISTS bad_ids AS 
SELECT id FROM clean_raw_tufts
WHERE (vote = '' OR vote = 'NULL' OR vote = 'null') OR id IN (SELECT id FROM multi_names WHERE cnt > 1)
OR id LIKE '%essex%' OR id LIKE '%middlesex%' OR id LIKE '%suffolk%' OR id LIKE '%lincoln%' OR id LIKE '%cumberland%'
OR id LIKE '%york%' OR id LIKE '%shire%' OR id LIKE '%ster%' OR id LIKE '%mouth%' OR id LIKE '%bristol%'
OR id LIKE '%spring%' OR id LIKE 'pa.congress.16.1824' OR id LIKE 'pa.congress.17.1824';

CREATE VIEW IF NOT EXISTS good_tufts_data AS SELECT state,
	district, 
	year,
	CAST(congress_nbr AS INT) AS congress_nbr,
	SUM(vote) AS vote
FROM clean_raw_tufts
WHERE id NOT IN (SELECT * FROM bad_ids)
GROUP BY id
ORDER BY state, district, congress_nbr;