$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$repo = "nikhil25803/wordly"
$installDir = if ($env:WORDLY_INSTALL_DIR) { $env:WORDLY_INSTALL_DIR } else { Join-Path $HOME ".local\bin" }
$architecture = [Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
$arch = switch ($architecture) {
    "x64" { "amd64" }
    "arm64" { "arm64" }
    default { throw "Unsupported architecture: $architecture" }
}

$archive = "wordly_windows_$arch.zip"
$baseUrl = if ($env:WORDLY_RELEASE_URL) { $env:WORDLY_RELEASE_URL.TrimEnd("/") } else { "https://github.com/$repo/releases/latest/download" }
$tempDir = Join-Path ([IO.Path]::GetTempPath()) ("wordly-" + [guid]::NewGuid())
$archivePath = Join-Path $tempDir $archive
$checksumsPath = Join-Path $tempDir "checksums.txt"

New-Item -ItemType Directory -Path $tempDir | Out-Null
try {
    & curl.exe -fsSL --retry 3 --retry-delay 1 --connect-timeout 10 `
        -o $archivePath "$baseUrl/$archive" `
        -o $checksumsPath "$baseUrl/checksums.txt"
    if ($LASTEXITCODE -ne 0) {
        throw "Download failed with curl exit code $LASTEXITCODE"
    }

    $checksumLine = Get-Content $checksumsPath | Where-Object { ($_ -split "\s+")[-1] -eq $archive } | Select-Object -First 1
    if (-not $checksumLine) {
        throw "No checksum found for $archive"
    }
    $expected = ($checksumLine -split "\s+")[0].ToLowerInvariant()
    $actual = (Get-FileHash -Algorithm SHA256 $archivePath).Hash.ToLowerInvariant()
    if ($actual -ne $expected) {
        throw "Checksum verification failed"
    }

    Expand-Archive -Path $archivePath -DestinationPath $tempDir -Force
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    Copy-Item (Join-Path $tempDir "wordly.exe") (Join-Path $installDir "wordly.exe") -Force
    Write-Output "Installed wordly.exe to $installDir"
} finally {
    Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
}
