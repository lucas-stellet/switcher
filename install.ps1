# switcher installer for Windows
# Usage: irm https://raw.githubusercontent.com/lucas-stellet/switcher/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo = "lucas-stellet/switcher"
$BinaryName = "switcher.exe"

function Get-Arch {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { throw "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
    }
}

function Get-LatestVersion {
    $release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
    return $release.tag_name
}

function Main {
    $arch = Get-Arch
    $version = Get-LatestVersion
    $versionNum = $version.TrimStart("v")

    $filename = "switcher_${versionNum}_windows_${arch}.zip"
    $url = "https://github.com/$Repo/releases/download/$version/$filename"

    # Install to user's local bin directory
    $installDir = "$env:LOCALAPPDATA\Programs\switcher"
    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }

    Write-Host "Installing switcher $version (windows/$arch)..."
    Write-Host "Downloading $url..."

    $tmpFile = Join-Path $env:TEMP $filename
    Invoke-WebRequest -Uri $url -OutFile $tmpFile

    Expand-Archive -Path $tmpFile -DestinationPath $installDir -Force
    Remove-Item $tmpFile

    # Add to PATH if not already there
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$installDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
        $env:Path = "$env:Path;$installDir"
        Write-Host ""
        Write-Host "Added $installDir to your PATH."
    }

    Write-Host ""
    Write-Host "Successfully installed switcher $version to $installDir\$BinaryName"
    Write-Host ""
    Write-Host "Restart your terminal, then run 'switcher --help' to get started."
}

Main
