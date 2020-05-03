###############################################################################
# UPLOADING
###############################################################################

STATES_UPLOADED = $(patsubst %,upload-states-%,${CONGRESSES})
DISTRICTS_UPLOADED = $(patsubst %,upload-districts-%,${CONGRESSES})
STYLES_UPLOADED = $(patsubst %,upload-style-%,${CONGRESSES})

define UPLOADED_RECIPES

.PHONY: upload-style-${congress}
upload-style-${congress}: $${OUTPUT}/${congress}-style.json $${UPLOAD}
	"$${UPLOAD}" style "$${MAPBOX_USER}" "$$<"

.PHONY: upload-states-${congress}
upload-states-${congress}: $${OUTPUT}/${congress}-style.json $${OUTPUT}/${congress}-states.mbtiles $${UPLOAD}
	"$${UPLOAD}" states "$${MAPBOX_USER}" "$${OUTPUT}/${congress}-style.json" "$${OUTPUT}/${congress}-states.mbtiles"

.PHONY: upload-districts-${congress}
upload-districts-${congress}: $${OUTPUT}/${congress}-style.json $${OUTPUT}/${congress}-districts.mbtiles $${UPLOAD}
	"$${UPLOAD}" districts "$${MAPBOX_USER}" "$${OUTPUT}/${congress}-style.json" "$${OUTPUT}/${congress}-districts.mbtiles"

endef

$(foreach congress,${CONGRESSES},$(eval ${UPLOADED_RECIPES}))