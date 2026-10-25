!ifndef VERSION
  !define VERSION "0.1"
!endif

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
  File /r "skills"
  File "README.md"
  File "config.example.json"

  CreateDirectory "$SMPROGRAMS\Argus"
  CreateShortcut "$DESKTOP\Argus.lnk" "$INSTDIR\argus.exe" "-addr 127.0.0.1:8787"
  CreateShortcut "$SMPROGRAMS\Argus\Argus.lnk" "$INSTDIR\argus.exe" "-addr 127.0.0.1:8787"
  WriteUninstaller "$INSTDIR\Uninstall Argus.exe"
SectionEnd

Section "Uninstall"
  Delete "$DESKTOP\Argus.lnk"
  Delete "$SMPROGRAMS\Argus\Argus.lnk"
  RMDir "$SMPROGRAMS\Argus"
  Delete "$INSTDIR\argus.exe"
  Delete "$INSTDIR\start.bat"
  Delete "$INSTDIR\README.md"
  Delete "$INSTDIR\config.example.json"
  Delete "$INSTDIR\Uninstall Argus.exe"
  RMDir /r "$INSTDIR\skills"
  RMDir "$INSTDIR"
SectionEnd
