# Release Process

## Creating a New Release

To create a new release, simply push a tag with the format `v*.*.*`:

```bash
git tag v0.1.0
git push origin v0.1.0
```

## What Happens

The release workflow (`.github/workflows/release.yml`) will automatically:

1. **Build binaries** for:
   - Linux/amd64
   - macOS/amd64 (Intel)
   - macOS/arm64 (Apple Silicon)

2. **Create a GitHub release** with:
   - All binary archives (.tar.gz files)
   - Individual checksum files for each platform
   - Auto-generated release notes from commits

3. **Update Homebrew tap** at `arl/homebrew-arl`:
   - Creates/updates the Nestor formula
   - Supports all three platforms
   - Users can install via: `brew install arl/arl/nestor`
   - **Note**: On macOS, the binary is unsigned. If you encounter quarantine issues, the formula will display instructions on how to remove the quarantine attribute.

## Supported Platforms

- **Linux**: amd64 only
- **macOS**: amd64 (Intel) and arm64 (Apple Silicon)
  - **Important**: Binaries are not signed. macOS users may need to remove the quarantine attribute using:
    ```bash
    xattr -d com.apple.quarantine $(which nestor)
    ```

## Release Naming

- Archives: `nestor_<version>_<os>_<arch>.tar.gz`
- Checksums: `nestor_<version>_checksums_<os>_<arch>.txt`

Example:
- `nestor_v0.1.0_darwin_arm64.tar.gz`
- `nestor_v0.1.0_checksums_darwin_arm64.txt`

## Testing Releases

For testing, create tags starting from `v0.1.0` onwards. Do not modify releases or tags before v0.1.0.

## Troubleshooting

If a release fails:
1. Check the Actions tab in GitHub
2. Review the build logs for the specific platform that failed
3. Common issues:
   - Missing dependencies (ensure all CGO dependencies are installed)
   - Network issues downloading dependencies
   - **Homebrew tap permission error**: The GITHUB_TOKEN may not have permission to push to the homebrew-arl repository. To fix:
     - Create a Personal Access Token (classic) with `repo` scope
     - Add it to repository secrets as `HOMEBREW_TAP_TOKEN`
     - Re-run the failed workflow

## Setup Requirements

For the Homebrew tap update to work, you need to:
1. Create a Personal Access Token (PAT) in GitHub settings
2. Give it the `repo` scope (full control of private repositories)
3. Add it to this repository's secrets as `HOMEBREW_TAP_TOKEN`

Without this token, the release will complete successfully, but the Homebrew formula won't be updated automatically.
