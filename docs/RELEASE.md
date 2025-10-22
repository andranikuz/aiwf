# Release Management

This document describes how to manage releases using the Makefile.

## Quick Start

```bash
# Check current version
make version

# Show commits since last tag
make log

# Create a new release
make tag-patch    # Patch release: v1.0.1 → v1.0.2
make tag-minor    # Minor release: v1.0.1 → v1.1.0
make tag-major    # Major release: v1.0.1 → v2.0.0

# Update 'latest' tag
make latest-tag
```

## Semantic Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** (e.g., v1.0.0 → v2.0.0): Breaking API changes
- **MINOR** (e.g., v1.0.0 → v1.1.0): New features, backwards compatible
- **PATCH** (e.g., v1.0.0 → v1.0.1): Bug fixes

## Workflow

### 1. Develop Features
```bash
# Create branch, make commits
git checkout -b feature/new-feature
# ... make changes, commit ...
git push origin feature/new-feature
```

### 2. Create Pull Request
```bash
# Open PR on GitHub for review
```

### 3. Merge and Release
```bash
# Merge PR and go to main
git checkout main
git pull origin main

# Check what's new
make log

# Create release
make tag-patch    # or tag-minor / tag-major
```

### 4. Update Latest Tag
```bash
# Mark new version as latest
make latest-tag
```

## Available Commands

| Command | Purpose |
|---------|---------|
| `make help` | Show all available commands |
| `make version` | Display current version and git status |
| `make tag-patch` | Create patch release (x.y.z+1) |
| `make tag-minor` | Create minor release (x.y+1.0) |
| `make tag-major` | Create major release (x+1.0.0) |
| `make latest-tag` | Update 'latest' tag to current commit |
| `make status` | Show git status before tagging |
| `make log` | Show commits and stats since last tag |

## Examples

### Example 1: Patch Release (Bug Fix)

```bash
$ make version
Current version: v1.0.1
Git status:
 M src/fix.go

$ make log
=== Commits since v1.0.1 ===
a1b2c3d fix: handle edge case in type resolution
d4e5f6g fix: improve error messages

$ make tag-patch
=== Creating Patch Release ===
Current version: v1.0.1
New version: v1.0.2

Commits in this release:
a1b2c3d fix: handle edge case in type resolution
d4e5f6g fix: improve error messages

Create tag v1.0.2? [y/N] y
✓ Tag v1.0.2 created and pushed
✓ Run 'make latest-tag' to update 'latest' tag

$ make latest-tag
✓ 'latest' tag updated to v1.0.2
```

### Example 2: Minor Release (New Feature)

```bash
$ make tag-minor
=== Creating Minor Release ===
Current version: v1.0.2
New version: v1.1.0

Commits in this release:
x1y2z3a feat: add workflow orchestration
...

Create tag v1.1.0? [y/N] y
✓ Tag v1.1.0 created and pushed
```

### Example 3: Major Release (Breaking Change)

```bash
$ make tag-major
=== Creating Major Release ===
Current version: v1.1.5
New version: v2.0.0

Commits in this release:
...

Create tag v2.0.0? [y/N] y
✓ Tag v2.0.0 created and pushed
```

## Tags

The project maintains two important tags:

- **Version tags** (e.g., `v1.0.1`, `v1.1.0`): Specific releases
- **`latest` tag**: Points to the most recent stable release

### Checking Current Release

```bash
# Show current version tag
git describe --tags

# Show what's at 'latest'
git show-ref --tags latest
```

## Remote Management

When using `make tag-*` or `make latest-tag`, tags are automatically:
1. Created locally with annotated messages
2. Pushed to remote repository

## Manual Tag Operations

If you need to manage tags manually:

```bash
# List all tags
git tag -l

# Show commits for a tag
git log <tag>..HEAD --oneline

# Delete a tag locally
git tag -d v1.0.0

# Delete tag from remote
git push origin --delete v1.0.0

# Force update 'latest' (use with caution!)
git tag -f latest <commit-or-tag>
git push -f origin latest
```

## CI/CD Integration

The Makefile can be integrated with CI/CD:

```yaml
# Example GitHub Actions
- name: Check version
  run: make version

- name: Show release notes
  run: make log
```

## Troubleshooting

### "tag already exists"
```bash
# The tag you're trying to create already exists
git tag -l v1.0.1    # Verify
git show v1.0.1      # View tag details
```

### "Your branch is ahead of origin"
```bash
# Push changes before creating release tag
git push origin main
```

### "interactive prompt not working"
The Makefile uses `read` command. If interactive prompts don't work:

```bash
# Create tag manually
git tag -a v1.0.2 -m "Release v1.0.2"
git push origin v1.0.2
git tag -f latest v1.0.2
git push -f origin latest
```

## Release Notes Template

When creating a major release, include:

```markdown
# Release v2.0.0

## Breaking Changes
- API endpoint format changed (see migration guide)
- Renamed `TypeProvider` to `TypeSchema`

## New Features
- Workflow orchestration engine
- Multi-round dialog support
- Thread management for conversations

## Bug Fixes
- Fixed JSON schema validation
- Improved error handling

## Deprecations
- `old_method()` is deprecated, use `new_method()` instead

## Migration Guide
See [MIGRATION.md](./MIGRATION.md) for upgrade instructions.
```

## See Also

- [CHANGELOG.md](./CHANGELOG.md) - Detailed changelog
- [Contributing guide](./CONTRIBUTING.md) - How to contribute
- [Semantic Versioning](https://semver.org/) - Versioning specification
