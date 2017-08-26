$ErrorActionPreference = 'Stop';

$kubevalExe = Get-ChildItem $(Split-Path -Parent $MyInvocation.MyCommand.Definition) | Where-Object -Property Name -Match "kubeval.exe"

if (-Not($kubevalExe)) 
{
    Write-Error -Message "kubeval.exe not found, please contact the maintainer of the package" -Category ResourceUnavailable
}

Write-Host "found kubeval.exe in $($kubevalExe.FullName)"
Write-Host "attempting to remove it" 
Remove-Item $kubevalExe.FullName
