STATE_DATA_URL = https://publications.newberry.org/ahcbp/downloads/gis/US_AtlasHCB_StateTerr_Gen001.zip
DISTRICTS_DATA_URL = http://cdmaps.polisci.ucla.edu/shp

CONGRESSES = $(shell seq -f '%03g' 2 5)

MAPBOX_USER = dshearer

# dirs
TMP = ${PWD}/.tmp
OUTPUT = ${PWD}/output

# targets
STATE_TILES = $(patsubst %,${OUTPUT}/%-states.mbtiles,${CONGRESSES})
STATES_UPLOADED = $(patsubst %,${TMP}/%-states.mbtiles.uploaded,${CONGRESSES})
DISTRICT_TILES = $(patsubst %,${OUTPUT}/%-districts.mbtiles,${CONGRESSES})
DISTRICTS_UPLOADED = $(patsubst %,${TMP}/%-districts.mbtiles.uploaded,${CONGRESSES})
STYLES = $(patsubst %,${OUTPUT}/%-style.json,${CONGRESSES})
STYLES_UPLOADED = $(patsubst %,${TMP}/%-style.json.uploaded,${CONGRESSES})

# programs
ADD_LABELS = ${TMP}/add-labels
CONGRESS_START_YEAR = ${TMP}/congress-start-year
EXTRACT_STATES_FOR_YEAR = ${TMP}/extract-states-for-year
MAKE_STYLE = ${TMP}/make-style
PROCESS_DISTRICTS = ${TMP}/process-districts
UPLOAD = ${TMP}/upload

.PHONY: all
all: ${STATE_TILES} ${DISTRICT_TILES} ${STYLES}

.PHONY: upload
upload: ${STYLES_UPLOADED} ${STATES_UPLOADED} ${DISTRICTS_UPLOADED}

include src/include.mk

#################################################
# STATES
#################################################

${TMP}/US_AtlasHCB_StateTerr_Gen001.zip:
	mkdir -p "${TMP}"
	curl --fail-early --fail "${STATE_DATA_URL}" > "$@"

${TMP}/states.geojson: ${TMP}/US_AtlasHCB_StateTerr_Gen001.zip
	unzip -d "${TMP}" -o "$<"
	ogr2ogr -f GeoJSON -t_srs crs:84 "$@" "$$(find "${TMP}" -name US_HistStateTerr_Gen001.shp)"

define STATE_RECIPES

$${TMP}/${congress}-states.geojson: $${TMP}/states.geojson $${CONGRESS_START_YEAR} $${EXTRACT_STATES_FOR_YEAR}
	"$${EXTRACT_STATES_FOR_YEAR}" "$$<" "$$$$(($$$$($${CONGRESS_START_YEAR} ${congress}) - 1))" > "$$@"

$${TMP}/${congress}-labelled-states.geojson: $${TMP}/${congress}-states.geojson $${ADD_LABELS}
	"$${ADD_LABELS}" < "$$<" > "$$@"

$${OUTPUT}/${congress}-states.mbtiles: $${TMP}/${congress}-labelled-states.geojson
	mkdir -p "$${OUTPUT}"
	tippecanoe -o "$$@" -f -z 12 -Z 0 -B 0 -pS -pp -l states -n "states ${congress}" "$$<"

endef

#################################################
# DISTRICTS
#################################################

define DISTRICT_RECIPES

$${TMP}/${congress}-districts.zip:
	mkdir -p "${TMP}"
	curl --fail-early --fail "${DISTRICTS_DATA_URL}/districts${congress}.zip" > "$$@"

$${TMP}/${congress}-districts.geojson: $${TMP}/${congress}-districts.zip
	# Get the shape file from the zipfile99p
	unzip -d "$$(dir $$@)" -jo "$$<"

	# Convert shape file to GeoJSON
	ogr2ogr -f GeoJSON -t_srs crs:84 "$$@" "$${TMP}/districts${congress}.shp"

$${TMP}/${congress}-proc-districts.geojson: $${TMP}/${congress}-districts.geojson $${PROCESS_DISTRICTS}
	"$${PROCESS_DISTRICTS}" < "$$<" > "$$@"

$${TMP}/${congress}-labelled-districts.geojson: $${TMP}/${congress}-proc-districts.geojson $${ADD_LABELS}
	"$${ADD_LABELS}" < "$$<" > "$$@"

$${OUTPUT}/${congress}-districts.mbtiles: $${TMP}/${congress}-labelled-districts.geojson
	mkdir -p "$${OUTPUT}"
	tippecanoe -o "$$@" -f -z 12 -Z 0 -B 0 -pS -pp -l districts -n "district ${congress}" "$$<"

endef

$(foreach congress,${CONGRESSES},$(eval ${DISTRICT_RECIPES}))

#################################################
# STYLE
#################################################

define STYLE_RECIPES

$${OUTPUT}/${congress}-style.json: $${MAKE_STYLE}
	mkdir -p "$${OUTPUT}"
	"$${MAKE_STYLE}" "${congress}" "$${MAPBOX_USER}" > "$$@"

endef

$(foreach congress,${CONGRESSES},$(eval ${STYLE_RECIPES}))

#################################################
# UPLOADING
#################################################

define UPLOADED_RECIPES

$${TMP}/${congress}-style.json.uploaded: $${OUTPUT}/${congress}-style.json $${UPLOAD}
	"$${UPLOAD}" style "$${MAPBOX_USER}" "$$<"
	touch "$$@"

$${TMP}/${congress}-states.mbtiles.uploaded: $${OUTPUT}/${congress}-style.json $${OUTPUT}/${congress}-states.mbtiles $${UPLOAD}
	"$${UPLOAD}" states "$${MAPBOX_USER}" "$${OUTPUT}/${congress}-style.json" "$${OUTPUT}/${congress}-states.mbtiles"
	touch "$$@"

$${TMP}/${congress}-districts.mbtiles.uploaded: $${OUTPUT}/${congress}-style.json $${OUTPUT}/${congress}-districts.mbtiles $${UPLOAD}
	"$${UPLOAD}" districts "$${MAPBOX_USER}" "$${OUTPUT}/${congress}-style.json" "$${OUTPUT}/${congress}-districts.mbtiles"
	touch "$$@"

endef

$(foreach congress,${CONGRESSES},$(eval ${UPLOADED_RECIPES}))

#################################################
# CLEAN
#################################################

.PHONY: clean
clean: clean-programs
	rm -rf "${TMP}"/*.geojson "${TMP}"/*.mbtiles *_start_year \
		$(patsubst %,${TMP}/%,${PROGRAMS}) ${STYLES} ${OUTPUT}

$(foreach congress,${CONGRESSES},$(eval ${STATE_RECIPES}))