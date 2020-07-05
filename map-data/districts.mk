###############################################################################
# DISTRICTS
#
# This Makefile defines recipes for making tile files for district borders.
# One tile file is made for a given session of Congress.
###############################################################################

DISTRICTS_DATA_URL = http://cdmaps.polisci.ucla.edu/shp

DISTRICT_TILES = $(patsubst %,${OUTPUT}/%-districts.mbtiles,${CONGRESSES})

define DISTRICT_RECIPES

$${DOWNLOADS}/${congress}-districts.zip:
	@echo DOWNLOAD ${congress}-districts.zip
	@mkdir -p "$${DOWNLOADS}"
	@curl --fail-early --fail "${DISTRICTS_DATA_URL}/districts${congress}.zip" > "$$@"

$${TMP}/${congress}-districts.geojson: $${DOWNLOADS}/${congress}-districts.zip
	# Get the shape file from the zipfile
	@echo UNZIP ${congress}-districts.zip
	@unzip -d "$$(dir $$@)" -jo "$$<"

	# Convert shape file to GeoJSON
	@echo OGR2OGR ${congress}-districts.geojson
	@ogr2ogr -f GeoJSON -t_srs crs:84 "$$@" "$${TMP}/districts${congress}.shp"

$${TMP}/${congress}-proc-districts-1.geojson: $${TMP}/${congress}-districts.geojson $${PROCESS_DISTRICTS}
	"$${PROCESS_DISTRICTS}" "${congress}" < "$$<" > "$$@"

$${TMP}/${congress}-proc-districts-2.geojson: $${TMP}/${congress}-proc-districts-1.geojson $${ADD_LABELS}
	"$${ADD_LABELS}" < "$$<" > "$$@"

$${TMP}/${congress}-proc-districts-3.geojson: $${TMP}/${congress}-proc-districts-2.geojson $${MARK_IRREG}
	"$${MARK_IRREG}" "${congress}" < "$$<" > "$$@"

$${TMP}/${congress}-proc-districts.geojson: $${TMP}/${congress}-proc-districts-3.geojson $${ADD_DIST_POP}
	"$${ADD_DIST_POP}" < "$$<" > "$$@"

$${OUTPUT}/${congress}-districts.mbtiles: $${TMP}/${congress}-proc-districts.geojson
	@echo MAKE ${congress}-districts.mbtiles
	@mkdir -p "$${OUTPUT}"
	@tippecanoe -o "$$@".tmp ${TIPPECANOE_OPTS} -l districts -n "district ${congress}" "$$<"
	@mv "$$@".tmp "$$@"

endef

$(foreach congress,${CONGRESSES},$(eval ${DISTRICT_RECIPES}))