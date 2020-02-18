###############################################################################
# DISTRICTS
#
# This Makefile defines recipes for making tile files for district borders.
# One tile file is made for a given session of Congress.
#
# Process (for nth session of Congress):
# I.   Download shape data for districts for nth Congress from
#      cdmaps.polisci.ucla.edu
# II.  Convert shape file into GeoJSON
# III. Process the district features, thus:
#    A. Clean up the feature's properties:
#        1. Parse the "ID" code into separate "stateFips", "congress",
#           and "district" properties
#        2. Add "state" property containing the USPS code for the state
#        3. Add "group" property containing "boundary" (used in MapBox style)
#    B. Filter out invalid districts
#    C. Add short and long titles as feature properties
# IV. Add label features, one for each district feature
# V.  Convert to tile file, and save it to output/n-districts.mbtiles
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
	"$${PROCESS_DISTRICTS}" < "$$<" > "$$@"

$${TMP}/${congress}-labelled-districts.geojson: $${TMP}/${congress}-proc-districts.geojson $${PROGRAMS}
	"$${ADD_LABELS}" < "$$<" > "$$@"

$${OUTPUT}/${congress}-districts.mbtiles: $${TMP}/${congress}-labelled-districts.geojson
	mkdir -p "$${OUTPUT}"
	tippecanoe -o "$$@" -f -z 12 -Z 0 -B 0 -pS -pp -l districts -n "district ${congress}" "$$<"

endef

$(foreach congress,${CONGRESSES},$(eval ${DISTRICT_RECIPES}))