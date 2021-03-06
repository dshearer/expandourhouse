GO_LIB_SOURCES := \
	$(wildcard src/bulkInserter/*.go) \
	$(wildcard src/congresses/*.go) \
	$(wildcard src/housedb/reps/*.go) \
	$(wildcard src/housedb/turnout/*) \
	$(wildcard src/housedb/*.go) \
	$(wildcard src/states/*.go) \
	$(wildcard src/stylemetadata/*.go) \
	$(wildcard src/utils/*.go) \
	src/cmd/make-style/styleTemplate.go \
	src/housedb/turnout/turnoutData.go \
	src/states/states.go \
	src/go.mod \
	src/go.sum

_PROGRAMS = \
	add-district-pop \
	add-labels \
	compute-stats \
	congress-start-year \
	extract-states-for-year \
	make-style \
	mark-irregular \
	process-districts \
	reduce-precision \
	upload

GO = GOPATH="${TMP}" go

src/states/states.go: src/states/scripts/makeStates.go src/states/placeholder.go
	@echo GO GENERATE $@
	@cd src/states && ${GO} generate

src/cmd/make-style/styleTemplate.go: src/cmd/make-style/scripts/includeStyle.go \
	src/cmd/make-style/main.go src/cmd/make-style/style-template.json
	@echo GO GENERATE $@
	@cd src/cmd/make-style && ${GO} generate

src/housedb/turnout/turnoutData.go: src/housedb/turnout/scripts/includeData.go \
	src/housedb/turnout/*.tsv src/housedb/turnout/turnout.go
	@echo GO GENERATE $@
	@cd src/housedb/turnout && ${GO} generate

define PROGRAM_TARGETS
$${TMP}/${prog}: $$(wildcard src/cmd/${prog}/*.go) $${GO_LIB_SOURCES}
	@echo GO BUILD ${prog}
	@cd "src/cmd/${prog}" && $${GO} build -o "$$@"

clean-${prog}:
	@echo GO CLEAN ${prog}
	@cd "src/cmd/${prog}" && $${GO} clean

.PHONY: clean-${prog}
endef

$(foreach prog,${_PROGRAMS},$(eval ${PROGRAM_TARGETS}))

.PHONY: programs clean-programs

programs: $(patsubst %,${TMP}/%,${_PROGRAMS})

clean-programs: $(patsubst %,clean-%,${_PROGRAMS})
	@echo RM GENERATED GO SOURCE
	@rm -f src/cmd/make-style/styleTemplate.go src/states/states.go \
		src/housedb/turnout/turnoutData.go