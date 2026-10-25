!ifndef VERSION
  !define VERSION "0.1"
!endif

!include "LogicLib.nsh"

Name "Argus"
OutFile "argus-${VERSION}-windows-amd64-setup.exe"
InstallDir "$PROGRAMFILES64\Argus"
RequestExecutionLevel admin
Unicode True

Page directory
Page instfiles
UninstPage uninstConfirm
UninstPage instfiles

Section "Argus"
  SetOutPath "$INSTDIR"
  File "argus.exe"
  File "start.bat"
  File "create-shortcuts.ps1"
  File "ArgusShortcuts.cs"
  File /r "skills"
  File "README.md"
  File "config.example.json"
  File /r "database-cache"

  ExecWait '"$SYSDIR\WindowsPowerShell\v1.0\powershell.exe" -NoProfile -NonInteractive -ExecutionPolicy Bypass -File "$INSTDIR\create-shortcuts.ps1"' $0
  ${If} $0 != 0
    Abort "Could not create Argus shortcuts."
  ${EndIf}
  WriteUninstaller "$INSTDIR\Uninstall Argus.exe"
SectionEnd

Section "Uninstall"
  Delete "$DESKTOP\Argus.lnk"
  Delete "$SMPROGRAMS\Argus\Argus.lnk"
  RMDir "$SMPROGRAMS\Argus"
  Delete "$INSTDIR\argus.exe"
  Delete "$INSTDIR\start.bat"
  Delete "$INSTDIR\create-shortcuts.ps1"
  Delete "$INSTDIR\ArgusShortcuts.cs"
  Delete "$INSTDIR\README.md"
  Delete "$INSTDIR\config.example.json"
  Delete "$INSTDIR\Uninstall Argus.exe"
  RMDir /r "$INSTDIR\skills"
  RMDir /r "$INSTDIR\database-cache"
  RMDir "$INSTDIR"
SectionEnd
