param(
    [Parameter(Mandatory)][string]$Notes,
    [string]$Key = (Join-Path ([Environment]::GetFolderPath('UserProfile')) '.toque\update-signing.key')
)
$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

$version = [regex]::Match((Get-Content internal\app\service.go -Raw), 'const Version = "([^"]+)"').Groups[1].Value
if (-not $version) { throw 'No encuentro la versión en internal\app\service.go' }
$tag = "v$version"
if (git status --porcelain) { throw 'Hay cambios sin commit' }
if (git tag --list $tag) { throw "La etiqueta $tag ya existe" }
if (-not (Test-Path $Key)) { throw "Falta la clave de firma en $Key" }

& .\build.ps1
$exe = Join-Path $PSScriptRoot 'build\bin\Toque.exe'
$dist = Join-Path $PSScriptRoot 'dist'
New-Item -ItemType Directory -Force $dist | Out-Null
$manifestPath = Join-Path $dist 'latest.json'
$manifest = [ordered]@{
    version   = $version
    url       = "https://github.com/dimelim/toque-releases/releases/download/$tag/Toque.exe"
    sha256    = (Get-FileHash $exe -Algorithm SHA256).Hash.ToLower()
    notes     = $Notes
    published = (Get-Date).ToUniversalTime().ToString('o')
}
[IO.File]::WriteAllText($manifestPath, ($manifest | ConvertTo-Json -Compress), (New-Object Text.UTF8Encoding $false))
go run ./cmd/toque-sign sign $Key $manifestPath
if ($LASTEXITCODE -ne 0) { throw 'No se pudo firmar el manifiesto' }

gh release create $tag --repo dimelim/toque-releases --title "Toque $version" --notes $Notes $exe $manifestPath "$manifestPath.sig"
if ($LASTEXITCODE -ne 0) { throw 'No se pudo publicar en toque-releases' }
git tag $tag
git push origin $tag
gh release create $tag --repo dimelim/toque --title "Toque $version" --notes $Notes $exe
if ($LASTEXITCODE -ne 0) { throw 'No se pudo crear el lanzamiento en dimelim/toque' }
Write-Host "== Toque $version publicada"
