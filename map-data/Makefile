STATE_DATA_URL = https://publications.newberry.org/ahcbp/downloads/gis/US_AtlasHCB_StateTerr_Gen001.zip
DISTRICTS_DATA_URL = http://cdmaps.polisci.ucla.edu/shp

CONGRESSES = $(shell seq -f '%03g' 100 100)

STATE_TILES = $(patsubst %,data/%.states-mbtiles,${CONGRESSES})
DISTRICT_TILES = $(patsubst %,data/%.districts-mbtiles,${CONGRESSES})
STYLE_UPLOADED = $(patsubst %,data/%.style-uploaded,${CONGRESSES})
STATES_UPLOADED = $(patsubst %,data/%.states-mbtiles-uploaded,${CONGRESSES})
DISTRICTS_UPLOADED = $(patsubst %,data/%.districts-mbtiles-uploaded,${CONGRESSES})
ZIPS = data/US_AtlasHCB_StateTerr_Gen001.zip $(patsubst %,data/districts%.zip,${CONGRESSES})

MAPBOX_USER = dshearer

#data/US_AtlasHCB_StateTerr_Gen001/US_HistStateTerr_Gen001_Shapefile/US_HistStateTerr_Gen001.shp

.PHONY: all
all: ${STATE_TILES} ${DISTRICT_TILES}

.PHONY: upload-styles
upload-styles: ${STYLE_UPLOADED}

.PHONY: upload
upload: upload-styles ${STATES_UPLOADED} ${DISTRICTS_UPLOADED}

.PHONY: zips
zips: ${ZIPS}

#################################################
# STATES
#################################################

data/US_AtlasHCB_StateTerr_Gen001.zip:
	curl --fail-early --fail "${STATE_DATA_URL}" > "$@"

data/states.geojson: data/US_AtlasHCB_StateTerr_Gen001.zip
	unzip -d data -o "$<"
	ogr2ogr -f GeoJSON -t_srs crs:84 "$@" "$$(find data -name US_HistStateTerr_Gen001.shp)"

data/%.states-geojson: data/states.geojson
	node extract-states-for-year.js "$<" "$$(($(shell node congress-start-year.js $*) - 1))" > "$@"

data/%.labelled-states-geojson: data/%.states-geojson
	node add-labels.js "$<" > "$@"
	cp "$@" "$@.bkp"

data/%.states-mbtiles: data/%.labelled-states-geojson
	tippecanoe -o "$@" -f -z 12 -Z 0 -B 0 -pS -pp -l states -n "states $*" "$<"

#################################################
# DISTRICTS
#################################################

data/districts%.zip:
	curl --fail-early --fail "${DISTRICTS_DATA_URL}/districts${*}.zip" > "$@"

data/%.districts-geojson: data/districts%.zip
	# Get the shape file from the zipfile
	unzip -d "$(dir $@)" -jo "$<"

	# Convert shape file to GeoJSON
	ogr2ogr -f GeoJSON -t_srs crs:84 "$@" "data/districts$*.shp"

data/%.proc-districts-geojson: data/%.districts-geojson
	node process-from-ucla.js "$<" "$@"

data/%.labelled-districts-geojson: data/%.proc-districts-geojson
	node add-labels.js "$<" > "$@"
	cp "$@" "$@.bkp"

data/%.districts-mbtiles: data/%.labelled-districts-geojson
	tippecanoe -o "$@" -f -z 12 -Z 0 -B 0 -pS -pp -l districts -n "district $*" "$<"

#################################################
# STYLE
#################################################

data/%.style: style-template.json
	node make-style.js "$*" "${MAPBOX_USER}" > "$@"

#################################################
# UPLOADING
#################################################

data/%.style-uploaded: data/%.style
	node upload.js style "$<" "${MAPBOX_USER}"
	touch "$@"

data/%.states-mbtiles-uploaded: data/%.style data/%.states-mbtiles
	node upload.js states $^
	touch "$@"

data/%.districts-mbtiles-uploaded: data/%.style data/%.districts-mbtiles
	node upload.js districts $^
	touch "$@"

.PHONY: clean-upload-styles
clean-upload-styles:
	rm -f data/*.style-uploaded

.PHONY: clean
clean:
	rm -rf data/US_AtlasHCB_StateTerr_Gen001 data/*.shp data/*.shx data/*mbtiles \
		data/states.geojson data/*.dbf data/*.prj data/*uploaded