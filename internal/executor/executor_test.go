package executor

import "testing"

func TestParseCommands(t *testing.T) {
	input := "rm -rf ~/Library/Caches/old\nrm ~/Downloads/big.dmg\n\n# comment\n"
	cmds := ParseCommands(input)
	if len(cmds) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(cmds))
	}
	if cmds[0].Raw != "rm -rf ~/Library/Caches/old" {
		t.Errorf("unexpected first command: %s", cmds[0].Raw)
	}
}

func TestParseCommandsEmpty(t *testing.T) {
	cmds := ParseCommands("")
	if len(cmds) != 0 {
		t.Errorf("expected 0 commands for empty input, got %d", len(cmds))
	}
}

func TestValidateBlocksDangerousPaths(t *testing.T) {
	dangerous := []string{
		"rm -rf /System/Library",
		"rm -rf /usr/bin",
		"rm /bin/bash",
		"rm -rf /sbin/",
		"rm /etc/hosts",
		"rm -rf /private/var/db/something",
	}
	for _, raw := range dangerous {
		cmd := Command{Raw: raw}
		err := ValidateCommand(&cmd)
		if err == nil {
			t.Errorf("expected error for dangerous command: %s", raw)
		}
	}
}

func TestValidateBlocksSudo(t *testing.T) {
	cmd := Command{Raw: "sudo rm -rf /tmp/old"}
	err := ValidateCommand(&cmd)
	if err == nil {
		t.Error("expected error for sudo command")
	}
}

func TestValidateAllowsSafePaths(t *testing.T) {
	safe := []string{
		"rm -rf ~/Library/Caches/com.old.app",
		"rm ~/Downloads/old-file.dmg",
		"rm -rf /tmp/build-output",
	}
	for _, raw := range safe {
		cmd := Command{Raw: raw}
		err := ValidateCommand(&cmd)
		if err != nil {
			t.Errorf("expected safe command to pass: %s, got: %v", raw, err)
		}
	}
}

func TestValidateWarnsLargeDelete(t *testing.T) {
	cmd := Command{Raw: "rm -rf ~/big-folder"}
	// ValidateCommand should set a warning but not error for unresolvable paths
	err := ValidateCommand(&cmd)
	if err != nil {
		t.Errorf("shouldn't error on normal path: %v", err)
	}
}
