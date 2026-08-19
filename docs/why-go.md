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

### 1. Sub-Millisecond Performance
Go compiles directly into native CPU machine instructions.
* It does not spawn separate interpreters or shell wrapper processes.
* Commands (like rendering themes, listing configurations, or updating statuses) execute in **under 1ms** (microsecond scale).
* This eliminates micro-stutters and maximizes laptop battery efficiency.

### 2. Zero-Dependency Portability
Go builds a single static binary.
* It does not require Python, PyYAML, tabulate, `awk`, or `sed` to run on the system.
* Upgrading packages on Arch Linux will never break the core environment manager.

### 3. Type-Safety and Stability
Go’s compiler catches errors at build time.
* It prevents runtime crashes due to mistyped variables or unhandled errors.
* Hashing filesystem directories and handling JSON/TOML data is safe, standard, and robust.
