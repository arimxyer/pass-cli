# Integration Testing Checklist

This checklist is extracted from `quickstart.md` to standardize with prior specs (e.g., 001).

## Terminal Size Warning Integration

- [ ] Default thresholds (60×30) work without config file
- [ ] Custom thresholds from config trigger warning correctly
- [ ] Disabled warning (warning_enabled: false) prevents modal
- [ ] Invalid threshold shows modal error, uses defaults
- [ ] Warning modal displays current/required sizes from config

## Keybinding Integration

- [ ] Default keybindings work without config file
- [ ] Custom single-key bindings work in TUI
- [ ] Custom modifier key bindings work (ctrl+, alt+, shift+)
- [ ] Status bar reflects custom keybindings
- [ ] Help modal reflects custom keybindings
- [ ] Form hints reflect custom keybindings
- [ ] Conflicting bindings show modal error, use defaults
- [ ] Unknown action shows modal error, uses defaults

## CLI Commands

- [ ] `pass-cli config init` creates file with examples
- [ ] `pass-cli config edit` opens system default editor
- [ ] `pass-cli config edit` respects EDITOR env var
- [ ] `pass-cli config edit` falls back through editor chain (nano > vim > vi) on Linux/macOS
- [ ] `pass-cli config edit` shows clear error if no editor found
- [ ] `pass-cli config validate` reports errors with details
- [ ] `pass-cli config validate` succeeds with valid config
- [ ] `pass-cli config validate` handles missing file gracefully
- [ ] `pass-cli config validate` displays warnings to stdout (unknown fields)
- [ ] `pass-cli config reset` creates backup before overwrite
- [ ] `pass-cli config reset` overwrites existing backup file
- [ ] `pass-cli config reset` restores defaults

## Error Handling

- [ ] Empty config file uses all defaults (no error)
- [ ] Partial config merges with defaults correctly
- [ ] YAML syntax errors show modal with line numbers
- [ ] File permission errors show modal, use defaults
- [ ] File too large error shows modal, use defaults
- [ ] Unknown fields generate warnings (not errors)
- [ ] Multiple validation errors all shown in modal

## Cross-Platform

- [ ] Config path correct on Windows (%APPDATA%\\pass-cli\\)
- [ ] Config path correct on macOS (~/Library/Application Support/pass-cli/)
- [ ] Config path correct on Linux (~/.config/pass-cli/)
- [ ] Editor selection works on Windows (notepad)
- [ ] Editor selection works on macOS (nano/vim)
- [ ] Editor selection works on Linux (nano/vim)

## Runtime Behavior

- [ ] Config changes require TUI restart (no hot-reload)
- [ ] Config loaded only once at TUI startup
- [ ] Config validation errors displayed as modal in TUI
- [ ] Config validation warnings displayed as modal in TUI
- [ ] Multiple config errors shown in single modal

## Audit Logging

- [ ] Config load attempts logged (file path, success/failure)
- [ ] Validation errors logged with details
- [ ] Config file not found logged (info level)
- [ ] Config parse errors logged (error level)

## Performance

- [ ] Config load completes in <10ms (valid config)
- [ ] Config validation completes in <5ms
- [ ] TUI startup with config completes in <500ms total
- [ ] Config file at 99KB size limit loads successfully
- [ ] Config file at 100KB+ size shows error within 50ms
