!include "MUI2.nsh"

Name "DevProc"
OutFile "devproc-installer.exe"
InstallDir "$LOCALAPPDATA\DevProc"
RequestExecutionLevel user

!define MUI_ABORTWARNING

!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Section "Install"
  SetOutPath "$INSTDIR"
  File "build\bin\devproc.exe"

  ; Create start menu shortcut
  CreateDirectory "$SMPROGRAMS\DevProc"
  CreateShortcut "$SMPROGRAMS\DevProc\DevProc.lnk" "$INSTDIR\devproc.exe"
  CreateShortcut "$SMPROGRAMS\DevProc\Uninstall.lnk" "$INSTDIR\uninstall.exe"

  ; Create startup shortcut (launch on login)
  CreateShortcut "$SMSTARTUP\DevProc.lnk" "$INSTDIR\devproc.exe"

  ; Create uninstaller
  WriteUninstaller "$INSTDIR\uninstall.exe"

  ; Add to Add/Remove Programs
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\DevProc" "DisplayName" "DevProc"
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\DevProc" "UninstallString" "$INSTDIR\uninstall.exe"
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\DevProc" "InstallLocation" "$INSTDIR"
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\DevProc" "Publisher" "DevProc"
  WriteRegDWORD HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\DevProc" "NoModify" 1
  WriteRegDWORD HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\DevProc" "NoRepair" 1
SectionEnd

Section "Uninstall"
  ; Kill running instance
  nsExec::ExecToLog 'taskkill /F /IM devproc.exe'

  Delete "$INSTDIR\devproc.exe"
  Delete "$INSTDIR\uninstall.exe"
  RMDir "$INSTDIR"

  Delete "$SMPROGRAMS\DevProc\DevProc.lnk"
  Delete "$SMPROGRAMS\DevProc\Uninstall.lnk"
  RMDir "$SMPROGRAMS\DevProc"
  Delete "$SMSTARTUP\DevProc.lnk"

  DeleteRegKey HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\DevProc"
SectionEnd
