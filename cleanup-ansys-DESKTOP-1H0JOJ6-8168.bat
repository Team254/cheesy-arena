@echo off
set LOCALHOST=%COMPUTERNAME%
if /i "%LOCALHOST%"=="DESKTOP-1H0JOJ6" (taskkill /f /pid 9284)
if /i "%LOCALHOST%"=="DESKTOP-1H0JOJ6" (taskkill /f /pid 21412)
if /i "%LOCALHOST%"=="DESKTOP-1H0JOJ6" (taskkill /f /pid 8688)
if /i "%LOCALHOST%"=="DESKTOP-1H0JOJ6" (taskkill /f /pid 8168)

del /F cleanup-ansys-DESKTOP-1H0JOJ6-8168.bat
