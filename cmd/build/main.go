package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func walkFunc(path string, info fs.FileInfo, err error) error {
	if !info.IsDir() {
		return nil
	}
	if info.Name() == "lambdas" {
		return nil
	}

	fmt.Println("building path", path)

	cmd := exec.Command("go", "build", "-o", "bootstrap")
	cmd.Dir = path
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running command: %w", err)
	}

	if exitCode := cmd.ProcessState.ExitCode(); exitCode != 0 {
		return fmt.Errorf("non-zero exit code: %v", exitCode)
	}

	return nil
}

func main() {
	if err := filepath.Walk("lambdas", walkFunc); err != nil {
		log.Panic("failed to build lambdas in lambdas dir")
	}
}
