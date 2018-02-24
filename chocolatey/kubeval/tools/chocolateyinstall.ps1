
$ErrorActionPreference = 'Stop'

$packageName= $env:ChocolateyPackageName
$toolsDir   = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$url        = "https://github.com/garethr/kubeval/releases/download/$($env:ChocolateyPackageVersion)/kubeval-windows-386.zip"
$url64      = "https://github.com/garethr/kubeval/releases/download/$($env:ChocolateyPackageVersion)/kubeval-windows-amd64.zip"

$packageArgs = @{
  packageName   = $packageName
  unzipLocation = $toolsDir
  url           = $url
  url64bit      = $url64

  checksum      = 'CA8CAAB937525275EA503E67B960DBE82D4D39694B03E7F7244B6B4795555BD0'
  checksumType  = 'sha256'
  checksum64    = '4EFEF42007D020CD8CCB54F52744DDB51F5336F52F8711B29F5884A9DD267952'
  checksumType64= 'sha256'
}

Install-ChocolateyZipPackage @packageArgs
