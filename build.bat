REM Snarky - Zero Knowledge Dead Drop
REM Copyright (C) 2026 Sapadian LLC.
REM
REM This program is free software: you can redistribute it and/or modify
REM it under the terms of the GNU Affero General Public License as published
REM by the Free Software Foundation, either version 3 of the License, or
REM (at your option) any later version.
REM
REM This program is distributed in the hope that it will be useful,
REM but WITHOUT ANY WARRANTY; without even the implied warranty of
REM MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
REM GNU Affero General Public License for more details.
REM
REM You should have received a copy of the GNU Affero General Public License
REM along with this program.  If not, see https://www.gnu.org/licenses/


@echo off
echo ==========================================
echo      Building Snarky for all platforms
echo ==========================================

REM Create a 'bin' folder to keep things tidy
if not exist bin mkdir bin

REM 1. FreeBSD (The Server)
echo [1/5] Building for FreeBSD (Server)...
set GOOS=freebsd
set GOARCH=amd64
go build -o bin/snarky-bsd main.go
if %errorlevel% neq 0 goto :error

REM 2. Windows (Client)
echo [2/5] Building for Windows (Client)...
set GOOS=windows
set GOARCH=amd64
go build -o bin/snarky.exe main.go
if %errorlevel% neq 0 goto :error

REM 3. Linux (Client)
echo [3/5] Building for Linux (Client)...
set GOOS=linux
set GOARCH=amd64
go build -o bin/snarky-linux main.go
if %errorlevel% neq 0 goto :error

REM 4. Mac Intel (Client)
echo [4/5] Building for Mac Intel...
set GOOS=darwin
set GOARCH=amd64
go build -o bin/snarky-mac-intel main.go
if %errorlevel% neq 0 goto :error

REM 5. Mac M1/M2/Silicon (Client)
echo [5/5] Building for Mac Silicon...
set GOOS=darwin
set GOARCH=arm64
go build -o bin/snarky-mac-m1 main.go
if %errorlevel% neq 0 goto :error

echo.
echo ==========================================
echo      SUCCESS! Binaries are in /bin
echo ==========================================
pause
exit /b 0

:error
echo.
echo !!!!!!! BUILD FAILED !!!!!!!
pause
exit /b 1