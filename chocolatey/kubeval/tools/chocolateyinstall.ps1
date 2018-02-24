
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

  checksum      = '1EB9F826E6E607B07E38C79AFFF425392016953368158D6B33817311D44F14EF'
  checksumType  = 'sha256'
  checksum64    = '4C7085BA0366F961FC2664DED0F09AE61FB54F393717A73AAC8601987D7BDA31' 
  checksumType64= 'sha256'
}

Install-ChocolateyZipPackage @packageArgs
