ROOT = ..

PROXY_DIR = ${ROOT}/.done-proxies
DOCKER_COMPOSE = cd "${ROOT}/local-dev" && docker-compose
DB_VOLUME = local-dev_db-data

.PHONY: help
help:
	@echo "Targets:"
	@echo "    build"
	@echo "    run"
	@echo "    logs"
	@echo "    clean"
	@echo "    deepclean"

.PHONY: build
build: ${PROXY_DIR}/image

${PROXY_DIR}/image: ${SOURCE}
	docker build -t "${DOCKER_IMAGE}" .
	mkdir -p "${PROXY_DIR}" && touch "$@"

.PHONY: run
run: ${PROXY_DIR}/image env
	docker container kill "${DOCKER_IMAGE}" 2>/dev/null ||:
	docker run ${DOCKER_OPTS} --name "${DOCKER_IMAGE}" --network local-dev_house --rm "${DOCKER_IMAGE}"

.PHONY: logs
logs:
	docker logs "${DOCKER_IMAGE}"

.PHONY: env
env: ${PROXY_DIR}/env

${PROXY_DIR}/env: ${ROOT}/schema.sql ${ROOT}/local-dev/docker-compose.yaml
	@echo "***** (Re)Making environment *****"
	$(call destroy-env)
	cp "${ROOT}/schema.sql" "${ROOT}/local-dev/postgres/"
	${DOCKER_COMPOSE} up --detach --build
	sleep 5
	mkdir -p "${PROXY_DIR}" && touch "$@"

.PHONY: clean
clean:
	-docker container kill "${DOCKER_IMAGE}"
	${DOCKER_COMPOSE} down
	rm -rf ${PROXY_DIR}

.PHONY: deepclean
deepclean: clean
	$(call destroy-env)

define destroy-env
	${DOCKER_COMPOSE} down --volumes --rmi local
	docker volume rm "${DB_VOLUME}" 2>/dev/null ||:
endef