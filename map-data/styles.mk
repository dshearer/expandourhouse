###############################################################################
# STYLES
#
# This Makefile defines recipes for making MapBox style files.
###############################################################################

STYLES = $(patsubst %,${OUTPUT}/%-style.json,${CONGRESSES})

define STYLE_RECIPES

$${OUTPUT}/${congress}-style.json: $${MAKE_STYLE}
	@echo MAKE-STYLE ${congress}-style.json
	@mkdir -p "$${OUTPUT}"
	@"$${MAKE_STYLE}" "${congress}" "$${MAPBOX_USER}" > "$$@"

endef

$(foreach congress,${CONGRESSES},$(eval ${STYLE_RECIPES}))