# FreeBet.guru Android APK Build Script (Windows)

param(
    [switch]$SkipDependencies,
    [switch]$SkipCacheClear
)

$ErrorActionPreference = "Stop"

Write-Host "FreeBet.guru Android APK Build Script (Windows)" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Green
Write-Host "Time: $(Get-Date)" -ForegroundColor Cyan
Write-Host ""

# Load environment variables from .env.local file
if (Test-Path ".env.local") {
    Write-Host "Loading variables from .env.local..." -ForegroundColor Blue
    Get-Content ".env.local" | Where-Object { $_ -notmatch "^#" -and $_ -match "=" } | ForEach-Object {
        $key, $value = $_ -split "=", 2
        [Environment]::SetEnvironmentVariable($key.Trim(), $value.Trim(), "Process")
    }
}

# Check if EXPO_TOKEN is set
if (-not $env:EXPO_TOKEN) {
    Write-Host "ERROR: EXPO_TOKEN not set!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Solution:" -ForegroundColor Cyan
    Write-Host "   Set EXPO_TOKEN in .env.local file" -ForegroundColor White
    Write-Host "   Get token from: https://expo.dev/settings/access-tokens" -ForegroundColor White
    exit 1
}

Write-Host "EXPO_TOKEN found" -ForegroundColor Green

# Check if eas-cli is available
try {
    $easVersion = eas --version 2>$null
    if ($easVersion) {
        Write-Host "EAS CLI found: $easVersion" -ForegroundColor Green
    } else {
        throw "EAS CLI not found"
    }
} catch {
    Write-Host "Installing EAS CLI..." -ForegroundColor Blue
    npm install -g eas-cli --unsafe-perm=true
    Write-Host "EAS CLI installed" -ForegroundColor Green
}

# Check authentication
$whoami = eas whoami 2>$null
if ($LASTEXITCODE -eq 0 -and $whoami) {
    Write-Host "Authenticated to EAS: $whoami" -ForegroundColor Green
} else {
    Write-Host "Logging into EAS..." -ForegroundColor Blue
    eas login
    if ($LASTEXITCODE -ne 0) {
        Write-Host "ERROR: EAS login failed" -ForegroundColor Red
        exit 1
    }
}

# Check if eas.json exists
if (-not (Test-Path "eas.json")) {
    Write-Host "ERROR: eas.json not found!" -ForegroundColor Red
    Write-Host "Run: eas init" -ForegroundColor White
    exit 1
}

Write-Host "EAS project configured" -ForegroundColor Green

# Check dependencies
Write-Host "Checking dependencies..." -ForegroundColor Blue
if (-not (Test-Path "node_modules")) {
    Write-Host "Installing dependencies..." -ForegroundColor Blue
    if (Test-Path "package-lock.json") {
        npm ci
    } else {
        npm install
    }
}

if (-not (Test-Path "node_modules")) {
    Write-Host "ERROR: Dependencies not installed" -ForegroundColor Red
    exit 1
}

Write-Host "Dependencies OK" -ForegroundColor Green

# Install expo-updates if needed
$packageJson = Get-Content "package.json" | ConvertFrom-Json
if (-not $packageJson.dependencies.'expo-updates' -and -not $packageJson.devDependencies.'expo-updates') {
    Write-Host "Installing expo-updates..." -ForegroundColor Blue
    npx expo install expo-updates
}

# Clear cache if requested
if (-not $SkipCacheClear) {
    Write-Host "Clearing cache..." -ForegroundColor Blue

    if (Test-Path ".expo") {
        Remove-Item -Recurse -Force ".expo"
        Write-Host "Cleared .expo directory" -ForegroundColor Blue
    }
    if (Test-Path ".expo-shared") {
        Remove-Item -Recurse -Force ".expo-shared"
        Write-Host "Cleared .expo-shared directory" -ForegroundColor Blue
    }

    if (Test-Path "node_modules") {
        Remove-Item -Recurse -Force "node_modules"
        if (Test-Path "package-lock.json") {
            Remove-Item "package-lock.json"
        }
        Write-Host "Reinstalling dependencies..." -ForegroundColor Blue
        npm install
    }

    Write-Host "Cache cleared" -ForegroundColor Green
}

Write-Host "Starting APK build..." -ForegroundColor Green
Write-Host "This will take several minutes..." -ForegroundColor Yellow
Write-Host ""

$currentDate = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

# Build APK
Write-Host "Building APK..." -ForegroundColor Green
$currentDate = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

# Build APK
# Temporarily disable error handling for Unicode output
$ErrorActionPreference = 'Continue'
try {
    eas build --platform android --profile production --message "Android Production build $currentDate" --wait 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "ERROR: Build failed with exit code $LASTEXITCODE" -ForegroundColor Red
        exit 1
    }
    Write-Host "Build completed!" -ForegroundColor Green
} finally {
    $ErrorActionPreference = 'Stop'
}

# Download APK
Write-Host "Downloading APK..." -ForegroundColor Blue

# Get build list
eas build:list --platform android --limit 1 --json > build_list.json
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to get build list" -ForegroundColor Red
    exit 1
}

$buildInfoJson = Get-Content "build_list.json" -Raw
$buildInfo = $buildInfoJson | ConvertFrom-Json

if ($buildInfo -and $buildInfo.Length -gt 0) {
    $buildId = $buildInfo[0].id
    $apkName = "FreeBet-Remote-$(Get-Date -Format 'yyyyMMdd-HHmmss').apk"

    # Download APK
    eas build:download $buildId --output $apkName
    if ($LASTEXITCODE -ne 0) {
        Write-Host "ERROR: Failed to download APK" -ForegroundColor Red
        exit 1
    }

    # Clean up temp file
    if (Test-Path "build_list.json") {
        Remove-Item "build_list.json"
    }

        if (Test-Path $apkName) {
            $fileSize = (Get-Item $apkName).Length / 1MB
            Write-Host "SUCCESS: APK downloaded!" -ForegroundColor Green
            Write-Host "File: $(Get-Location)\$apkName" -ForegroundColor White
            Write-Host "Size: $([math]::Round($fileSize, 2)) MB" -ForegroundColor White
            Write-Host ""
            Write-Host "APK ready for distribution!" -ForegroundColor Green
        } else {
            Write-Host "ERROR: APK download failed" -ForegroundColor Red
            exit 1
        }
    } else {
        Write-Host "ERROR: No build found" -ForegroundColor Red
        exit 1
    }
