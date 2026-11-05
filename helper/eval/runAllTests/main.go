package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

const mainPath = "~/Uni/HiWi/Advocate/Experiments/Progs/PLDI/"

var logMutex = sync.Mutex{}

func main() {
	names := []string{
		// "canonicalTests",
		"argo-cd",
		"bleve",
		"bosun",
		"caddy",
		"dns",
		"etcd",
		"fabiolb",
		"frp",
		"gin",
		"go-ethereum",
		"gorums",
		"grpc",
		"kubernetes",
		"moby",
		"nsq",
		"ollama",
		"pholcus",
		"prometheus",
		"syncthing",
		"terraform",
		"zinx",
	}

	if err := os.Chdir("../../toolchain"); err != nil {
		fmt.Printf("Failed to change directory: %v\n", err)
		return
	}

	wd, _ := os.Getwd()
	fmt.Printf("Changed working directory to: %s\n\n", wd)

	analysisTimeout := "1800" // 1800s = 0.5h
	replayTimeout := "-1"     // 100 * recording time
	maxProgAtSameTime := 2

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxProgAtSameTime)

	for _, name := range names {
		wg.Add(1)
		go runProg(name, mainPath, analysisTimeout, replayTimeout, &wg, sem)
		time.Sleep(2 * time.Second)
	}

	wg.Wait()

}

func runProg(name, mainPath, analysisTimeout string, replayTimeout string, wg *sync.WaitGroup, sem chan struct{}) {

	defer wg.Done()
	defer func() { <-sem }()
	sem <- struct{}{}

	fmt.Println("\n\nRun prog ", name)
	logInfo(name, "Start")
	path := filepath.Join(mainPath, name)

	cmd := exec.Command("./tool", "test", "-a", "~/Uni/HiWi/ADVOCATE", "-f", path, "-m", "-s", "-t", "-N", name, "-T", analysisTimeout, "-R", replayTimeout)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println(cmd.String())
	err := cmd.Run()

	if err != nil {
		fmt.Println("Could not run ", name, err)
		file, err2 := os.OpenFile("failed.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err2 != nil {
			fmt.Println("Error opening file:", err2)
			return
		}

		_, err2 = file.WriteString("# " + name)
		if err2 != nil {
			fmt.Println("Error writing string:", err2)
			file.Close()
			return
		}

		_, err2 = file.WriteString(err.Error())
		if err2 != nil {
			fmt.Println("Error writing string:", err2)
			file.Close()
			return
		}
		file.Close()
		logInfo(name, err.Error())
	}

	fmt.Println("Finished running", name)
	logInfo(name, "End")
}

func logInfo(prog string, message string) {
	logMutex.Lock()
	defer logMutex.Unlock()

	fileName := "./../runAllTests.log"

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("%s: %s - %s\n", currentTime, prog, message)

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Write the message to the file
	if _, err := file.WriteString(logMessage); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}
}
