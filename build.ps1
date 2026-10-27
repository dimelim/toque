param([switch]$Obfuscate, [string]$PresentMon = "$PSScriptRoot\.dev\bin\PresentMon.exe")
$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot
$payload = Join-Path $PSScriptRoot 'internal\install\payload'
New-Item -ItemType Directory -Force $payload | Out-Null
$ldflags = '-s -w -H windowsgui -buildid='
$env:CGO_ENABLED = '0'

Write-Host '== agent'
if ($Obfuscate) {
    garble -literals -tiny -seed=random build -trimpath -ldflags $ldflags -o "$payload\ToqueAgent.exe" ./cmd/toque-agent
} else {
    go build -trimpath -ldflags $ldflags -o "$payload\ToqueAgent.exe" ./cmd/toque-agent
}
if ($LASTEXITCODE -ne 0) { throw 'agent build failed' }

Write-Host '== presentmon'
$sig = Get-AuthenticodeSignature $PresentMon
if ($sig.Status -ne 'Valid' -or $sig.SignerCertificate.Subject -notmatch 'Intel Corporation') { throw 'PresentMon signature check failed' }
Copy-Item $PresentMon "$payload\PresentMon.exe" -Force

Write-Host '== app'
$wailsArgs = @('build', '-clean', '-trimpath', '-ldflags', '-s -w -buildid=', '-webview2', 'download')
if ($Obfuscate) { $wailsArgs += @('-obfuscated', '-garbleargs', '-literals -tiny -seed=random') }
& wails @wailsArgs
if ($LASTEXITCODE -ne 0) { throw 'wails build failed' }

$exe = Join-Path $PSScriptRoot 'build\bin\Toque.exe'
$hash = (Get-FileHash $exe -Algorithm SHA256).Hash
Write-Host ("== done {0} {1:N1} MB sha256={2}" -f $exe, ((Get-Item $exe).Length / 1MB), $hash)
