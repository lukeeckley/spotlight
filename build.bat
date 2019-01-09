@echo off
REM SET PATH=%PATH%;C:\msys64\mingw64\bin

REM rsrc must be installed to create the .syso file for the icon to be embedded in the exe
REM rsrc -ico pinkparty.ico -o spotlight.syso
REM Using goversioninfo to set file attributes and icon
REM goversioninfo -icon=pinkparty.ico

REM First compile it with go and the appropriate flags
SET FLAGS="-H windowsgui -s -w"
go build -ldflags %FLAGS%

REM how to get the current directory name
for %%* in (.) do SET CurrentDir=%%~nx*

REM If upx.exe is found in the path, use it to make this exe even smaller
for %%X in (upx.exe) do (set FOUND=%%~$PATH:X)
if defined FOUND (
	upx.exe -9 %CurrentDir%.exe	
)

pause