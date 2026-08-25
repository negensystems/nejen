package main

type CommandFunc func(args []string)

var commands = map[string]CommandFunc{}

func registerCommand(name string, f CommandFunc) {
	commands[name] = f
}

func runCommand(name string, args []string) bool {
	if f, ok := commands[name]; ok {
		f(args)
		return true
	}
	return false
}

func listCommands() []string {
	var list []string
	for name := range commands {
		list = append(list, name)
	}
	return list
}
