.PHONY: help version tag tag-patch tag-minor tag-major latest-tag clean

# Color output
GREEN := \033[0;32m
BLUE := \033[0;34m
YELLOW := \033[0;33m
NC := \033[0m # No Color

# Get current version
CURRENT_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
CURRENT_MAJOR := $(shell echo $(CURRENT_VERSION) | cut -d. -f1 | cut -dv -f2)
CURRENT_MINOR := $(shell echo $(CURRENT_VERSION) | cut -d. -f2)
CURRENT_PATCH := $(shell echo $(CURRENT_VERSION) | cut -d. -f3)

help:
	@echo "$(BLUE)=== AIWF Release Management ===$(NC)"
	@echo ""
	@echo "$(GREEN)Available targets:$(NC)"
	@echo "  $(YELLOW)make version$(NC)              - Show current version"
	@echo "  $(YELLOW)make tag-patch$(NC)            - Create patch release (e.g. v0.9.0 → v0.9.1)"
	@echo "  $(YELLOW)make tag-minor$(NC)            - Create minor release (e.g. v0.9.0 → v0.10.0)"
	@echo "  $(YELLOW)make tag-major$(NC)            - Create major release (e.g. v0.9.0 → v1.0.0)"
	@echo "  $(YELLOW)make latest-tag$(NC)           - Tag current commit as latest"
	@echo "  $(YELLOW)make status$(NC)               - Show git status before tagging"
	@echo "  $(YELLOW)make log$(NC)                  - Show commits since last tag"
	@echo ""
	@echo "$(BLUE)Examples:$(NC)"
	@echo "  $(YELLOW)make tag-patch$(NC)            # Creates v0.9.1 from v0.9.0"
	@echo "  $(YELLOW)make tag-minor$(NC)            # Creates v0.10.0 from v0.9.0"
	@echo ""

version:
	@echo "$(GREEN)Current version:$(NC) $(CURRENT_VERSION)"
	@echo "$(GREEN)Git status:$(NC)"
	@git status --short

status:
	@echo "$(BLUE)=== Git Status ===$(NC)"
	@git status
	@echo ""
	@echo "$(BLUE)=== Commits since last tag ===$(NC)"
	@git log $(CURRENT_VERSION)..HEAD --oneline || echo "No commits since last tag"

log:
	@echo "$(BLUE)=== Commits since $(CURRENT_VERSION) ===$(NC)"
	@git log $(CURRENT_VERSION)..HEAD --oneline || echo "No commits since last tag"
	@echo ""
	@echo "$(BLUE)=== Git log stats ===$(NC)"
	@git log $(CURRENT_VERSION)..HEAD --stat || true

tag-patch:
	@echo "$(BLUE)=== Creating Patch Release ===$(NC)"
	@NEW_VERSION=v$(CURRENT_MAJOR).$(CURRENT_MINOR).$$(($(CURRENT_PATCH) + 1)); \
	echo "$(GREEN)Current version:$(NC) $(CURRENT_VERSION)"; \
	echo "$(GREEN)New version:$(NC) $$NEW_VERSION"; \
	echo ""; \
	echo "$(YELLOW)Commits in this release:$(NC)"; \
	git log $(CURRENT_VERSION)..HEAD --oneline; \
	echo ""; \
	read -p "$(YELLOW)Create tag $$NEW_VERSION? [y/N]$(NC) " -n 1 -r; \
	echo ""; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		git tag -a $$NEW_VERSION -m "Release $$NEW_VERSION"; \
		git push origin $$NEW_VERSION; \
		echo "$(GREEN)✓ Tag $$NEW_VERSION created and pushed$(NC)"; \
		echo "$(GREEN)✓ Run 'make latest-tag' to update 'latest' tag$(NC)"; \
	else \
		echo "$(YELLOW)Cancelled$(NC)"; \
	fi

tag-minor:
	@echo "$(BLUE)=== Creating Minor Release ===$(NC)"
	@NEW_VERSION=v$(CURRENT_MAJOR).$$(($(CURRENT_MINOR) + 1)).0; \
	echo "$(GREEN)Current version:$(NC) $(CURRENT_VERSION)"; \
	echo "$(GREEN)New version:$(NC) $$NEW_VERSION"; \
	echo ""; \
	echo "$(YELLOW)Commits in this release:$(NC)"; \
	git log $(CURRENT_VERSION)..HEAD --oneline; \
	echo ""; \
	read -p "$(YELLOW)Create tag $$NEW_VERSION? [y/N]$(NC) " -n 1 -r; \
	echo ""; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		git tag -a $$NEW_VERSION -m "Release $$NEW_VERSION"; \
		git push origin $$NEW_VERSION; \
		echo "$(GREEN)✓ Tag $$NEW_VERSION created and pushed$(NC)"; \
		echo "$(GREEN)✓ Run 'make latest-tag' to update 'latest' tag$(NC)"; \
	else \
		echo "$(YELLOW)Cancelled$(NC)"; \
	fi

tag-major:
	@echo "$(BLUE)=== Creating Major Release ===$(NC)"
	@NEW_VERSION=v$$(($(CURRENT_MAJOR) + 1)).0.0; \
	echo "$(GREEN)Current version:$(NC) $(CURRENT_VERSION)"; \
	echo "$(GREEN)New version:$(NC) $$NEW_VERSION"; \
	echo ""; \
	echo "$(YELLOW)Commits in this release:$(NC)"; \
	git log $(CURRENT_VERSION)..HEAD --oneline; \
	echo ""; \
	read -p "$(YELLOW)Create tag $$NEW_VERSION? [y/N]$(NC) " -n 1 -r; \
	echo ""; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		git tag -a $$NEW_VERSION -m "Release $$NEW_VERSION"; \
		git push origin $$NEW_VERSION; \
		echo "$(GREEN)✓ Tag $$NEW_VERSION created and pushed$(NC)"; \
		echo "$(GREEN)✓ Run 'make latest-tag' to update 'latest' tag$(NC)"; \
	else \
		echo "$(YELLOW)Cancelled$(NC)"; \
	fi

latest-tag:
	@echo "$(BLUE)=== Updating 'latest' Tag ===$(NC)"
	@LATEST_TAG=$$(git describe --tags --abbrev=0); \
	echo "$(GREEN)Latest version:$(NC) $$LATEST_TAG"; \
	git tag -f latest $$LATEST_TAG; \
	git push -f origin latest; \
	echo "$(GREEN)✓ 'latest' tag updated to $$LATEST_TAG$(NC)"

clean:
	@echo "$(YELLOW)This command does not delete tags.$(NC)"
	@echo "$(YELLOW)To delete a tag, use:$(NC)"
	@echo "  git tag -d <tag-name>                 # Delete locally"
	@echo "  git push origin :<tag-name>           # Delete on remote"
	@echo "  git push origin --delete <tag-name>   # Delete on remote (safer)"

.PHONY: help version tag-patch tag-minor tag-major latest-tag status log clean
