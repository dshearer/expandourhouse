CENSUS_DATA_FILE = tl_2019_us_cd116
CENSUS_DATA_FILE_URL = ftp://ftp2.census.gov/geo/tiger/TIGER2019/CD/${CENSUS_DATA_FILE}.zip

.PHONY: main
main: data/${CENSUS_DATA_FILE}.mbtiles

data/${CENSUS_DATA_FILE}.zip:
	# Download district boundaries from Census
	mkdir -p data
	wget -P data "${CENSUS_DATA_FILE_URL}"

data/${CENSUS_DATA_FILE}.shp: data/${CENSUS_DATA_FILE}.zip
	# Get the shape file from the zipfile
	unzip "$<" -d data
	chmod -x data/*
	touch data/*

%.geojson: %.shp
	# Convert shape file to GeoJSON
	ogr2ogr -f GeoJSON -t_srs crs:84 "$@" "$<"

%.processed.geojson: %.geojson process.js
	npm i
	node process.js "$<" "$@"

%.mbtiles: %.processed.geojson
	tippecanoe -o "$@" -f -z 12 -Z 0 -B 0 -pS -pp -l districts -n "US Congressional Districts" "$<"

.PHONY: upload
upload: data/${CENSUS_DATA_FILE}.mbtiles
	node upload.js "$<" dshearer

.PHONY: clean
clean:
	rm -f data/*.shp data/*.geojson data/*.cpg data/*.dbf data/*.prj data/*.xml data/*.shx data/*.mbtiles

.PHONY: deepclean
deepclean: clean
	rm -rf node_modules