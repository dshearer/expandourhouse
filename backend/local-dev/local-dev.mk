ROOT = ..

PROXY_DIR = ${ROOT}/.done-proxies

DOCKER_COMPOSE = cd "${ROOT}/local-dev" && docker-compose

.PHONY: help
help:
	@echo "Targets:"
	@echo "    run"
	@echo "    logs"
	@echo "    clean"
	@echo "    deepclean"

.PHONY: run
run: env
	docker build -t "${DOCKER_IMAGE}" .
	docker container kill "${DOCKER_IMAGE}" 2>/dev/null ||:
	docker run ${DOCKER_OPTS} --name "${DOCKER_IMAGE}" --network local-dev_house --rm "${DOCKER_IMAGE}"

.PHONY:
logs:
	docker logs "${DOCKER_IMAGE}"

env: ${ROOT}/schema.sql ${PROXY_DIR}/db-volume
	cp "$<" "${ROOT}/local-dev/postgres/"
	${DOCKER_COMPOSE} up --detach --build

${PROXY_DIR}/db-volume: ${ROOT}/schema.sql
	${DOCKER_COMPOSE} down
	docker volume rm local-dev_db-data 2>/dev/null ||:
	mkdir -p "${PROXY_DIR}" && touch "$@"

.PHONY: clean
clean:
	-docker container kill "${DOCKER_IMAGE}"
	${DOCKER_COMPOSE} down

.PHONY: deepclean
deepclean: clean
	${DOCKER_COMPOSE} down --volumes --rmi local