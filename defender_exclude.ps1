# defender_exclude.ps1
$installPath = "C:\Program Files\Jarvis"

# Admin tekshir
if (-NOT ([Security.Principal.WindowsPrincipalWindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]"Administrator")) {
    Start-Process PowerShell -Verb RunAs -ArgumentList "-File `"$PSCommandPath`""
    exit
}

# Exclusion qo'sh
Add-MpPreference -ExclusionPath $installPath
Add-MpPreference -ExclusionProcess "jarvis.exe"
Add-MpPreference -ExclusionProcess "Setup.exe"

Write-Host "✅ Windows Defender exclusion qo'shildi!" -ForegroundColor Green
Write-Host "Endi Setup.exe ni ishga tushiring." -ForegroundColor Cyan
pause