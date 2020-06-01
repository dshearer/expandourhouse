###############################################################################
# STATES
#
# This Makefile defines recipes for making tile files for state borders.
# One tile file is made for a given session of Congress.
###############################################################################

STATE_DATA_FILE = US_AtlasHCB_StateTerr_Gen001.zip
STATE_DATA_URL = https://publications.newberry.org/ahcbp/downloads/gis/${STATE_DATA_FILE}
STATE_TILES = $(patsubst %,${OUTPUT}/%-states.mbtiles,${CONGRESSES})

${TMP}/${STATE_DATA_FILE}:
	mkdir -p "${TMP}"
	curl --fail-early --fail "${STATE_DATA_URL}" > "$@"

${TMP}/states.geojson: ${TMP}/${STATE_DATA_FILE}
	unzip -d "${TMP}" -o "$<"
	ogr2ogr -f GeoJSON -t_srs crs:84 "$@" "$$(find "${TMP}" -name US_HistStateTerr_Gen001.shp)"

define STATE_RECIPES

$${TMP}/${congress}-states.geojson: $${TMP}/states.geojson $${EXTRACT_STATES_FOR_YEAR}
	YEAR=$$$$(($$$$($${CONGRESS_START_YEAR} ${congress}) - 1)) && \
	"$${EXTRACT_STATES_FOR_YEAR}" "$$<" "$$$${YEAR}" > "$$@"

$${TMP}/${congress}-proc-states.geojson: $${TMP}/${congress}-states.geojson $${ADD_LABELS} $${MARK_IRREG}
	"$${ADD_LABELS}" < "$$<" | "$${MARK_IRREG}" "${congress}" > "$$@"

$${OUTPUT}/${congress}-states.mbtiles: $${TMP}/${congress}-proc-states.geojson
	mkdir -p "$${OUTPUT}"
	tippecanoe -o "$$@" ${TIPPECANOE_OPTS} -l states -n "states ${congress}" "$$<"

endef

$(foreach congress,${CONGRESSES},$(eval ${STATE_RECIPES}))