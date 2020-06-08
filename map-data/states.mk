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
	@echo DOWNLOAD ${STATE_DATA_FILE}
	@mkdir -p "${TMP}"
	@curl --fail-early --fail "${STATE_DATA_URL}" > "$@" || rm "$@"

${TMP}/states.geojson: ${TMP}/${STATE_DATA_FILE}
	@echo UNZIP ${STATE_DATA_FILE}
	@unzip -d "${TMP}" -o "$<"
	@echo OGR2OGR US_HistStateTerr_Gen001.shp
	@ogr2ogr -f GeoJSON -t_srs crs:84 "$@" "$$(find "${TMP}" -name US_HistStateTerr_Gen001.shp)"

define STATE_RECIPES

$${TMP}/${congress}-states.geojson: $${TMP}/states.geojson $${EXTRACT_STATES_FOR_YEAR} $${CONGRESS_START_YEAR}
	@echo MAKE ${congress}-states.geojson
	@YEAR=$$$$(($$$$($${CONGRESS_START_YEAR} ${congress}) - 1)) && \
		"$${EXTRACT_STATES_FOR_YEAR}" "$$<" "$$$${YEAR}" > "$$@".tmp
	@mv "$$@".tmp "$$@"

$${TMP}/${congress}-proc-states.geojson: $${TMP}/${congress}-states.geojson $${ADD_LABELS} $${MARK_IRREG}
	@echo MAKE ${congress}-proc-states.geojson
	@"$${ADD_LABELS}" < "$$<" | "$${MARK_IRREG}" "${congress}" > "$$@".tmp
	@mv "$$@".tmp "$$@"

$${OUTPUT}/${congress}-states.mbtiles: $${TMP}/${congress}-proc-states.geojson
	@echo MAKE ${congress}-states.mbtiles
	@mkdir -p "$${OUTPUT}"
	@tippecanoe -o "$$@".tmp ${TIPPECANOE_OPTS} -l states -n "states ${congress}" "$$<"
	@mv "$$@".tmp "$$@"

endef

$(foreach congress,${CONGRESSES},$(eval ${STATE_RECIPES}))