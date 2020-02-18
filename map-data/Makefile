###############################################################################
# PARAMETERS 
#
# Use these to control for which congresses we will generate map data.
###############################################################################
FIRST_CONGRESS = 11
LAST_CONGRESS = 11

# derived params
CONGRESSES = $(shell seq -f '%03g' ${FIRST_CONGRESS} ${LAST_CONGRESS})

# MapBox settings
MAPBOX_USER = dshearer

# dirs
TMP = ${PWD}/.tmp
OUTPUT = ${PWD}/output

# programs
ADD_LABELS = ${TMP}/add-labels
CONGRESS_START_YEAR = ${TMP}/congress-start-year
EXTRACT_STATES_FOR_YEAR = ${TMP}/extract-states-for-year
MAKE_STYLE = ${TMP}/make-style
PROCESS_DISTRICTS = ${TMP}/process-districts
UPLOAD = ${TMP}/upload
PROGRAMS = \
	${ADD_LABELS} \
	${CONGRESS_START_YEAR} \
	${EXTRACT_STATES_FOR_YEAR} \
	${MAKE_STYLE} \
	${PROCESS_DISTRICTS} \
	${UPLOAD}

.PHONY: default
default: build

include src/include.mk
include states.mk
include districts.mk
include styles.mk
include uploading.mk

#################################################
# MAIN RECIPES
#################################################

.PHONY: build
build: ${STATE_TILES} ${DISTRICT_TILES} ${STYLES}

.PHONY: upload
upload: ${STYLES_UPLOADED} ${STATES_UPLOADED} ${DISTRICTS_UPLOADED}

#################################################
# CLEAN
#################################################

.PHONY: clean
clean: clean-programs
	rm -rf "${TMP}"/*.geojson "${TMP}"/*.mbtiles *_start_year \
		$(patsubst %,${TMP}/%,${PROGRAMS}) ${STYLES} ${OUTPUT}
