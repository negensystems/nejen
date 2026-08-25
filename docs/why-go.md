# Why NEJEN is Built with Go

NEJEN's core desktop management and execution layer is built as a single compiled binary written in **Go**. This document explains the architectural and performance motivations behind this design choice compared to traditional shell-script based environments.

---

## The Core Problems with Bash + Python

Many traditional Linux desktop environments rely on dispatcher shell scripts that route commands to dozens of distinct Bash and Python scripts. This approach introduces several critical issues:

1. **Slow Execution (Process Spawning Overhead):** 
   Every time a status bar updates, a shortcut key is pressed, or a theme is changed, the system has to:
   * Spawn a Bash process.
   * Spawn a Python interpreter.
   * Spawn multiple shell utilities (`awk`, `sed`, `grep`, `cat`).
   
   This heavy process-spawning loop takes between **50ms and 200ms** of CPU time, causing micro-stuttering and wasting resources on low-power or battery-operated laptops.

2. **Brittle Dependency Chain:**
   Shell scripts depend on the host operating system's specific versions of command-line tools. Python scripts depend on the host Python interpreter version and library configurations. A routine system package update (`pacman -Syu`) can easily break scripts if syntax or paths change.

3. **Silent Failures and Lack of Safety:**
   Bash does not have compile-time checks. Variables containing spaces or unusual quotes can silently break execution paths. Debugging issues across interconnected scripts is error-prone.

---

## The Go Solution

Building the system management utility as a single compiled Go binary (`nejen`) resolves these issues:

### 1. No Interpreter Startup
Go compiles directly into native CPU machine instructions.
* Reaching the code that does the work costs one process, not a chain of them: no shell wrapper, no Python interpreter, no `awk`/`sed`/`grep` pipeline in between.
* This matters most where it happens most often. The Waybar modules (`nejen weather`, `nejen countdown`, the indicators) are polled continuously, and a keybinding has to feel instant.
* Commands that drive other programs (`hyprctl`, `pacman`, `notify-send`) are still only as fast as those programs. What Go removes is the overhead NEJEN itself adds on top.

### 2. One Binary, Few Dependencies
Go builds a single statically linked binary (`CGO_ENABLED=0`).
* No Python, PyYAML, or tabulate to keep in step with whatever the system interpreter happens to be this week.
* A routine `pacman -Syu` cannot break the dispatcher by moving an interpreter or bumping a library soname.
* This is not the same as having no dependencies. NEJEN is a desktop shell, so the binary still calls `hyprctl`, `bash`, `git` and `notify-send`, and ships a handful of first-run shell scripts; `nejen doctor` reports which of them are present. The claim is narrower and more useful: the *tooling layer* carries no interpreter stack of its own.

### 3. Type-Safety and Stability
Go’s compiler catches errors at build time.
* Mistyped variables and wrong signatures fail the build, rather than failing at the moment a keybinding fires.
* Handling JSON and TOML goes through real parsers with real error values, instead of quoting rules that break the first time a wallpaper has a space in its name.
