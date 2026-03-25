# Quick Start: Real Google Import Test

Get up and running with real Google data import in 5 minutes.

## For gedw99@gmail.com

### Step 1: Set OAuth Credentials (One-time)

```bash
# Export your Google OAuth credentials
export GOOGLE_CLIENT_ID='your-client-id.apps.googleusercontent.com'
export GOOGLE_CLIENT_SECRET='GOCSPX-your-secret'
```

**Don't have credentials yet?** Follow [integration/SETUP_OAUTH.md](integration/SETUP_OAUTH.md) (~10 minutes to set up).

### Step 2: Run the Test

```bash
cd /Users/apple/workspace/go/src/github.com/joeblew999/timelinize-plug

E2E_REAL=1 go test ./tests/integration -run TestGoogleOAuthFlow -v
```

### Step 3: Follow the Prompts

The script will:

1. ✅ **Check credentials** - Verifies OAuth setup
2. ✅ **Build & start server** - Launches NATS + PocketBase + UI
3. 🔐 **Authenticate** - Opens browser for Google OAuth
   - Sign in with `gedw99@gmail.com`
   - Grant Photos + Gmail permissions
4. 📊 **Choose import** - Select what to import:
   - Google Photos (sample)
   - Gmail (sample)
   - Both
5. 📈 **Watch progress** - Real-time updates at http://127.0.0.1:12002

### Step 4: View Results

After import completes:

```bash
# Check timeline data
ls -lh ./_timeline_gedw99/

# View in UI
open http://127.0.0.1:12002?tenant=gedw99

# Check logs
tail -f server.log
```

Press `Ctrl+C` to stop the server.

## What This Tests

- ✅ **OAuth Flow** - Real Google authentication
- ✅ **Token Storage** - Encrypted in PocketBase
- ✅ **Import Pipeline** - Timelinize data extraction
- ✅ **Progress Events** - NATS → SSE → Datastar UI
- ✅ **Multi-tenant** - Tenant isolation (gedw99)
- ✅ **End-to-end** - Full stack integration

## Troubleshooting

### "GOOGLE_CLIENT_ID not set"
```bash
# Make sure you exported the vars
echo $GOOGLE_CLIENT_ID
echo $GOOGLE_CLIENT_SECRET
```

### "Port already in use"
```bash
# Kill old processes
killall tplug
lsof -ti:4222 | xargs kill -9
lsof -ti:12002 | xargs kill -9
```

### "Access blocked"
- Add `gedw99@gmail.com` as test user in Google Cloud Console
- See [SETUP_OAUTH.md](integration/SETUP_OAUTH.md)

### "Build failed - timelinize"
The test works without `-tags timelinize` - it will simulate import with test events.

To build with real timelinize support:
```bash
go build -tags timelinize -o tplug ./cmd/tplug
```

## Next Steps

1. **Try different sources**:
   - Gmail only
   - Photos only
   - Both

2. **Watch live updates**:
   - Open http://127.0.0.1:12002 before running import
   - See progress bars update in real-time

3. **Test multi-tenant**:
   - Change `TENANT_ID` in the script
   - Import to different tenant

4. **Explore the data**:
   - Check `./_timeline_gedw99/` for SQLite database
   - Query with `sqlite3`

## Need Help?

- 📖 Full setup guide: [integration/SETUP_OAUTH.md](integration/SETUP_OAUTH.md)
- 📖 Test docs: [README.md](README.md)
- 📖 Use case docs: [../docs/USE_CASE_IMPORT_PROGRESS.md](../docs/USE_CASE_IMPORT_PROGRESS.md)
