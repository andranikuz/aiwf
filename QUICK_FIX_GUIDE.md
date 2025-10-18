# Quick Fix Guide: TypeMetadata Error on Fly.io

## The Error
```
Invalid schema for response_format 'aiwf_output':
'additionalProperties' is required to be supplied and to be false.
```

## The Cause (in 1 sentence)
digest project is using OLD SDK files that don't inject TypeProvider into agents.

## The Fix (5 steps)

### Step 1: Update AIWF dependency
```bash
cd digest_project
go get -u github.com/andranikuz/aiwf@latest
```

### Step 2: Regenerate SDK
```bash
aiwf generate ./aiwf-digest.yaml --output ./pkg/sdk
```

### Step 3: Verify the fix
Open `pkg/sdk/service.go` and look for:
```go
agent.Types = s  // ← This line MUST be present
```

If you see it, ✅ the fix worked.

### Step 4: Commit changes
```bash
git add pkg/sdk/
git commit -m "chore: regenerate SDK with latest AIWF"
git push
```

### Step 5: Deploy to Fly.io
```bash
fly deploy
```

## How to Verify It Works

Check Fly.io logs:
```bash
fly logs
```

Look for:
```
[DEBUG] CallModel: Agent=digest_assistant, Types=true, OutputTypeName=DigestOutput
[DEBUG] CallModel: Got TypeMetadata for DigestOutput
```

If you see `Types=true` and no "WARNING: TypeMetadata is nil" message, you're ✅ fixed!

## Still Broken?

1. **Check the commit**: `git log pkg/sdk/service.go` - was it actually regenerated?
2. **Check AIWF version**: `go list -m github.com/andranikuz/aiwf` - is it latest?
3. **Check the code**: Does regenerated `service.go` have `agent.Types = s`?
4. **Check deployment**: Did new files get deployed to Fly.io? Check git on the server.

## Why This Happens

When AIWF is updated with TypeProvider injection pattern, old SDK files don't have this code. The digest project needs to regenerate its SDK to get the new code.

```
OLD SDK:              NEW SDK:
(no injection)        (with injection)
❌ agent.Types        ✅ agent.Types = s
  not set
```

## One-Liner
After pulling latest aiwf, just run: `aiwf generate ./aiwf-digest.yaml --output ./pkg/sdk`
