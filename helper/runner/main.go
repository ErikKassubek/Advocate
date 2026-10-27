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

const advocatePath = "/home/advocate/Advocate/Advocate/advocate/"
const srcPath = "/home/advocate/Advocate/Experiments/Progs/PLDI/"

var progs = []string{
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

var now = time.Now()
var id = now.Format("2006-01-02-15-04-05")
var logName = "runner_" + id + ".log"

var modes = []string{
	// "GoPie",
	// "GFuzz",
	// "GoCR",
}

var settings = "-timeoutRun 30 -stats -noInfo -maxNumberElements 10000000"

var maxWorker = 2

var fileMutex sync.Mutex

func main() {
	readFlag()
	log("START RUNNING")
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, maxWorker)
	// for _, mode := range modes {
	for i, p := range progs {
		wg.Add(1)
		// m := mode
		path := filepath.Join(srcPath, p)
		go worker(path, p, &wg, sem, i == len(progs)-1)
		time.Sleep(time.Second)
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

func worker(path, name string, wg *sync.WaitGroup, sem chan struct{}, last bool) {
	defer wg.Done()
	sem <- struct{}{}

	log(fmt.Sprintf("START : %s", name))

	cmdStr := fmt.Sprintf("analysis -path %s %s", path, settings)

	cmd := exec.Command("./advocate", strings.Split(cmdStr, " ")...)

	// Set the working directory
	cmd.Dir = advocatePath

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log(fmt.Sprintf("ERROR : %s %s", name, err.Error()))
	}

	log(fmt.Sprintf("FINISH: %s", name))

	if !last {
		time.Sleep(30 * time.Minute)
	}

	<-sem
}

func log(c string) {
	fileMutex.Lock()
	defer fileMutex.Unlock()

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
