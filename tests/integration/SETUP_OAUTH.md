# Setting Up Google OAuth for Real Tests

This guide shows how to set up Google OAuth credentials to run real import tests with your Gmail account (gedw99@gmail.com).

## Prerequisites

- Google Cloud account
- Access to Google Cloud Console
- gedw99@gmail.com account

## Step 1: Create Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
   - Project name: `timelinize-plug-testing` (or your choice)
   - Note the Project ID

## Step 2: Enable Required APIs

Enable these APIs for your project:

1. Go to **APIs & Services** → **Library**

2. Search and enable:
   - **Google Photos Library API**
   - **Gmail API**

## Step 3: Create OAuth 2.0 Credentials

1. Go to **APIs & Services** → **Credentials**

2. Click **+ CREATE CREDENTIALS** → **OAuth client ID**

3. If prompted, configure OAuth consent screen:
   - User Type: **External** (for testing)
   - App name: `Timelinize Plug Test`
   - User support email: `gedw99@gmail.com`
   - Developer contact: `gedw99@gmail.com`
   - Scopes: Add these manually later
   - Test users: Add `gedw99@gmail.com`

4. Create OAuth Client ID:
   - Application type: **Desktop app**
   - Name: `timelinize-plug-desktop`

5. Click **CREATE**

6. **IMPORTANT**: Copy your credentials:
   ```
   Client ID: xxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxx.apps.googleusercontent.com
   Client secret: GOCSPX-xxxxxxxxxxxxxxxxxxxxxxxxxxxx
   ```

## Step 4: Configure OAuth Consent Screen Scopes

1. Go to **APIs & Services** → **OAuth consent screen**

2. Click **EDIT APP**

3. In **Scopes**, click **ADD OR REMOVE SCOPES**

4. Add these scopes:
   ```
   https://www.googleapis.com/auth/photoslibrary.readonly
   https://www.googleapis.com/auth/gmail.readonly
   ```

5. Save and continue

6. In **Test users**, add:
   - `gedw99@gmail.com`

## Step 5: Set Environment Variables

Export your credentials in your shell:

```bash
export GOOGLE_CLIENT_ID='your-client-id.apps.googleusercontent.com'
export GOOGLE_CLIENT_SECRET='GOCSPX-your-client-secret'
```

Or add to your `~/.zshrc` / `~/.bashrc`:

```bash
# Google OAuth - timelinize-plug testing
export GOOGLE_CLIENT_ID='your-client-id.apps.googleusercontent.com'
export GOOGLE_CLIENT_SECRET='GOCSPX-your-client-secret'
```

Then reload:
```bash
source ~/.zshrc  # or ~/.bashrc
```

## Step 6: Verify Setup

```bash
# Check credentials are set
echo $GOOGLE_CLIENT_ID
echo $GOOGLE_CLIENT_SECRET

# Should output your credentials (not empty)
```

## Step 7: Run the Real Test

```bash
cd /Users/apple/workspace/go/src/github.com/joeblew999/timelinize-plug

E2E_REAL=1 go test ./tests/integration -run TestGoogleOAuthFlow -v
```

The test will:
1. Check for OAuth credentials
2. Build tplug
3. Start the server
4. Open browser for OAuth authentication
5. Let you choose what to import (Photos/Gmail)
6. Run the import with live progress tracking
7. Show results

## Troubleshooting

### "GOOGLE_CLIENT_ID not set"

Make sure you exported the environment variables:
```bash
export GOOGLE_CLIENT_ID='...'
export GOOGLE_CLIENT_SECRET='...'
```

### "Access blocked: This app's request is invalid"

- Make sure you added `gedw99@gmail.com` as a test user
- Check that the OAuth consent screen is configured
- Verify redirect URI is: `http://127.0.0.1:8008/oauth2-redirect`

### "Error: redirect_uri_mismatch"

Add the redirect URI to your OAuth client:
1. Go to **Credentials** → Your OAuth 2.0 Client
2. Under **Authorized redirect URIs**, add:
   ```
   http://127.0.0.1:8008/oauth2-redirect
   ```

### "Access denied: The developer hasn't given you access to this app"

Add yourself as a test user:
1. **OAuth consent screen** → **Test users**
2. Add `gedw99@gmail.com`

### "This app is blocked"

Your app is in testing mode and you're not a test user:
1. Add yourself as test user (above)
2. OR publish the app (not recommended for testing)

## Security Notes

- ⚠️ **Never commit OAuth credentials to git**
- Store them in environment variables only
- The credentials are for testing only
- Tokens are encrypted and stored in PocketBase

## Alternative: Credentials File

You can also create a credentials file:

```bash
# Create credentials directory
mkdir -p ~/.tplug

# Create credentials file
cat > ~/.tplug/credentials.yaml <<EOF
oauth2_clients:
  google:
    client_id: "your-client-id.apps.googleusercontent.com"
    client_secret: "GOCSPX-your-client-secret"
    redirect_url: "http://127.0.0.1:8008/oauth2-redirect"
EOF

# Secure the file
chmod 600 ~/.tplug/credentials.yaml
```

The app will read from this file if environment variables aren't set.

## Next Steps

After setting up OAuth:

1. Run the guided test:
   ```bash
   E2E_REAL=1 go test ./tests/integration -run TestGoogleOAuthFlow -v
   ```
2. Watch live progress at: http://127.0.0.1:12002
3. Check imported data in `./_timeline_gedw99/`
4. View server logs: `tail -f server.log`

## Reference Links

- [Google Cloud Console](https://console.cloud.google.com/)
- [OAuth 2.0 Playground](https://developers.google.com/oauthplayground/)
- [Photos Library API](https://developers.google.com/photos/library/guides/get-started)
- [Gmail API](https://developers.google.com/gmail/api/guides)
