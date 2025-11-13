---
title: "TUI Issues"
weight: 4
toc: true
---

Solutions for Terminal User Interface rendering and interaction problems.

## TUI Issues

### TUI Display is Garbled or Has Rendering Artifacts

**Symptom**: Terminal displays corrupted characters, boxes, or strange symbols in TUI mode

**Cause**: Terminal doesn't support required Unicode characters or colors

**Solutions**:

1. **Check terminal emulator**
   ```bash
   # Verify TERM variable
   echo $TERM

   # Should be: xterm-256color, screen-256color, or similar
   ```

2. **Set proper TERM variable**
   ```bash
   # For most terminals
   export TERM=xterm-256color

   # Add to ~/.bashrc or ~/.zshrc for persistence
   echo 'export TERM=xterm-256color' >> ~/.bashrc
   ```

3. **Try different terminal**
   - Windows: Use Windows Terminal instead of CMD
   - macOS: Use iTerm2 or default Terminal.app
   - Linux: Use GNOME Terminal, Konsole, or Alacritty

4. **Check font support**
   - Use monospace font with Unicode support
   - Recommended: Fira Code, JetBrains Mono, Cascadia Code

---

### Keyboard Shortcuts Not Working

**Symptom**: Pressing `n`, `/`, `Ctrl+P` or other shortcuts does nothing

**Cause**: Terminal emulator intercepts keys or conflicts with system shortcuts

**Solutions**:

1. **Check terminal keybindings**
   - Review terminal preferences for conflicting shortcuts
   - Disable terminal shortcuts that overlap with TUI keys

2. **Verify focus**
   ```bash
   # Ensure TUI has focus, not search field or modal
   # Press Esc to close any open modal
   # Then try shortcut again
   ```

3. **Test basic shortcuts**
   - `q` - Should quit (if not in modal)
   - `?` - Should show help
   - If these don't work, terminal may be blocking input

4. **Platform-specific fixes**

   **Windows Terminal:**
   ```json
   // settings.json - Remove conflicting keybindings
   {
     "actions": [
       // Remove any that conflict with pass-cli
     ]
   }
   ```

   **macOS:**
   - System Preferences → Keyboard → Shortcuts
   - Disable conflicting shortcuts

   **Linux:**
   - Check GNOME/KDE shortcuts in system settings

---

### TUI Launches to Black or Unresponsive Screen

**Symptom**: TUI opens but shows black screen or doesn't respond to input

**Cause**: Terminal size too small or initialization issue

**Solutions**:

1. **Check terminal size**
   ```bash
   # Check dimensions
   tput cols  # Width (should be ≥60)
   tput lines # Height (should be ≥30)

   # Resize terminal window if needed
   ```

2. **Force TUI refresh**
   - Press `Ctrl+L` to redraw screen
   - Or quit and restart

3. **Try CLI mode to verify vault works**
   ```bash
   # If vault is problem, TUI will also fail
   pass-cli list

   # If CLI works but TUI doesn't, report bug
   ```

4. **Check for error messages**
   ```bash
   # Run TUI with verbose flag to capture debug output
   pass-cli tui --verbose 2>&1 | tee tui-error.log
   ```

---

### Search Function (`/`) Not Filtering Results

**Symptom**: Press `/` but search doesn't filter credentials

**Cause**: Focus not on main view or search input not activated

**Solutions**:

1. **Ensure main view focused**
   ```bash
   # Close any open modals first
   # Press Esc to close modal
   # Press Esc again to clear search (if active)
   # Then press / to start new search
   ```

2. **Verify you're in TUI mode**
   ```bash
   # Launch TUI
   pass-cli  # No arguments

   # Not:
   pass-cli list  # This is CLI mode
   ```

3. **Test search activation**
   - Press `/`
   - Input field should appear at top of table
   - Type search query
   - Results should filter in real-time
   - Press `Esc` to exit search

4. **Check for key conflicts**
   - `/` might be intercepted by terminal
   - Try different terminal emulator

---

### Ctrl+P Password Toggle Not Working

**Symptom**: Pressing `Ctrl+P` in add/edit forms doesn't toggle password visibility

**Cause**: Not in form context, or terminal intercepts `Ctrl+P` as backspace

**Solutions**:

1. **Verify you're in a form**
   ```bash
   # Open add form
   press 'n'  # Should open "Add Credential" modal

   # Navigate to password field
   press Tab until focused on Password field

   # Toggle visibility
   press Ctrl+P  # Should show/hide password
   ```

2. **Terminal backspace mapping**
   - Some terminals map `Ctrl+P` to backspace
   - Try pressing `Backspace` first to test
   - If `Ctrl+P` deletes character, terminal is intercepting

3. **Alternative verification method**
   ```bash
   # Check password field label
   # Should change from "Password" to "Password [VISIBLE]"
   ```

---

### Sidebar or Detail Panel Not Visible

**Symptom**: Sidebar or detail panel missing or doesn't appear

**Cause**: Terminal too narrow or panel hidden by toggle state

**Solutions**:

1. **Check terminal width**
   ```bash
   # Check columns
   tput cols

   # Sidebar auto-hides below 80 cols
   # Detail panel auto-hides below 100 cols
   ```

2. **Toggle visibility**
   - Press `s` to toggle sidebar (cycles 3 states)
   - Press `i` to toggle detail panel (cycles 3 states)
   - States: Auto (responsive) → Hide → Show → Auto

3. **Check status bar**
   - After pressing `s` or `i`, status bar shows current state
   - "Sidebar: Auto (responsive)" means it follows width rules
   - "Sidebar: Visible" means forced to show

4. **Force show on narrow terminal**
   ```bash
   # Press 's' twice to force sidebar visible
   # Press 'i' twice to force detail panel visible
   ```

---

### Usage Locations Not Appearing in Detail Panel

**Symptom**: "Usage Locations" section missing from credential details

**Cause**: Credential hasn't been accessed yet via `pass-cli get`

**Expected Behavior**:
- Usage tracking only records `pass-cli get <service>` commands
- TUI viewing doesn't count as "usage"
- New credentials have no usage data

**Solutions**:

1. **Access credential from CLI to generate usage**
   ```bash
   # Change to project directory
   cd ~/projects/my-app

   # Access credential
   pass-cli get github

   # Now check TUI detail panel - should show usage
   pass-cli
   ```

2. **Verify expected behavior**
   - Only credentials accessed via `pass-cli get` show usage
   - Usage shows: working directory, access count, timestamp
   - Empty state shows: "No usage recorded"

---

