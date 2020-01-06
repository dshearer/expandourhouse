GO_LIB_SOURCES := \
	$(wildcard src/congresses/*.go) \
	$(wildcard src/featureProc/*.go) \
	$(wildcard src/states/*.go) \
	$(wildcard src/stylemetadata/*.go) \
	$(wildcard src/utils/*.go) \
	src/go.mod \
	src/go.sum

PROGRAMS = \
	add-labels \
	chop \
	congress-start-year \
	extract-states-for-year \
	make-style \
	process-districts \
	upload

GO = GOPATH="${TMP}" go

define PROGRAM_TARGETS
$${TMP}/${prog}: $$(wildcard src/cmd/${prog}/*.go) $${GO_LIB_SOURCES}
	cd "src/cmd/${prog}" && $${GO} generate && $${GO} build -o "$$@"

clean-${prog}:
	cd "src/cmd/${prog}" && $${GO} clean
endef

$(foreach prog,${PROGRAMS},$(eval ${PROGRAM_TARGETS}))

.PHONY: clean-programs
clean-programs:
	rm -f src/cmd/make-style/styleTemplate.go src/states/states.go