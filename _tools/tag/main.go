package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "" {
		fmt.Println("Usage: make tag vX.Y.Z")
		os.Exit(1)
	}

	version := os.Args[1]
	modules := os.Args[2:]

	// Tag root
	fmt.Printf("Creating tag %s for root...\n", version)
	mustRun("git", "tag", version)

	// Tag modules
	for _, mod := range modules {
		tag := fmt.Sprintf("%s/%s", mod, version)
		fmt.Printf("Creating tag %s...\n", tag)
		mustRun("git", "tag", tag)
	}

	// Push tags
	fmt.Printf("Pushing tag %s...\n", version)
	mustRun("git", "push", "origin", version)

	for _, mod := range modules {
		tag := fmt.Sprintf("%s/%s", mod, version)
		fmt.Printf("Pushing tag %s...\n", tag)
		mustRun("git", "push", "origin", tag)
	}

	fmt.Println("All tags created and pushed successfully.")
}

func mustRun(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running command %s %v: %v\n", name, args, err)
		os.Exit(1)
	}
}
