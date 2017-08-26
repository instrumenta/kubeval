
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

  checksum      = '4532878B3D12B9A38B2CB3674842427E436E90E407D6CE6E2AE4CE18FEE70DD6'
  checksumType  = 'sha256'
  checksum64    = '2261E82FB2032C77D5ADB08E830A17F20357F1203FA4D267BA1795B0ECD1DC5D'
  checksumType64= 'sha256'
}

Install-ChocolateyZipPackage @packageArgs
