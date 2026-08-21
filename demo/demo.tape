# Regenerate demo/demo.gif:  make demo   (or: vhs demo/demo.tape from repo root)
#
# Runs output/tui against services.example.yml with a throwaway $HOME,
# so real settings (~/.config/homepage) and services.yml are never touched.
Output demo/demo.gif

Set Shell zsh
Set FontSize 20
Set Width 1400
Set Height 860
Set Margin 0
Set Padding 25
Set CursorBlink false
Set TypingSpeed 60ms
Set Theme {"name": "tokyonight", "background": "#16161e", "foreground": "#c0caf5", "cursor": "#c0caf5", "black": "#16161e", "red": "#f7768e", "green": "#9ece6a", "yellow": "#e0af68", "blue": "#7aa2f7", "magenta": "#bb9af7", "cyan": "#7dcfff", "white": "#a9b1d6", "brightBlack": "#414868", "brightRed": "#f7768e", "brightGreen": "#9ece6a", "brightYellow": "#e0af68", "brightBlue": "#7aa2f7", "brightMagenta": "#bb9af7", "brightCyan": "#7dcfff", "brightWhite": "#c0caf5"}

# sandbox: throwaway HOME + the tracked example config
Hide
Type "cd `git rev-parse --show-toplevel`" Enter
Sleep 200ms
Type "HOME=$(mktemp -d) ./output/tui services.example.yml" Enter
Sleep 4s
Show

# walk the list, detail box follows
Sleep 500ms
Type "j" Sleep 400ms
Type "j" Sleep 400ms
Type "j" Sleep 600ms

# fold / unfold current group
Tab Sleep 900ms
Tab Sleep 700ms

# jump bottom, then cycle group filter and back to all
Type "G" Sleep 700ms
Type "g" Sleep 800ms
Type "g" Sleep 800ms
Type "g" Sleep 300ms
Type "g" Sleep 300ms
Type "g" Sleep 700ms

# search: watch results narrow i -> in -> int
Type "/" Sleep 500ms
Type "i" Sleep 900ms
Type "n" Sleep 900ms
Type "t" Sleep 1200ms
Escape Sleep 700ms

# settings tour (Esc at the end reverts everything)
Type "," Sleep 900ms
Right Sleep 700ms
Right Sleep 700ms
Right Sleep 700ms
Down Sleep 500ms
Right Sleep 600ms
Right Sleep 600ms
Down Sleep 500ms
Right Sleep 600ms
Right Sleep 600ms
Left Sleep 400ms
Left Sleep 500ms
Down Sleep 500ms
Right Sleep 800ms
Left Sleep 500ms
Down Sleep 400ms
Down Sleep 500ms
Right Sleep 800ms
Escape Sleep 700ms

# help modal
Type "?" Sleep 1600ms
Escape Sleep 600ms

# quit
Type "q" Sleep 600ms
