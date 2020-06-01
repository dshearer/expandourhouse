###############################################################################
# PARAMETERS 
#
# Use these to control for which congresses we will generate map data.
###############################################################################
FIRST_CONGRESS = 13
LAST_CONGRESS = 13

# derived params
CONGRESSES = $(shell seq -f '%03g' ${FIRST_CONGRESS} ${LAST_CONGRESS})

# Settings
MAPBOX_USER = dshearer
COORD_PRECISION = 6

# dirs
TMP = ${PWD}/.tmp
OUTPUT = ${PWD}/output

# programs
ADD_LABELS = ${TMP}/add-labels
CONGRESS_START_YEAR = ${TMP}/congress-start-year
EXTRACT_STATES_FOR_YEAR = ${TMP}/extract-states-for-year
MAKE_STYLE = ${TMP}/make-style
MARK_IRREG = ${TMP}/mark-irregular
PROCESS_DISTRICTS = ${TMP}/process-districts
REDUCE_PRECISION = ${TMP}/reduce-precision
UPLOAD = ${TMP}/upload
PROGRAMS = \
	${ADD_LABELS} \
	${CONGRESS_START_YEAR} \
	${EXTRACT_STATES_FOR_YEAR} \
	${MAKE_STYLE} \
	${MARK_IRREG_STATES} \
	${PROCESS_DISTRICTS} \
	${REDUCE_PRECISION} \
	${UPLOAD}

TIPPECANOE_OPTS = --force -z 10 -Z 0 --no-line-simplification -b0 -r1 -pk -pf
#--read-parallel

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
build: build-states build-districts build-styles

.PHONY: build-states
build-states: ${STATE_TILES}

.PHONY: build-districts
build-districts: ${DISTRICT_TILES}

.PHONY: build-styles
build-styles: ${STYLES}

.PHONY: upload
upload: upload-states upload-districts upload-styles

.PHONY: upload-states
upload-states: ${STATES_UPLOADED}

.PHONY: upload-districts
upload-districts: ${DISTRICTS_UPLOADED}

.PHONY: upload-styles
upload-styles: ${STYLES_UPLOADED}

#################################################
# CLEAN
#################################################

.PHONY: clean
clean: clean-programs
	rm -rf "${TMP}"/*.geojson "${TMP}"/*.mbtiles *_start_year \
		$(patsubst %,${TMP}/%,${PROGRAMS}) ${STYLES} ${OUTPUT}
