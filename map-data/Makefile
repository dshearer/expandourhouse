CENSUS_DATA_FILE = tl_2019_us_cd116
CENSUS_DATA_FILE_URL = ftp://ftp2.census.gov/geo/tiger/TIGER2019/CD/${CENSUS_DATA_FILE}.zip

.PHONY: main
main: data/${CENSUS_DATA_FILE}.mbtiles

data/${CENSUS_DATA_FILE}.zip:
	mkdir -p data
	wget -P data "${CENSUS_DATA_FILE_URL}"

data/${CENSUS_DATA_FILE}.shp: data/${CENSUS_DATA_FILE}.zip
	unzip "$<" -d data
	chmod -x data/*
	touch data/*

data/${CENSUS_DATA_FILE}.geojson: data/${CENSUS_DATA_FILE}.shp
	ogr2ogr -f GeoJSON -t_srs crs:84 "$@" "$<"

data/${CENSUS_DATA_FILE}-processed.geojson: data/${CENSUS_DATA_FILE}.geojson process.js
	npm i
	node process.js "$<" "$@"

data/${CENSUS_DATA_FILE}.mbtiles: data/${CENSUS_DATA_FILE}-processed.geojson
	tippecanoe -o "$@" -f -z 12 -Z 0 -B 0 -pS -pp -l districts -n "US Congressional Districts" "$<"

.PHONY: upload
upload: data/${CENSUS_DATA_FILE}.mbtiles
	node upload.js "$<" dshearer

.PHONY: clean
clean:
	rm data/*.shp data/*.geojson data/*.cpg data/*.dbf data/*.prj data/*.xml data/*.shx data/*.mbtiles

.PHONY: deepclean
deepclean: clean
	rm -rf node_modules