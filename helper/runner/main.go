package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var advocatePath = "/home/advocate/Advocate/Advocate/advocate"
var srcPath = "/home/advocate/Advocate/Experiments/Progs/PLDIClosure/"

var progs = []string{
	"argo-cd",
	"bleve",
	"bosun",
	"caddy",
	// "dns",
	// "etcd",
	// "fabiolb",
	// "frp",
	// "gin",
	// "go-ethereum",
	// "gorums",
	"grpc",
	"kubernetes",
	// "moby",
	// "nsq",
	"ollama",
	// "pholcus",
	// "prometheus",
	// "syncthing",
	// "terraform",
	// "zinx",
}

var consts = []string{
	"scpc", "hb", "order", "replay",
}

var stable = []bool{
	true, false,
}

var now = time.Now()
var id = now.Format("2006-01-02-15-04-05")
var closure string
var logName string

var modes = []string{
	// "GoPie",
	// "GFuzz",
	// "GoCR",
}

var settings = "-timeoutRun 30 -timeoutFuz 420 -stats -noInfo -maxNumberElements 10000000"

var maxWorker = 3

var fileMutex sync.Mutex

func main() {
	readFlag()
	log("START RUNNING")
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, maxWorker)
	// for _, mode := range modes {
	for i, p := range progs {
		for _, s := range stable {
			for _, c := range consts {
				wg.Add(1)
				// m := mode

				path := srcPath
				if s {
					path = filepath.Join(path, "stable")
				} else {
					path = filepath.Join(path, "nonStable")
				}

				path = filepath.Join(path, c)
				path = filepath.Join(path, p)

				go worker(path, p, c, s, &wg, sem, i == len(progs)-1)
			}
		}
	}
	// }
	wg.Wait()
	log("END RUNNING")
}

func readFlag() {
	valuesFlag := flag.String("p", "", "Space-separated list of progs")
	threadsFlags := flag.String("w", "", "Number of progs run parallel")

	// Parse the flags
	flag.Parse()

	// Split the input string by spaces
	if *valuesFlag != "" {
		progs = strings.Fields(*valuesFlag)
	}

	if *threadsFlags != "" {
		var err error
		maxWorker, err = strconv.Atoi(*threadsFlags)
		if err != nil {
			fmt.Println("Error:", err)
			panic("Invalid value t")
		}
	}

	log("Run: " + id + " " + *valuesFlag + " with " + *threadsFlags + " workers")
}

func worker(path, prog string, closure string, stable bool, wg *sync.WaitGroup, sem chan struct{}, last bool) {
	defer wg.Done()
	sem <- struct{}{}

	log(fmt.Sprintf("START : %s", prog))

	cmdStr := fmt.Sprintf("fuzzing -path %s -closure %s %s", path, closure, settings)

	if !stable {
		cmdStr += " -disableStableID"
	}

	cmd := exec.Command("./advocate", strings.Split(cmdStr, " ")...)

	println(cmd.String())

	// Set the working directory
	cmd.Dir = advocatePath

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// err := cmd.Run()
	// if err != nil {
	// 	log(fmt.Sprintf("ERROR : %s %s", prog, err.Error()))
	// }

	log(fmt.Sprintf("FINISH: %s", prog))

	if !last {
		time.Sleep(20 * time.Minute)
	}

	<-sem
}

func log(c string) {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	logName = "runner_" + id + ".log"

	file, err := os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	now := time.Now().Format("2006/01/02 15:04:05")

	_, err = file.WriteString(now + " " + c + "\n")

	fmt.Println(now + " " + c)
}
