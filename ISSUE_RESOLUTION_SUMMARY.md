# Issue Resolution Summary: TypeMetadata Not Passed to OpenAI API

## The Problem

On Fly.io, the digest application receives a 400 error from OpenAI API:

```
Invalid schema for response_format 'aiwf_output':
In context=(), 'additionalProperties' is required to be supplied and to be false.
```

The OpenAI logs show a minimal schema with empty properties instead of the full schema.

## Root Cause Identified

The issue is **NOT** in the AIWF framework code. The generated SDK files are **correct**:

✅ `sdk/service.go` - Contains TypeProvider injection: `agent.Types = s`
✅ `sdk/agents.go` - Contains OutputTypeName in config: `"DigestOutput"`
✅ `sdk/types.go` - Contains complete TypeMetadata with `additionalProperties: false`

The issue is that **the digest project is using OLD SDK files** that don't have TypeProvider injection.

### Why This Happens

1. User copied SDK files from aiwf to digest project: ✅ Correct approach for the time
2. Later, AIWF was updated to add TypeProvider injection pattern
3. Digest project's SDK files were NOT regenerated
4. Result: digest project uses OLD SDK → a.Types is nil → TypeMetadata is nil → error

## What We've Done

### 1. Added Debugging Logs

**File**: `runtime/go/aiwf/sdk_helpers.go` (lines 47-60)

When the code runs, it now prints:
```
[DEBUG] CallModel: Agent=digest_assistant, Types=false, OutputTypeName=DigestOutput
[DEBUG] CallModel: No TypeProvider or OutputTypeNameempty
```

This will immediately show whether TypeProvider was injected.

**File**: `providers/openai/provider.go` (lines 270, 277)

When building the schema, it now prints:
```
openai: buildJSONSchemaFormat - WARNING: TypeMetadata is nil, using minimal schema
```

This shows when the fallback minimal schema is being used.

### 2. Created Documentation

- **DEBUGGING_TYPEMETADATA_ISSUE.md** - Detailed guide on how to debug the issue
- **TYPEMETADATA_FLOW_DIAGRAM.md** - Visual flowchart of the entire data flow

### 3. Verified Generated Code

Confirmed that ALL generated SDKs in the framework contain:
- ✅ TypeProvider injection in service.go
- ✅ OutputTypeName in agent configs
- ✅ Complete TypeMetadata map in types.go

## What User Needs to Do

### Step 1: Deploy Latest AIWF

The digest project needs to use the latest version of AIWF that includes the debugging logs.

```bash
go get -u github.com/andranikuz/aiwf@latest
```

### Step 2: Regenerate SDK in Digest Project

Run the AIWF generator on the digest YAML template:

```bash
cd digest_project
aiwf generate ./aiwf-digest.yaml --output ./pkg/sdk
```

This will regenerate the SDK files with:
- TypeProvider injection in service.go ✅
- Complete TypeMetadata in types.go ✅

### Step 3: Verify the Generated Files

Check that `pkg/sdk/service.go` contains:

```go
func NewService(client aiwf.ModelClient) *Service {
    s := &Service{client: client}

    agent := NewDigestAssistantAgent(client)
    agent.Types = s  // ← This line must be present

    s.agents = &Agents{
        DigestAssistant: agent,
    }
    return s
}
```

### Step 4: Deploy Updated Digest Project

Deploy the digest project with updated SDK files.

### Step 5: Check Logs on Fly.io

Look for debug output showing:
```
[DEBUG] CallModel: Agent=digest_assistant, Types=true, OutputTypeName=DigestOutput
[DEBUG] CallModel: Got TypeMetadata for DigestOutput
```

If you see this, TypeMetadata is being properly retrieved and used!

## Expected Result

After deployment with updated SDK:

1. OpenAI receives full schema with `additionalProperties: false` at all levels
2. OpenAI API accepts the request and returns ✅ 200 OK
3. Digest generation works correctly

## Prevention for Future

### Best Practice

**Each project should:**
1. Have its own `.yaml` template in its repository
2. Generate SDK locally using AIWF CLI
3. Commit ONLY the generated SDK files to that project's repository

**Never:**
- Copy SDK files between projects (versions can become mismatched)
- Manually edit generated code (edit the generator instead, then regenerate)
- Share a single SDK folder between projects

### Why This Matters

- AIWF improvements (like TypeProvider injection) are automatically reflected when you regenerate
- Each project stays in sync with the AIWF framework version it depends on
- No manual coordination needed between projects

## Technical Details for Reference

### The Flow (Simplified)

```
SDK Service.NewService()
  ↓ Injects itself as TypeProvider into agents
  ↓ a.Types = s
  ↓
Agent.CallModel()
  ↓ Checks if a.Types != nil
  ↓ Retrieves TypeMetadata: a.Types.GetTypeMetadata(OutputTypeName)
  ↓
OpenAI Provider.buildJSONSchemaFormat()
  ↓ If TypeMetadata != nil, uses full schema
  ↓ If TypeMetadata == nil, uses minimal schema (CAUSES ERROR)
  ↓
OpenAI API
  ↓ Validates schema
  ✓ Success if additionalProperties: false present everywhere
  ✗ Error if additionalProperties missing
```

### Why "additionalProperties: false" is Required

OpenAI Responses API requires strict JSON schema validation. The `additionalProperties: false` setting means:
- Objects can ONLY have the properties defined in the schema
- No extra properties are allowed
- This ensures the LLM returns exactly the structure specified

## Questions to Ask When Debugging

If the issue persists after updating:

1. **Is the debug output showing Types=false?**
   - If yes, TypeProvider injection didn't happen
   - Check that service.go was regenerated

2. **Is the debug output showing GetTypeMetadata error?**
   - If yes, TypeMetadata lookup is failing
   - Check that types.go was regenerated

3. **Which version of AIWF is the digest project using?**
   - Might still be on an old version
   - Run `go get -u github.com/andranikuz/aiwf@latest`

4. **Is the digest project's SDK folder in git?**
   - Make sure updated SDK files were actually deployed
   - Check git history to confirm new files are present

## Related Files

- `DEBUGGING_TYPEMETADATA_ISSUE.md` - How to debug the issue
- `TYPEMETADATA_FLOW_DIAGRAM.md` - Visual flow diagrams
- `commit: c1e306b` - Debugging logs implementation
- `commit: 857d91d` - Documentation
