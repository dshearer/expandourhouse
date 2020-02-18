###############################################################################
# UPLOADING
###############################################################################

STATES_UPLOADED = $(patsubst %,${TMP}/%-states.mbtiles.uploaded,${CONGRESSES})
DISTRICTS_UPLOADED = $(patsubst %,${TMP}/%-districts.mbtiles.uploaded,${CONGRESSES})
STYLES_UPLOADED = $(patsubst %,${TMP}/%-style.json.uploaded,${CONGRESSES})

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