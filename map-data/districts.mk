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
	mkdir -p "${TMP}"
	curl --fail-early --fail "${DISTRICTS_DATA_URL}/districts${congress}.zip" > "$$@"

$${TMP}/${congress}-districts.geojson: $${TMP}/${congress}-districts.zip
	# Get the shape file from the zipfile
	unzip -d "$$(dir $$@)" -jo "$$<"

	# Convert shape file to GeoJSON
	ogr2ogr -f GeoJSON -t_srs crs:84 "$$@" "$${TMP}/districts${congress}.shp"

$${TMP}/${congress}-proc-districts.geojson: $${TMP}/${congress}-districts.geojson $${PROGRAMS}
	"$${PROCESS_DISTRICTS}" < "$$<" | "$${REDUCE_PRECISION}" 3 | "$${ADD_LABELS}" > "$$@"

$${OUTPUT}/${congress}-districts.mbtiles: $${TMP}/${congress}-proc-districts.geojson
	mkdir -p "$${OUTPUT}"
	tippecanoe -o "$$@" -f -z 12 -Z 0 -B 0 -pS -pp --read-parallel -l districts -n "district ${congress}" "$$<"

endef

$(foreach congress,${CONGRESSES},$(eval ${DISTRICT_RECIPES}))