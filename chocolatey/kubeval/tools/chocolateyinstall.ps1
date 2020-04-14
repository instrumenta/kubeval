
$ErrorActionPreference = 'Stop'

$packageName= $env:ChocolateyPackageName
$toolsDir   = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$url        = "https://github.com/instrumenta/kubeval/releases/download/$($env:ChocolateyPackageVersion)/kubeval-windows-386.zip"
$url64      = "https://github.com/instrumenta/kubeval/releases/download/$($env:ChocolateyPackageVersion)/kubeval-windows-amd64.zip"

$packageArgs = @{
  packageName   = $packageName
  unzipLocation = $toolsDir
  url           = $url
  url64bit      = $url64

  checksum      = '5DED35273DD35993C0FC52A08D9CC268487620736C4782077BC72723CC7224D0'
  checksumType  = 'sha256'
  checksum64    = '2A844518981848A7D77CCED9B51A05174BA9C17FC007A1C48CD2AF0D3FB021D7'
  checksumType64= 'sha256'
}

Install-ChocolateyZipPackage @packageArgs
