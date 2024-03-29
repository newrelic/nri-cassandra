<#
    .SYNOPSIS
        This script creates the win .MSI
#>
param (
    # Target architecture: amd64 (default) or 386
    [string]$integration="none",
    [ValidateSet("amd64", "386")]
    [string]$arch="amd64",
    [string]$tag="v0.0.0",
    [string]$pfx_passphrase="none",
    [string]$pfx_certificate_description="none"
)

$buildYear = (Get-Date).Year

$version=$tag.substring(1)

# verifying version number format
$v = $version.Split(".")

if ($v.Length -ne 3) {
    echo "-version must follow a numeric major.minor.patch semantic versioning schema (received: $version)"
    exit -1
}

$wrong = $v | ? { (-Not [System.Int32]::TryParse($_, [ref]0)) -or ( $_.Length -eq 0) -or ([int]$_ -lt 0)} | % { 1 }
if ($wrong.Length  -ne 0) {
    echo "-version major, minor and patch must be valid positive integers (received: $version)"
    exit -1
}

$noSign = $env:NO_SIGN ?? "false"
if ($noSign -ieq "true") {
    echo "===> Import .pfx certificate is disabled by environment variable"
} else {
    echo "===> Import .pfx certificate from GH Secrets"
    Import-PfxCertificate -FilePath wincert.pfx -Password (ConvertTo-SecureString -String $pfx_passphrase -AsPlainText -Force) -CertStoreLocation Cert:\CurrentUser\My

    echo "===> Show certificate installed"
    Get-ChildItem -Path cert:\CurrentUser\My\
}

echo "===> Checking MSBuild.exe..."
$msBuild = (Get-ItemProperty hklm:\software\Microsoft\MSBuild\ToolsVersions\4.0).MSBuildToolsPath
if ($msBuild.Length -eq 0) {
    echo "Can't find MSBuild tool. .NET Framework 4.0.x must be installed"
    exit -1
}
echo $msBuild

echo "===> Building integration"
Push-Location -Path "build\package\windows\nri-$arch-installer"

. $msBuild/MSBuild.exe nri-installer.wixproj /p:IntegrationVersion=${version} /p:IntegrationName=$integration /p:Year=$buildYear /p:NoSign=$noSign /p:pfx_certificate_description=$pfx_certificate_description

if (-not $?)
{
    echo "Failed building installer"
    Pop-Location
    exit -1
}

echo "===> Making versioned installed copy"
cd bin\Release
cp "nri-$integration-$arch.msi" "nri-$integration-nodeps-$arch.$version.msi"

Pop-Location

# Copy integration MSI to bundle dir
cp "build\package\windows\nri-$arch-installer\bin\Release\nri-$integration-$arch.msi" "build\package\windows\bundle"

echo "===> Building nrjmx bundle"
Push-Location -Path "build\package\windows\bundle"

. $msBuild/MSBuild.exe bundle.wixproj /p:IntegrationVersion=${version} /p:IntegrationName=$integration /p:Year=$buildYear /p:NoSign=$noSign /p:pfx_certificate_description=$pfx_certificate_description

if (-not $?)
{
    echo "Failed building bundle"
    Pop-Location
    exit -1
}

echo "===> Making versioned bundle copy"
cp "bin\Release\nri-$integration-bundle-$arch.exe" "bin\Release\nri-$integration-$arch-installer.$version.exe"

Pop-Location
