# Pass-CLI Themes

Pass-CLI TUI supports multiple color themes that can be configured via the config file.

## Available Themes

### 1. Dracula (Default)
Dark theme with vibrant purples, pinks, and cyans. Perfect for low-light environments.

**Colors:**
- Background: Deep dark purple (#282a36)
- Accents: Cyan, pink, purple
- Status: Green (success), red (error), yellow (warning)

### 2. Nord
Cool, bluish theme inspired by arctic ice and polar nights.

**Colors:**
- Background: Dark blue-gray (#2e3440)
- Accents: Frost blues and teals
- Status: Muted greens, reds, and yellows

### 3. Gruvbox
Warm, retro theme with earthy tones and high contrast.

**Colors:**
- Background: Dark gray-brown (#282828)
- Accents: Warm aqua, yellow, orange
- Status: Vibrant greens, reds, yellows

### 4. Monokai
Vibrant, colorful theme popular in code editors.

**Colors:**
- Background: Very dark gray (#272822)
- Accents: Bright cyan, purple, yellow
- Status: Neon greens, hot pinks, bright yellows

## Configuration

To change your theme, edit your config file at `~/.pass-cli/config.yml`:

```yaml
# Valid themes: dracula, nord, gruvbox, monokai
theme: "nord"
```

## Changing Themes

1. Edit your config file:
   ```bash
   # macOS/Linux
   nano ~/.pass-cli/config.yml
   
   # Windows (PowerShell)
   notepad $env:USERPROFILE\.pass-cli\config.yml
   ```

2. Set the theme:
   ```yaml
   theme: "nord"  # or dracula, gruvbox, monokai
   ```

3. Save and restart pass-cli TUI:
   ```bash
   pass-cli tui
   ```

## Theme Preview

Launch the TUI and try different themes to see which one you prefer:

```bash
# Try Nord
echo 'theme: "nord"' >> ~/.pass-cli/config.yml
pass-cli tui

# Try Gruvbox
echo 'theme: "gruvbox"' >> ~/.pass-cli/config.yml  
pass-cli tui

# Try Monokai
echo 'theme: "monokai"' >> ~/.pass-cli/config.yml
pass-cli tui
```

## Validation

If you specify an invalid theme name, Pass-CLI will:
1. Show a warning
2. Fall back to the default Dracula theme
3. Continue running normally

Valid theme names are: `dracula`, `nord`, `gruvbox`, `monokai`

## Future Plans

- More themes (Solarized Dark, Tokyo Night, One Dark, etc.)
- Custom color overrides
- Light themes
