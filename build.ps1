# Build script for go-mcp-atlassian (Windows PowerShell)
# Creates binaries for multiple platforms

$ErrorActionPreference = "Stop"

$APP_NAME = "go-mcp-atlassian"
$VERSION = if ($env:VERSION) { $env:VERSION } else { "dev" }

# Clean previous builds
if (Test-Path "dist") {
    Remove-Item -Recurse -Force "dist"
}
New-Item -ItemType Directory -Path "dist" | Out-Null

Write-Host "Building $APP_NAME version $VERSION..."

# Build for each platform
$platforms = @(
    @{ GOOS = "darwin"; GOARCH = "amd64" },
    @{ GOOS = "darwin"; GOARCH = "arm64" },
    @{ GOOS = "linux"; GOARCH = "amd64" },
    @{ GOOS = "linux"; GOARCH = "arm64" },
    @{ GOOS = "windows"; GOARCH = "amd64" }
)

foreach ($platform in $platforms) {
    $GOOS = $platform.GOOS
    $GOARCH = $platform.GOARCH
    $OUTPUT = "dist\${APP_NAME}-${GOOS}-${GOARCH}"

    if ($GOOS -eq "windows") {
        $OUTPUT = "${OUTPUT}.exe"
    }

    Write-Host "Building for $GOOS/$GOARCH..."
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    go build -ldflags="-s -w -X main.Version=$VERSION" -o $OUTPUT .
}

# Reset environment
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue

# Generate checksums
Write-Host "Generating checksums..."
Push-Location dist
Get-ChildItem -File | ForEach-Object {
    $hash = (Get-FileHash $_.Name -Algorithm SHA256).Hash.ToLower()
    "$hash  $($_.Name)"
} | Out-File -FilePath "checksums.txt" -Encoding ASCII
Pop-Location

Write-Host "Build complete! Binaries are in the dist/ directory."
Get-ChildItem dist
