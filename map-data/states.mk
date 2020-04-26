###############################################################################
# STATES
#
# This Makefile defines recipes for making tile files for state borders.
# One tile file is made for a given session of Congress.
###############################################################################

STATE_DATA_URL = https://publications.newberry.org/ahcbp/downloads/gis/US_AtlasHCB_StateTerr_Gen001.zip

STATE_TILES = $(patsubst %,${OUTPUT}/%-states.mbtiles,${CONGRESSES})

${TMP}/US_AtlasHCB_StateTerr_Gen001.zip:
	mkdir -p "${TMP}"
	curl --fail-early --fail "${STATE_DATA_URL}" > "$@"

${TMP}/states.geojson: ${TMP}/US_AtlasHCB_StateTerr_Gen001.zip
	unzip -d "${TMP}" -o "$<"
	ogr2ogr -f GeoJSON -t_srs crs:84 "$@" "$$(find "${TMP}" -name US_HistStateTerr_Gen001.shp)"

define STATE_RECIPES

$${TMP}/${congress}-states.geojson: $${TMP}/states.geojson $${PROGRAMS}
	YEAR=$$$$(($$$$($${CONGRESS_START_YEAR} ${congress}) - 1)) && \
	"$${EXTRACT_STATES_FOR_YEAR}" "$$<" "$$$${YEAR}" > "$$@"

$${TMP}/${congress}-labelled-states.geojson: $${TMP}/${congress}-states.geojson $${PROGRAMS}
	"$${ADD_LABELS}" < "$$<" | "$${MARK_IRREG_STATES}" ${congress} > "$$@"

$${OUTPUT}/${congress}-states.mbtiles: $${TMP}/${congress}-labelled-states.geojson
	mkdir -p "$${OUTPUT}"
	tippecanoe -o "$$@" -f -z 12 -Z 0 -B 0 -pS -pp -l states -n "states ${congress}" "$$<"

endef

$(foreach congress,${CONGRESSES},$(eval ${STATE_RECIPES}))