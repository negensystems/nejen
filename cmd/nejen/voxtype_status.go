package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

func init() {
	registerCommand("voxtype status", runVoxtypeStatus)
}

func runVoxtypeStatus(args []string) {
	if _, err := exec.LookPath("voxtype"); err != nil {
		fmt.Println(`{"alt": "", "tooltip": ""}`)
		os.Exit(0)
	}

	cmd := exec.Command("voxtype", "status", "--follow", "--extended", "--format", "json")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		os.Exit(1)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(line), &data); err == nil {
			if classVal, ok := data["class"].(string); ok {
				data["alt"] = classVal
			}
			if out, err := json.Marshal(data); err == nil {
				fmt.Println(string(out))
			}
		}
	}

	cmd.Wait()
}
