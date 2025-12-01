!include "MUI2.nsh"

Name "GoConnect"
OutFile "GoConnect-Setup.exe"
InstallDir "$PROGRAMFILES\GoConnect"
InstallDirRegKey HKCU "Software\GoConnect" ""
RequestExecutionLevel admin

!define MUI_ABORTWARNING

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "LICENSE"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_LANGUAGE "English"

Section "GoConnect" SecGoConnect
  SetOutPath "$INSTDIR"
  
  ; Install Desktop App
  ; Assumes the binary is built and located here
  File "desktop/src-tauri/target/release/GoConnect.exe"
  
  ; Install CLI
  ; Assumes the binary is built and located here
  File "cli/goconnect.exe"
  
  ; Create Shortcuts
  CreateDirectory "$SMPROGRAMS\GoConnect"
  CreateShortcut "$SMPROGRAMS\GoConnect\GoConnect.lnk" "$INSTDIR\GoConnect.exe"
  CreateShortcut "$SMPROGRAMS\GoConnect\GoConnect CLI.lnk" "$INSTDIR\goconnect.exe"
  
  ; Write Uninstaller
  WriteUninstaller "$INSTDIR\Uninstall.exe"
  
  ; Registry keys for Add/Remove programs
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\GoConnect" "DisplayName" "GoConnect"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\GoConnect" "UninstallString" "$\"$INSTDIR\Uninstall.exe$\""
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\GoConnect" "Publisher" "GoConnect"
SectionEnd

Section "Uninstall"
  Delete "$INSTDIR\GoConnect.exe"
  Delete "$INSTDIR\goconnect.exe"
  Delete "$INSTDIR\Uninstall.exe"
  
  RMDir "$INSTDIR"
  
  Delete "$SMPROGRAMS\GoConnect\GoConnect.lnk"
  Delete "$SMPROGRAMS\GoConnect\GoConnect CLI.lnk"
  RMDir "$SMPROGRAMS\GoConnect"
  
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\GoConnect"
  DeleteRegKey HKCU "Software\GoConnect"
SectionEnd
