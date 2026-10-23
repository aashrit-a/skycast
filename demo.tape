# Records assets/demo.gif. Run with: make gif  (needs charmbracelet/vhs)
Output assets/demo.gif

Set FontSize 16
Set Width 1200
Set Height 600
Set Theme "Catppuccin Mocha"
# Match the animation's 12 fps so the GIF stays small (target < 5 MB).
Set Framerate 12
Set TypingSpeed 60ms

Type "skycast --demo"
Sleep 500ms
Enter
# Each demo scene shows for 6 seconds: this captures the first six.
Sleep 37s
Type "q"
