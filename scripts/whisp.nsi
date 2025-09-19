;NSIS Installer Script for Whisp
;This script creates a professional Windows installer for the Whisp messaging application

!include "MUI2.nsh"
!include "FileFunc.nsh"
!include "LogicLib.nsh"

;General Configuration
Name "Whisp"
OutFile "whisp-windows-installer.exe"
Unicode True
InstallDir "$PROGRAMFILES\Whisp"
InstallDirRegKey HKCU "Software\Whisp" ""
RequestExecutionLevel admin

;Modern UI Configuration
!define MUI_ABORTWARNING
!define MUI_ICON "assets\icons\icon.ico"
!define MUI_UNICON "assets\icons\icon.ico"
!define MUI_HEADERIMAGE
!define MUI_HEADERIMAGE_BITMAP "assets\icons\icon.bmp"
!define MUI_WELCOMEFINISHPAGE_BITMAP "assets\icons\icon.bmp"

;Pages
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "LICENSE"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_WELCOME
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

;Languages
!insertmacro MUI_LANGUAGE "English"

;Version Information
VIProductVersion "1.0.0.0"
VIAddVersionKey "ProductName" "Whisp"
VIAddVersionKey "CompanyName" "OPD AI"
VIAddVersionKey "FileVersion" "1.0.0"
VIAddVersionKey "ProductVersion" "1.0.0"
VIAddVersionKey "FileDescription" "Secure Cross-Platform Messaging Application"

;Installer Sections
Section "Whisp" SecApp
    SectionIn RO

    SetOutPath "$INSTDIR"

    ;Copy application files
    DetailPrint "Installing Whisp application..."
    File "build\windows\whisp.exe"

    ;Create data directory
    CreateDirectory "$APPDATA\Whisp"
    CreateDirectory "$APPDATA\Whisp\media"
    CreateDirectory "$APPDATA\Whisp\cache"

    ;Create desktop shortcut
    CreateShortCut "$DESKTOP\Whisp.lnk" "$INSTDIR\whisp.exe" "" "$INSTDIR\whisp.exe" 0

    ;Create start menu entries
    CreateDirectory "$SMPROGRAMS\Whisp"
    CreateShortCut "$SMPROGRAMS\Whisp\Whisp.lnk" "$INSTDIR\whisp.exe" "" "$INSTDIR\whisp.exe" 0
    CreateShortCut "$SMPROGRAMS\Whisp\Uninstall.lnk" "$INSTDIR\Uninstall.exe" "" "$INSTDIR\Uninstall.exe" 0

    ;Store installation folder
    WriteRegStr HKCU "Software\Whisp" "" $INSTDIR

    ;Create uninstaller
    WriteUninstaller "$INSTDIR\Uninstall.exe"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Whisp" "DisplayName" "Whisp"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Whisp" "UninstallString" "$INSTDIR\Uninstall.exe"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Whisp" "DisplayVersion" "1.0.0"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Whisp" "Publisher" "OPD AI"
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Whisp" "NoModify" 1
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Whisp" "NoRepair" 1
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Whisp" "DisplayIcon" "$INSTDIR\whisp.exe"

    ;Register file associations (optional)
    WriteRegStr HKCR ".whisp" "" "Whisp.File"
    WriteRegStr HKCR "Whisp.File" "" "Whisp Data File"
    WriteRegStr HKCR "Whisp.File\DefaultIcon" "" "$INSTDIR\whisp.exe,0"
    WriteRegStr HKCR "Whisp.File\shell\open\command" "" '"$INSTDIR\whisp.exe" "%1"'

SectionEnd

;Uninstaller Section
Section "Uninstall"

    ;Remove files
    Delete "$INSTDIR\whisp.exe"
    Delete "$INSTDIR\Uninstall.exe"

    ;Remove shortcuts
    Delete "$DESKTOP\Whisp.lnk"
    Delete "$SMPROGRAMS\Whisp\Whisp.lnk"
    Delete "$SMPROGRAMS\Whisp\Uninstall.lnk"
    RMDir "$SMPROGRAMS\Whisp"

    ;Remove registry entries
    DeleteRegKey HKCU "Software\Whisp"
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Whisp"
    DeleteRegKey HKCR ".whisp"
    DeleteRegKey HKCR "Whisp.File"

    ;Remove directories (only if empty)
    RMDir "$INSTDIR"
    RMDir "$APPDATA\Whisp\cache"
    RMDir "$APPDATA\Whisp\media"
    RMDir "$APPDATA\Whisp"

SectionEnd

;Functions
Function .onInit
    ;Check if already installed
    ReadRegStr $R0 HKCU "Software\Whisp" ""
    ${If} $R0 != ""
        MessageBox MB_YESNO "Whisp is already installed. Do you want to reinstall?" IDYES continue
        Abort
        continue:
    ${EndIf}
FunctionEnd

Function .onInstSuccess
    MessageBox MB_YESNO "Installation completed successfully! Would you like to launch Whisp now?" IDNO end
    Exec '"$INSTDIR\whisp.exe"'
    end:
FunctionEnd
