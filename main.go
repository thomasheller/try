package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/dustin/go-humanize/english"
)

const (
	ConstantBackoff BackoffStrategy = iota
	LinearBackoff
	ExponentialBackoff
)

type BackoffStrategy int

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: try [strategy] [command] [arg]...")
		fmt.Println("strategies: constant (default), linear, exponential")
		os.Exit(1)
	}

	var strategy BackoffStrategy
	var shiftArgs bool

	switch os.Args[1] {
	case "constant":
		strategy = ConstantBackoff
		shiftArgs = true
	case "linear":
		strategy = LinearBackoff
		shiftArgs = true
	case "exponential":
		strategy = ExponentialBackoff
		shiftArgs = true
	default:
		strategy = ConstantBackoff
		shiftArgs = false
	}

	if shiftArgs && len(os.Args) == 2 {
		ShowExample(strategy)
		os.Exit(1)
	}

	var command []string

	if shiftArgs {
		command = os.Args[2:]
	} else {
		command = os.Args[1:]
	}

	attempts := 0
	firstStart := time.Now()

	var lastStart time.Time

	for {
		attempts++
		lastStart = time.Now()

		var ctx context.Context
		ctx = context.Background()
		cmd := exec.CommandContext(ctx, command[0], command[1:]...)

		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin

		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}

		err := cmd.Wait()
		if err == nil {
			break
		}

		pause := Backoff(strategy, attempts)

		log.Printf("[try] Command failed after %s (started trying %v), trying again %v...",
			english.Plural(attempts, "attempt", ""),
			humanize.Time(firstStart),
			humanize.Time(time.Now().Add(pause)))

		time.Sleep(pause)
	}

	log.Printf("[try] Command succeeded after %s (took %v, started trying %v)",
		english.Plural(attempts, "attempt", ""),
		time.Since(lastStart),
		humanize.Time(firstStart))
}

func ShowExample(strategy BackoffStrategy) {
	const retries = 20

	fmt.Printf("Example for the first %d retries:\n", retries)

	pauseTotal := time.Now()
	pause := time.Duration(0)

	for i := 1; i <= retries; i++ {
		fmt.Printf("%v attempt after %v (%v)\n", humanize.Ordinal(i), pause, humanize.Time(pauseTotal))
		pause = Backoff(strategy, i)
		pauseTotal = pauseTotal.Add(pause)
	}

	fmt.Println("...")
}

func Backoff(strategy BackoffStrategy, attempt int) time.Duration {
	switch strategy {
	case ConstantBackoff:
		return time.Duration(2) * time.Second
	case LinearBackoff:
		return time.Duration(2*attempt) * time.Second
	case ExponentialBackoff:
		return time.Duration(2*attempt*attempt) * time.Second
	}

	panic("Unsupported backoff strategy")
}
