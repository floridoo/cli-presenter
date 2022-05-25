package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const startDelay = 1 * time.Second
const endDelay = 2 * time.Second
const typeDelay = 50 * time.Millisecond
const enterDelay = 300 * time.Millisecond
const lineDelay = 1 * time.Second
const promptMarker = "###"
const prompt = `\$ `

func main() {
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Printf("Usage: %s <presentation file>\n", filepath.Base(os.Args[0]))
		fmt.Printf(`
	File syntax:

	# a comment
	! a description what is going on
	sleep 1
	$ execute me

			`)
		os.Exit(1)
	}
	input, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatal("can't read input file")
	}
	commands := strings.Split(string(input), "\n")

	inputReader, inputWriter := io.Pipe()
	outputReader, outputWriter := io.Pipe()
	waitChan := make(chan bool)

	write := func(str string) {
		inputWriter.Write([]byte(str))
	}
	writeLn := func(str string) {
		write(str + "\n")
		<-waitChan
	}
	typeText := func(w io.Writer, text string) {
		for _, letter := range text {
			delay := typeDelay/2 + time.Duration(rand.Intn(int(typeDelay)))
			if letter == ' ' {
				delay = delay * 2
			}
			time.Sleep(delay)
			w.Write([]byte(string(letter)))
		}
	}

	go func() {
		shouldOutput := false
		for {
			buf := make([]byte, 1024)
			n, err := outputReader.Read(buf)
			if err != nil {
				log.Fatal(err)
			}
			output := string(buf[:n])
			if strings.Contains(output, promptMarker) {
				if !shouldOutput {
					offset := strings.Index(output, promptMarker) + len(promptMarker) + len(prompt) + 2
					if offset >= len(output) {
						offset = len(output)
					}
					output = output[offset:]
				}
				shouldOutput = true
				output = strings.ReplaceAll(output, promptMarker, "")
				fmt.Print(output)
				waitChan <- true
			} else if shouldOutput {
				fmt.Print(output)
			}
		}
	}()

	go func() {
		// make bash exit on CTRL-C
		write(`trap "exit 255" SIGINT` + "\n")
		// change bash prompt
		writeLn(`PS1="` + promptMarker + prompt + `"`)
		time.Sleep(startDelay)
		for _, command := range commands {
			if len(command) == 0 {
				continue
			}
			commandParts := strings.SplitN(command, " ", 2)
			if len(commandParts) == 1 {
				commandParts = append(commandParts, "")
			}
			action, rest := commandParts[0], commandParts[1]
			switch action {
			case "#":
				continue
			case "!":
				fmt.Print("\033[1m\033[3m\033[33m") // bold, italic, orange
				fmt.Print("# ")
				typeText(os.Stdout, rest)
				fmt.Print("\033[0m") // normal
				time.Sleep(enterDelay)
				writeLn("")
			case "$":
				typeText(inputWriter, rest)
				time.Sleep(enterDelay)
				writeLn("")
			case "sleep":
				seconds, err := strconv.ParseFloat(rest, 64)
				if err != nil {
					log.Fatalf("unable to sleep duration %q", rest)
				}
				time.Sleep(time.Duration(seconds*1000) * time.Millisecond)
			default:
				log.Fatalf("unknown command %q", action)
			}
			time.Sleep(lineDelay)
		}
		time.Sleep(endDelay)
		inputWriter.Close()
	}()

	cmd := exec.Command("bash", "-i")
	cmd.Stdout = outputWriter
	cmd.Stderr = outputWriter
	cmd.Stdin = inputReader
	cmd.Run()
}
