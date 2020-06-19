APP_DIR = ${PWD}/../app
STATS = ${APP_DIR}/src/stats.js

${APP_DIR}/src/stats.js: ${COMP_STATS}
	"${COMP_STATS}" > "$@"