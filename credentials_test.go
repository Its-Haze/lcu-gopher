package lcu

import (
	"errors"
	"path/filepath"
	"testing"
)

// fakeLister stands in for the OS process table.
type fakeLister struct {
	procs []ProcessInfo
	err   error
}

func (f fakeLister) LeagueUxProcesses() ([]ProcessInfo, error) {
	return f.procs, f.err
}

const realCmdline = `"C:\Riot Games\League of Legends\LeagueClientUx.exe" ` +
	`--riotclient-auth-token=abc --app-port=52931 --remoting-auth-token=xY-9_tokenZ --locale=en_GB`

func TestFindCredentialsFromProcess(t *testing.T) {
	tests := []struct {
		name       string
		lister     fakeLister
		wantErr    bool
		wantPort   int
		wantPass   string
		wantLeague string
	}{
		{
			name:       "reads port and token from the command line",
			lister:     fakeLister{procs: []ProcessInfo{{Cmdline: realCmdline, Exe: `C:\Riot Games\League of Legends\LeagueClientUx.exe`}}},
			wantPort:   52931,
			wantPass:   "xY-9_tokenZ",
			wantLeague: filepath.Dir(`C:\Riot Games\League of Legends\LeagueClientUx.exe`),
		},
		{
			name:       "falls back to the command line when the exe path is unreadable",
			lister:     fakeLister{procs: []ProcessInfo{{Cmdline: realCmdline}}},
			wantPort:   52931,
			wantPass:   "xY-9_tokenZ",
			wantLeague: filepath.Dir(`C:\Riot Games\League of Legends\LeagueClientUx.exe`),
		},
		{
			name:     "finds a non-default install directory",
			lister:   fakeLister{procs: []ProcessInfo{{Cmdline: realCmdline, Exe: `D:\Games\Riot\League of Legends\LeagueClientUx.exe`}}},
			wantPort: 52931,
			wantPass: "xY-9_tokenZ",
			// The old lockfile scan only looked at drive roots, so this
			// install was undiscoverable.
			wantLeague: filepath.Dir(`D:\Games\Riot\League of Legends\LeagueClientUx.exe`),
		},
		{
			name:    "skips a process whose command line carries no credentials",
			lister:  fakeLister{procs: []ProcessInfo{{Cmdline: `"C:\x\LeagueClientUx.exe" --locale=en_GB`}}},
			wantErr: true,
		},
		{
			name:    "reports no running client",
			lister:  fakeLister{},
			wantErr: true,
		},
		{
			// A live client shows one LeagueClientUx.exe next to several
			// LeagueClientUxRender.exe subprocesses, none of which carry the
			// port. Only the real one must be used.
			name: "ignores renderer subprocesses that carry no credentials",
			lister: fakeLister{procs: []ProcessInfo{
				{Cmdline: `"C:\Riot Games\League of Legends\LeagueClientUxRender.exe" --type=renderer`, Exe: `C:\Riot Games\League of Legends\LeagueClientUxRender.exe`},
				{Cmdline: `"C:\Riot Games\League of Legends\LeagueClientUxRender.exe" --type=gpu-process`, Exe: `C:\Riot Games\League of Legends\LeagueClientUxRender.exe`},
				{Cmdline: realCmdline, Exe: `C:\Riot Games\League of Legends\LeagueClientUx.exe`},
			}},
			wantPort:   52931,
			wantPass:   "xY-9_tokenZ",
			wantLeague: `C:\Riot Games\League of Legends`,
		},
		{
			name:    "propagates a process table failure",
			lister:  fakeLister{err: errors.New("access denied")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.ProcessLister = tt.lister

			creds, err := findCredentialsFromProcess(cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got credentials %+v", creds)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if creds.Port != tt.wantPort {
				t.Errorf("port = %d, want %d", creds.Port, tt.wantPort)
			}
			if creds.Password != tt.wantPass {
				t.Errorf("password = %q, want %q", creds.Password, tt.wantPass)
			}
			if cfg.LeaguePath != tt.wantLeague {
				t.Errorf("LeaguePath = %q, want %q", cfg.LeaguePath, tt.wantLeague)
			}
		})
	}
}

func TestIsLeagueUx(t *testing.T) {
	tests := []struct {
		name string
		proc string
		want bool
	}{
		{"the client Ux process", "LeagueClientUx.exe", true},
		{"name matching is case insensitive", "leagueclientux.exe", true},
		{"the launcher is not the Ux process", "LeagueClient.exe", false},
		{"a renderer subprocess is not the Ux process", "LeagueClientUxRender.exe", false},
		{"the game itself is not the client", "League of Legends.exe", false},
		{"an unrelated process", "chrome.exe", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLeagueUx(tt.proc); got != tt.want {
				t.Errorf("isLeagueUx(%q) = %v, want %v", tt.proc, got, tt.want)
			}
		})
	}
}

func TestParseProcessOutput(t *testing.T) {
	t.Run("extracts port and token", func(t *testing.T) {
		creds, err := parseProcessOutput(realCmdline)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if creds.Port != 52931 || creds.Password != "xY-9_tokenZ" {
			t.Errorf("got %+v", creds)
		}
		if creds.Protocol != "https" {
			t.Errorf("protocol = %q, want https", creds.Protocol)
		}
	})

	t.Run("rejects a command line without a token", func(t *testing.T) {
		if _, err := parseProcessOutput("--app-port=52931"); err == nil {
			t.Error("expected an error")
		}
	})

	t.Run("rejects empty input", func(t *testing.T) {
		if _, err := parseProcessOutput(""); err == nil {
			t.Error("expected an error")
		}
	})
}

func TestCommandLineExe(t *testing.T) {
	tests := []struct {
		name    string
		cmdline string
		want    string
	}{
		{
			name:    "quoted path with spaces",
			cmdline: `"C:\Riot Games\League of Legends\LeagueClientUx.exe" --app-port=1`,
			want:    `C:\Riot Games\League of Legends\LeagueClientUx.exe`,
		},
		{
			name:    "unquoted path",
			cmdline: `C:\lol\LeagueClientUx.exe --app-port=1`,
			want:    `C:\lol\LeagueClientUx.exe`,
		},
		{
			name:    "no arguments",
			cmdline: `C:\lol\LeagueClientUx.exe`,
			want:    `C:\lol\LeagueClientUx.exe`,
		},
		{"unrelated executable", `C:\Windows\explorer.exe /n`, ""},
		{"unterminated quote", `"C:\lol\LeagueClientUx.exe`, ""},
		{"empty", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := commandLineExe(tt.cmdline); got != tt.want {
				t.Errorf("commandLineExe(%q) = %q, want %q", tt.cmdline, got, tt.want)
			}
		})
	}
}

func TestInstallDirFromCmdline(t *testing.T) {
	tests := []struct {
		name    string
		cmdline string
		want    string
	}{
		{
			// Exactly how a live client presents it: every argument quoted,
			// argv[0] with forward slashes, the flag with backslashes.
			name:    "quoted flag in a real command line",
			cmdline: `"C:/Riot Games/League of Legends/LeagueClientUx.exe" "--app-port=50379" "--install-directory=C:\Riot Games\League of Legends" "--app-name=LeagueClient"`,
			want:    `C:\Riot Games\League of Legends`,
		},
		{
			name:    "unquoted flag followed by another",
			cmdline: `LeagueClientUx.exe --install-directory=D:\Games\LoL --app-name=LeagueClient`,
			want:    `D:\Games\LoL`,
		},
		{
			name:    "unquoted flag at the end",
			cmdline: `LeagueClientUx.exe --install-directory=D:\Games\LoL`,
			want:    `D:\Games\LoL`,
		},
		{
			name:    "a path with spaces, unquoted and last",
			cmdline: `LeagueClientUx.exe --install-directory=D:\My Games\LoL`,
			want:    `D:\My Games\LoL`,
		},
		{"flag absent", `LeagueClientUx.exe --app-port=1`, ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := installDirFromCmdline(tt.cmdline); got != tt.want {
				t.Errorf("installDirFromCmdline() = %q, want %q", got, tt.want)
			}
		})
	}
}

// The flag is authoritative: it wins over the executable's own directory,
// which can differ if the Ux binary ever moves into a subdirectory.
func TestLeagueDirPrefersTheInstallFlag(t *testing.T) {
	proc := ProcessInfo{
		Cmdline: `"LeagueClientUx.exe" "--install-directory=C:\Riot Games\League of Legends"`,
		Exe:     `C:\Riot Games\League of Legends\subdir\LeagueClientUx.exe`,
	}
	if got, want := leagueDirFromProcess(proc), `C:\Riot Games\League of Legends`; got != want {
		t.Errorf("leagueDirFromProcess() = %q, want %q", got, want)
	}
}
