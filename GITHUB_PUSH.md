# GitHub Authentication & Push Guide

## 🔐 Setting Up Authentication

Since GitHub no longer supports password authentication, you need to use either a Personal Access Token (PAT) or SSH keys.

### Option 1: Personal Access Token (Recommended)

1. **Create a Personal Access Token:**
   - Go to GitHub.com → Settings → Developer settings → Personal access tokens → Tokens (classic)
   - Click "Generate new token (classic)"
   - Select scopes: `repo` (full repository access)
   - Copy the token (save it securely - you won't see it again!)

2. **Configure Git with Token:**
   ```bash
   # Set your GitHub username
   git config --global user.name "ahur-system"
   git config --global user.email "your-email@example.com"
   
   # When prompted for password, use your PAT instead
   ```

3. **Push to GitHub:**
   ```bash
   cd /root/projects/sysmedic
   git remote set-url origin https://github.com/ahur-system/sysmedic.git
   git push -u origin main
   git push origin --tags
   ```

### Option 2: SSH Keys (Alternative)

1. **Generate SSH Key:**
   ```bash
   ssh-keygen -t ed25519 -C "your-email@example.com"
   # Press Enter for default location
   # Set a passphrase (recommended)
   ```

2. **Add SSH Key to GitHub:**
   ```bash
   # Copy public key
   cat ~/.ssh/id_ed25519.pub
   
   # Go to GitHub.com → Settings → SSH and GPG keys → New SSH key
   # Paste the public key content
   ```

3. **Use SSH Remote:**
   ```bash
   cd /root/projects/sysmedic
   git remote set-url origin git@github.com:ahur-system/sysmedic.git
   git push -u origin main
   git push origin --tags
   ```

## 🚀 Push Commands

Once authentication is set up:

```bash
cd /root/projects/sysmedic

# Push main branch
git push -u origin main

# Push all tags
git push origin --tags

# Verify everything was pushed
git log --oneline
git tag
```

## 📦 Upload Release Assets

After pushing the code, upload the release package:

1. **Go to GitHub Releases:**
   - Visit: https://github.com/ahur-system/sysmedic/releases
   - Click "Create a new release"

2. **Create Release:**
   - Tag version: `v1.0.0`
   - Release title: `🚀 SysMedic v1.0.0 - Production Release`
   - Description: Use content from `RELEASE.md`

3. **Upload Assets:**
   ```bash
   # The release package is ready at:
   ls -la /root/projects/sysmedic/dist/sysmedic-v1.0.0-linux-amd64.tar.gz
   
   # Size: ~3.1MB
   # Upload this file as a release asset
   ```

## 🔍 Verification

After pushing, verify your repository:

1. **Check Repository:**
   - Visit: https://github.com/ahur-system/sysmedic
   - Verify all files are present
   - Check that README.md displays correctly

2. **Verify Release:**
   - Check tags are visible
   - Ensure release assets can be downloaded
   - Test clone and build:
   ```bash
   cd /tmp
   git clone https://github.com/ahur-system/sysmedic.git
   cd sysmedic
   make build
   ./build/sysmedic --version
   ```

## 🎯 Next Steps After Publishing

1. **Enable Repository Features:**
   - Issues (for bug reports)
   - Discussions (for community support)
   - Wiki (for extended documentation)

2. **Add Repository Topics:**
   ```
   monitoring, system-monitoring, linux, golang, cli, daemon,
   systemd, sqlite, user-tracking, server-monitoring, sysadmin,
   devops, performance, alerting, email-notifications
   ```

3. **Set Repository Description:**
   ```
   Cross-platform Linux server monitoring CLI tool with user-centric resource tracking and persistent usage detection
   ```

## 🛠️ Troubleshooting

**Authentication Failed:**
- Ensure PAT has correct permissions
- Check username is correct
- Verify token hasn't expired

**Permission Denied:**
- Check SSH key is added to GitHub
- Verify SSH agent is running: `ssh-add -l`

**Push Rejected:**
- Repository might not be empty
- Check if you need to pull first: `git pull origin main`

Your SysMedic repository is ready for the world! 🌟