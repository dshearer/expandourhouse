STATE_DATA_URL = https://publications.newberry.org/ahcbp/downloads/gis/US_AtlasHCB_StateTerr_Gen001.zip
DISTRICTS_DATA_URL = http://cdmaps.polisci.ucla.edu/shp

CONGRESSES = $(shell seq -f '%03g' 2 10)

STATE_TILES = $(patsubst %,data/%-states.mbtiles,${CONGRESSES})
DISTRICT_TILES = $(patsubst %,data/%-districts.mbtiles,${CONGRESSES})
STYLES = $(patsubst %,data/%-style.json,${CONGRESSES})

STATES_UPLOADED = $(patsubst %,%.uploaded,${STATE_TILES})
DISTRICTS_UPLOADED = $(patsubst %,%.uploaded,${DISTRICT_TILES})
STYLES_UPLOADED = $(patsubst %,%.uploaded,${STYLES})

ZIPS = data/US_AtlasHCB_StateTerr_Gen001.zip $(patsubst %,data/districts%.zip,${CONGRESSES})

MAPBOX_USER = dshearer

#data/US_AtlasHCB_StateTerr_Gen001/US_HistStateTerr_Gen001_Shapefile/US_HistStateTerr_Gen001.shp

.PHONY: all
all: ${STATE_TILES} ${DISTRICT_TILES} ${STYLES}

.PHONY: upload
upload: ${STYLES_UPLOADED} ${STATES_UPLOADED} ${DISTRICTS_UPLOADED}

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

define STATE_RECIPES

${congress}_CONGRESS_START_YEAR := $(shell node congress-start-year.js ${congress})

data/${congress}-states.geojson: data/states.geojson
	node extract-states-for-year.js "$$<" "$$$$(($${${congress}_CONGRESS_START_YEAR} - 1))" > "$$@"

data/${congress}-labelled-states.geojson: data/${congress}-states.geojson
	node add-labels.js "$$<" > "$$@"

data/${congress}-states.mbtiles: data/${congress}-labelled-states.geojson
	tippecanoe -o "$$@" -f -z 12 -Z 0 -B 0 -pS -pp -l states -n "states $$*" "$$<"

endef

$(foreach congress,${CONGRESSES},$(eval ${STATE_RECIPES}))

#################################################
# DISTRICTS
#################################################

define DISTRICT_RECIPES

data/${congress}-districts.zip:
	curl --fail-early --fail "${DISTRICTS_DATA_URL}/districts${congress}.zip" > "$$@"

data/${congress}-districts.geojson: data/${congress}-districts.zip
	# Get the shape file from the zipfile
	unzip -d "$$(dir $$@)" -jo "$$<"

	# Convert shape file to GeoJSON
	ogr2ogr -f GeoJSON -t_srs crs:84 "$$@" "data/districts${congress}.shp"

data/${congress}-proc-districts.geojson: data/${congress}-districts.geojson
	node process-from-ucla.js "$$<" "$$@"

data/${congress}-labelled-districts.geojson: data/${congress}-proc-districts.geojson
	node add-labels.js "$$<" > "$$@"

data/${congress}-districts.mbtiles: data/${congress}-labelled-districts.geojson
	tippecanoe -o "$$@" -f -z 12 -Z 0 -B 0 -pS -pp -l districts -n "district ${congress}" "$$<"

endef

$(foreach congress,${CONGRESSES},$(eval ${DISTRICT_RECIPES}))

#################################################
# STYLE
#################################################

define STYLE_RECIPES

data/${congress}-style.json: style-template.json
	node make-style.js "${congress}" "$${MAPBOX_USER}" > "$$@"

endef

$(foreach congress,${CONGRESSES},$(eval ${STYLE_RECIPES}))

#################################################
# UPLOADING
#################################################

define UPLOADED_RECIPES

data/${congress}-style.json.uploaded: data/${congress}-style.json
	node upload.js style "$$<" "$${MAPBOX_USER}"
	touch "$$@"

data/${congress}-states.mbtiles.uploaded: data/${congress}-style.json data/${congress}-states.mbtiles
	node upload.js states $$^
	touch "$$@"

data/${congress}-districts.mbtiles.uploaded: data/${congress}-style.json data/${congress}-districts.mbtiles
	node upload.js districts $$^
	touch "$$@"

endef

$(foreach congress,${CONGRESSES},$(eval ${UPLOADED_RECIPES}))

#################################################
# CLEANING
#################################################

.PHONY: clean
clean:
	rm -rf data/US_AtlasHCB_StateTerr_Gen001 data/*.shp data/*.shx data/*.mbtiles \
		data/*.geojson data/*.dbf data/*.prj data/*.uploaded