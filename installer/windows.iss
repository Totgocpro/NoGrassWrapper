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
  UpdatePage: TOutputMsgWizardPage;

function IsUpdateInstall(): Boolean;
var
  UninstallKey: string;
begin
  // Already installed if the executable exists in the install dir,
  // an uninstall entry exists, or user data already exists.
  Result := FileExists(ExpandConstant('{app}\{#MyAppExeName}'));
  if not Result then
  begin
    UninstallKey := Format('Software\Microsoft\Windows\CurrentVersion\Uninstall\%s_is1', ['{#MyAppName}']);
    Result := RegValueExists(HKEY_CURRENT_USER, UninstallKey, 'UninstallString');
  end;
  if not Result then
    Result := FileExists(ExpandConstant('{userappdata}\NoGrassWrapper\usage.json'));
end;

procedure InitializeWizard;
begin
  // Update page (only shown when updating an existing installation)
  UpdatePage := CreateOutputMsgPage(
    wpWelcome,
    'Updating {#MyAppName}',
    'Updating to version {#MyAppVersion}',
    'An existing installation of {#MyAppName} was detected on this computer.' + #13#10 +
    'The installer will update your current installation to version {#MyAppVersion}.' + #13#10 + #13#10 +
    'Your username, avatar and usage history will be preserved.'
  );

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

function ShouldSkipPage(PageID: Integer): Boolean;
begin
  Result := False;
  if PageID = UpdatePage.ID then
    // Update page is only shown when updating an existing installation
    Result := not IsUpdateInstall
  else if (PageID = UsernamePage.ID) or (PageID = AvatarPage.ID) then
    // Profile pages are only shown on a fresh install
    Result := IsUpdateInstall;
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  ConfigDir, ConfigFile: string;
  Username, AvatarPath, DestAvatar, Ext: string;
  Json: string;
begin
  if CurStep = ssPostInstall then
  begin
    // On update, keep the existing profile and usage data untouched
    if IsUpdateInstall then
      exit;

    ConfigDir := ExpandConstant('{userappdata}\NoGrassWrapper');
    if not DirExists(ConfigDir) then
      ForceDirectories(ConfigDir);

    ConfigFile := ConfigDir + '\usage.json';
    Username := UsernamePage.Values[0];
    AvatarPath := AvatarPage.Values[0];

    // Copy avatar to config dir if one was selected
    DestAvatar := '';
    if AvatarPath <> '' then
    begin
      Ext := '.png';
      if Pos('.jpg', Lowercase(AvatarPath)) > 0 then Ext := '.jpg'
      else if Pos('.jpeg', Lowercase(AvatarPath)) > 0 then Ext := '.jpeg'
      else if Pos('.gif', Lowercase(AvatarPath)) > 0 then Ext := '.gif'
      else if Pos('.webp', Lowercase(AvatarPath)) > 0 then Ext := '.webp';
      DestAvatar := ConfigDir + '\avatar' + Ext;
      FileCopy(AvatarPath, DestAvatar, False);
    end;

    // Escape special characters for JSON
    StringChange(Username, '\', '\\');
    StringChange(Username, '"', '\"');
    StringChange(DestAvatar, '\', '\\');
    StringChange(DestAvatar, '"', '\"');

    // Build minimal JSON with username and avatar_path
    Json := '{' +
      '"version":1,' +
      '"username":"' + Username + '",' +
      '"avatar_path":"' + DestAvatar + '",' +
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
end.
