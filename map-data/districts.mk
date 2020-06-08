###############################################################################
# DISTRICTS
#
# This Makefile defines recipes for making tile files for district borders.
# One tile file is made for a given session of Congress.
###############################################################################

DISTRICTS_DATA_URL = http://cdmaps.polisci.ucla.edu/shp

DISTRICT_TILES = $(patsubst %,${OUTPUT}/%-districts.mbtiles,${CONGRESSES})

define DISTRICT_RECIPES

$${TMP}/${congress}-districts.zip:
	@echo DOWNLOAD ${congress}-districts.zip
	@mkdir -p "${TMP}"
	@curl --fail-early --fail "${DISTRICTS_DATA_URL}/districts${congress}.zip" > "$$@"

$${TMP}/${congress}-districts.geojson: $${TMP}/${congress}-districts.zip
	# Get the shape file from the zipfile
	@echo UNZIP ${congress}-districts.zip
	@unzip -d "$$(dir $$@)" -jo "$$<"

	# Convert shape file to GeoJSON
	@echo OGR2OGR ${congress}-districts.geojson
	@ogr2ogr -f GeoJSON -t_srs crs:84 "$$@" "$${TMP}/districts${congress}.shp"

$${TMP}/${congress}-proc-districts.geojson: $${TMP}/${congress}-districts.geojson $${PROCESS_DISTRICTS} $${ADD_LABELS} $${MARK_IRREG} $${ADD_DIST_POP}
	@echo MAKE ${congress}-proc-districts.geojson
	@"$${PROCESS_DISTRICTS}" "${congress}" < "$$<" | "$${ADD_LABELS}" | "$${MARK_IRREG}" "${congress}" \
		| "$${ADD_DIST_POP}" > "$$@".tmp
	@mv "$$@".tmp "$$@"

$${OUTPUT}/${congress}-districts.mbtiles: $${TMP}/${congress}-proc-districts.geojson
	@echo MAKE ${congress}-districts.mbtiles
	@mkdir -p "$${OUTPUT}"
	@tippecanoe -o "$$@".tmp ${TIPPECANOE_OPTS} -l districts -n "district ${congress}" "$$<"
	@mv "$$@".tmp "$$@"

endef

$(foreach congress,${CONGRESSES},$(eval ${DISTRICT_RECIPES}))