package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func main() {
	create := flag.Bool("create", false, "Create tags")
	push := flag.Bool("push", false, "Push tags to remote")
	delete := flag.Bool("delete", false, "Delete tags locally and remotely")
	flag.Parse()

	if !*create && !*push && !*delete {
		fmt.Println("Usage: go run . --create|--push|--delete <version>")
		fmt.Println("Example: go run . --create v0.2.0")
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) < 1 || args[0] == "" {
		fmt.Println("Error: version is required")
		fmt.Println("Example: go run . --create v0.2.0")
		os.Exit(1)
	}

	version := args[0]
	modules := readModulesFromGoWork()

	if len(modules) == 0 {
		fmt.Println("Warning: no modules found in go.work")
	}

	if *create {
		createTags(version, modules)
	}

	if *push {
		pushTags(version, modules)
	}

	if *delete {
		deleteTags(version, modules)
	}
}

func readModulesFromGoWork() []string {
	file, err := os.Open("go.work")
	if err != nil {
		fmt.Printf("Error opening go.work: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var modules []string
	scanner := bufio.NewScanner(file)
	inUseBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "use (") {
			inUseBlock = true
			continue
		}

		if inUseBlock && line == ")" {
			inUseBlock = false
			continue
		}

		if inUseBlock {
			if idx := strings.Index(line, "//"); idx != -1 {
				line = strings.TrimSpace(line[:idx])
			}

			line = strings.Trim(line, "\"")

			if line == "" {
				continue
			}

			module := strings.TrimPrefix(line, "./")

			if strings.HasPrefix(module, "_tools/") {
				continue
			}

			modules = append(modules, module)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading go.work: %v\n", err)
		os.Exit(1)
	}

	return modules
}

func createTags(version string, modules []string) {
	fmt.Printf("Creating tag %s for root...\n", version)
	mustRun("git", "tag", version)

	runParallel(modules, 5, func(mod string) {
		tag := fmt.Sprintf("%s/%s", mod, version)
		fmt.Printf("Creating tag %s...\n", tag)
		mustRun("git", "tag", tag)
	})

	fmt.Println("All tags created successfully.")
}

func pushTags(version string, modules []string) {
	fmt.Printf("Pushing tag %s...\n", version)
	mustRun("git", "push", "origin", version)

	runParallel(modules, 5, func(mod string) {
		tag := fmt.Sprintf("%s/%s", mod, version)
		fmt.Printf("Pushing tag %s...\n", tag)
		mustRun("git", "push", "origin", tag)
	})

	fmt.Println("All tags pushed successfully.")
}

func deleteTags(version string, modules []string) {
	fmt.Printf("Deleting local tag %s...\n", version)
	run("git", "tag", "-d", version)
	fmt.Printf("Deleting remote tag %s...\n", version)
	run("git", "push", "origin", ":refs/tags/"+version)

	runParallel(modules, 5, func(mod string) {
		tag := fmt.Sprintf("%s/%s", mod, version)
		fmt.Printf("Deleting local tag %s...\n", tag)
		run("git", "tag", "-d", tag)
		fmt.Printf("Deleting remote tag %s...\n", tag)
		run("git", "push", "origin", ":refs/tags/"+tag)
	})

	fmt.Println("Tag deletion completed (some tags may not have existed).")
}

// runParallel executes fn for each module with a maximum concurrency limit
func runParallel(modules []string, maxConcurrency int, fn func(string)) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrency)

	for _, mod := range modules {
		wg.Add(1)
		go func(m string) {
			defer wg.Done()
			semaphore <- struct{}{}        // acquire
			defer func() { <-semaphore }() // release
			fn(m)
		}(mod)
	}

	wg.Wait()
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: command %s %v failed: %v\n", name, args, err)
	}
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
