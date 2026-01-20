GO        ?= go
COVERMODE ?= atomic

BADGE_DIR   ?= .public
BADGE_LABEL ?= coverage

COVERPROFILE_ALL ?= coverage.out

WORK_MODULES ?= cqrs di env meta mut oas search util validate
WORK_PACKAGES := $(addsuffix /...,$(addprefix ./,$(WORK_MODULES)))

BADGE_TOOL_PKG ?= ./_tools/coveragebadge

.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo "Targets:"
	@echo "  make test            # go test em todos os módulos"
	@echo "  make cover           # gera $(COVERPROFILE_ALL)"
	@echo "  make badge           # gera $(BADGE_DIR)/coverage.svg"
	@echo "  make module-badges   # gera $(BADGE_DIR)/*-coverage.svg"
	@echo "  make badges          # geral + por módulo"
	@echo "  make ci              # test + badges"
	@echo "  make clean           # remove profiles e svgs"

.PHONY: test
test:
	$(GO) test $(WORK_PACKAGES)

# ---- coverage geral (arquivo de verdade) ----
$(COVERPROFILE_ALL):
	$(GO) test $(WORK_PACKAGES) -coverprofile=$@ -covermode=$(COVERMODE)

.PHONY: cover
cover: $(COVERPROFILE_ALL)

# ---- badge geral ----
$(BADGE_DIR)/coverage.svg: $(COVERPROFILE_ALL)
	$(GO) run $(BADGE_TOOL_PKG) -in $< -out $@ -label $(BADGE_LABEL)

.PHONY: badge
badge: $(BADGE_DIR)/coverage.svg

# ---- regras por módulo ----
define module_rules
$(1).coverage.out:
	$(GO) test ./$(1)/... -coverprofile=$$@ -covermode=$(COVERMODE)

$(BADGE_DIR)/$(1)-coverage.svg: $(1).coverage.out
	$(GO) run $(BADGE_TOOL_PKG) -in $$< -out $$@ -label $(BADGE_LABEL)
endef

$(foreach m,$(WORK_MODULES),$(eval $(call module_rules,$(m))))

.PHONY: module-badges
module-badges: $(foreach m,$(WORK_MODULES),$(BADGE_DIR)/$(m)-coverage.svg)

.PHONY: badges
badges: badge module-badges

.PHONY: ci
ci: test badges

.PHONY: clean
clean:
	$(GO) clean -testcache
	rm -rf $(COVERPROFILE_ALL) $(BADGE_DIR)/*.svg
