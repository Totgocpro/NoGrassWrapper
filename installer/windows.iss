; NoGrassWrapper Windows Installer — Inno Setup Script
; Requires Inno Setup 6+ (https://jrsoftware.org/isdl.php)

#define MyAppName "NoGrassWrapper"
#define MyAppExeName "nograsswrapper.exe"
#define MyAppPublisher "NoGrassWrapper"
#define MyAppVersion "0.0.0"

[Setup]
AppName={#MyAppName}
AppPublisher={#MyAppPublisher}
AppVersion={#MyAppVersion}
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
OutputDir=..\dist
OutputBaseFilename=NoGrassWrapper-Setup
Compression=lzma2
SolidCompression=yes
UninstallDisplayIcon={app}\{#MyAppExeName}
PrivilegesRequired=lowest
DisableProgramGroupPage=yes
WizardStyle=modern
ChangesEnvironment=no

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "autostart"; Description: "Start {#MyAppName} when I log in"; GroupDescription: "Startup options:"; Flags: checkedonce

[Files]
Source: "..\{#MyAppExeName}"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\README.md"; DestDir: "{app}"; Flags: ignoreversion
; No additional DLLs needed — Go builds statically linked binaries

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\Uninstall {#MyAppName}"; Filename: "{uninstallexe}"
Name: "{userstartup}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: autostart

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "Launch {#MyAppName}"; Flags: nowait postinstall skipifsilent

; ─── Custom pages ──────────────────────────────────────────────

[Code]

var
  UsernamePage: TInputQueryWizardPage;
  AvatarPage: TInputFileWizardPage;

procedure InitializeWizard;
begin
  // Username page
  UsernamePage := CreateInputQueryPage(
    wpSelectTasks,
    'Profile Setup',
    'Tell us a bit about yourself',
    'Enter your display name (optional). This will appear on your wrapper images.'
  );
  UsernamePage.Add('Username:', False);

  // Avatar page
  AvatarPage := CreateInputFilePage(
    UsernamePage.ID,
    'Avatar Selection',
    'Choose a profile picture',
    'Select an image file to use as your avatar on wrapper images (optional).' + #13#10 +
    'Supports PNG, JPG, GIF, and WebP formats.'
  );
  AvatarPage.Add(
    'Avatar file:',
    'Image files|*.png;*.jpg;*.jpeg;*.gif;*.webp|All files|*.*',
    '.png'
  );
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  ConfigDir, ConfigFile: string;
  Username, AvatarPath: string;
  Json: string;
begin
  if CurStep = ssPostInstall then
  begin
    ConfigDir := ExpandConstant('{userappdata}\NoGrassWrapper');
    if not DirExists(ConfigDir) then
      ForceDirectories(ConfigDir);

    ConfigFile := ConfigDir + '\usage.json';
    Username := UsernamePage.Values[0];
    AvatarPath := AvatarPage.Values[0];

    // Escape quotes for JSON
    StringChange(Username, '"', '\"');
    StringChange(AvatarPath, '"', '\"');

    // Build minimal JSON with username and avatar_path
    Json := '{' +
      '"version":1,' +
      '"username":"' + Username + '",' +
      '"avatar_path":"' + AvatarPath + '",' +
      '"daily_records":{},' +
      '"current_day":"",' +
      '"current_app":"",' +
      '"streak":0,' +
      '"longest_streak":0,' +
      '"unlocked_achievements":[],' +
      '"cpu_avg":0,' +
      '"cpu_samples":0,' +
      '"ram_avg":0,' +
      '"ram_samples":0,' +
      '"gpu_avg":0,' +
      '"gpu_samples":0' +
      '}';

    SaveStringToFile(ConfigFile, Json, False);
  end;
end;
