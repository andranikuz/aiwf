# Release Cheatsheet

Quick reference for the most common release tasks.

## TL;DR - Common Workflow

```bash
# Before releasing, check what changed
make log

# Create the appropriate release
make tag-patch    # Bug fixes: v1.0.1 → v1.0.2
make tag-minor    # New features: v1.0.1 → v1.1.0
make tag-major    # Breaking changes: v1.0.1 → v2.0.0

# Update 'latest' tag (done automatically, but you can refresh)
make latest-tag
```

## One-Liners

```bash
# What's the current version?
make version

# What changed since last release?
make log

# I just merged a bug fix, release it
make tag-patch

# I added cool new features, release it
make tag-minor

# I broke the API, release it
make tag-major

# Show all available commands
make help
```

## Step-by-Step Examples

### Bug Fix Release

```bash
# 1. Check what's new
$ make log
=== Commits since v1.0.1 ===
a1b2c3d fix: handle edge case
d4e5f6g fix: improve error message

# 2. Create patch release
$ make tag-patch
Current version: v1.0.1
New version: v1.0.2
Create tag v1.0.2? [y/N] y
✓ Tag v1.0.2 created and pushed

# 3. Mark as latest (optional, automatic)
$ make latest-tag
✓ 'latest' tag updated to v1.0.2
```

### Feature Release

```bash
# 1. Check commits
$ make log
=== Commits since v1.0.2 ===
x1y2z3a feat: add workflow support
p9q8r7s feat: add dialog support

# 2. Create minor release
$ make tag-minor
Current version: v1.0.2
New version: v1.1.0
Create tag v1.1.0? [y/N] y
✓ Tag v1.1.0 created and pushed

# 3. Update latest
$ make latest-tag
✓ 'latest' tag updated to v1.1.0
```

### Breaking Change Release

```bash
# 1. Check what changed
$ make log
=== Commits since v1.1.5 ===
m5n4o3p BREAKING: redesign TypeProvider API
l2k1j0i BREAKING: change Agent interface

# 2. Create major release
$ make tag-major
Current version: v1.1.5
New version: v2.0.0
Create tag v2.0.0? [y/N] y
✓ Tag v2.0.0 created and pushed

# 3. Verify
$ make version
Current version: v2.0.0
```

## Version Numbers

**Format:** `vMAJOR.MINOR.PATCH`

| Scenario | Change | Example |
|----------|--------|---------|
| Bug fix | PATCH+1 | v1.0.0 → v1.0.1 |
| New feature | MINOR+1, PATCH=0 | v1.0.5 → v1.1.0 |
| Breaking change | MAJOR+1, others=0 | v1.5.2 → v2.0.0 |

## Git Tags

```bash
# List all versions (newest first)
git tag -l --sort=-version:refname

# See specific tag details
git show v1.0.2
git show latest

# Check which commits are in a tag
git log v1.0.1..v1.0.2 --oneline

# See what's unreleased
git log latest..HEAD --oneline
```

## Common Issues

### "I created the tag but 'latest' isn't updated"

```bash
make latest-tag
```

### "I made a mistake with the version"

```bash
# Delete the tag
git tag -d v1.0.2
git push origin --delete v1.0.2

# Try again
make tag-patch
```

### "The prompts aren't working"

Just do it manually:

```bash
git tag -a v1.0.2 -m "Release v1.0.2"
git push origin v1.0.2
git tag -f latest v1.0.2
git push -f origin latest
```

## Real-World Examples

### Monday: Fix Critical Bug

```bash
# Monday morning: production issue found
make log
# Shows: commit abc123 fix: critical security issue

make tag-patch
# Creates v1.0.0 → v1.0.1 (patch)

make latest-tag
# Users download latest, get the fix
```

### Friday: Ship New Features

```bash
# This week built new features
make log
# Shows: 5 commits with feat: prefix

make tag-minor
# Creates v1.0.1 → v1.1.0 (minor)

# Update docs/blog/changelog
# Announce v1.1.0
```

### Quarterly: Major Update

```bash
# Redesigned entire system
make log
# Shows: BREAKING changes and new APIs

make tag-major
# Creates v1.5.3 → v2.0.0 (major)

# Write migration guide
# Announce v2.0.0 with breaking changes
```

## Pro Tips

1. **Always run `make log` before deciding the release type** - It shows you exactly what changed

2. **Use semantic versioning** - v1.2.3 tells users what kind of change this is

3. **Tag when merging to main** - Make a habit of tagging every merge

4. **Update 'latest' after each tag** - Let users know you have a new stable version

5. **Write good commit messages** - Your future self will thank you when reading the log

6. **Keep CHANGELOG.md updated** - Helps users understand what changed

7. **Test before tagging** - The tag is permanent!

## Integration with GitHub

When you push tags:

```bash
git push origin v1.0.2      # Push specific tag
git push origin --tags      # Push all tags
```

GitHub will:
- Show releases on the "Releases" page
- Create downloadable archives (.zip, .tar.gz)
- Show tag information in commit history

## See Also

- [RELEASE.md](./RELEASE.md) - Full release documentation
- [CHANGELOG.md](./CHANGELOG.md) - Changelog
- [semver.org](https://semver.org/) - Semantic versioning spec
