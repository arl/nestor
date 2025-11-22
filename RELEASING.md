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

## Supported Platforms

- **Linux**: amd64 only
- **macOS**: amd64 (Intel) and arm64 (Apple Silicon)

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
   - Permission issues with Homebrew tap (check GITHUB_TOKEN permissions)
