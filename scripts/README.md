# Scripts

Utility scripts for PixTui development.

## Available Scripts

| Script | Description | Usage |
|--------|-------------|-------|
| `build.sh` | Build all packages and demo binary | `./scripts/build.sh [output-dir]` |
| `clean.sh` | Remove build artifacts and caches | `./scripts/clean.sh` |
| `test.sh` | Run all tests with coverage | `./scripts/test.sh` |
| `release.sh` | Create a new release | `./scripts/release.sh <version>` |

## Examples

```bash
# Build the project
./scripts/build.sh

# Run tests
./scripts/test.sh

# Create a release
./scripts/release.sh v1.0.0

# Clean up
./scripts/clean.sh
```
