package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/leminhnguyenai/personal-blog/internal"
)

var (
	debugMode bool
	dirPath   string
)

func init() {
	// Set up a random free port if none is specified in .env
	if os.Getenv("PORT") == "" {
		a, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			log.Fatal(err)
		}

		l, err := net.ListenTCP("tcp", a)
		defer l.Close()
		if err != nil {
			log.Fatal(err)
		}

		os.Setenv("PORT", strconv.Itoa(l.Addr().(*net.TCPAddr).Port))
	}

	parseFlags(os.Args[1:])
}

// NOTE: Preview mode is disable for now, no plan to bring back in the future
func parseFlags(args []string) {
	// Check for debug flag
	flag.BoolVar(&debugMode, "d", false, "Debug mode")

	if err := flag.CommandLine.Parse(args); err != nil {
		log.Fatal(err)
	}

	if len(flag.CommandLine.Args()) == 0 {
		log.Fatalf("No path specified\n")
	}

	dirPath = flag.CommandLine.Args()[0]
}

func main() {
	srv, err := internal.NewServer(debugMode)
	if err != nil {
		log.Fatal(err)
	}

	if err = srv.Construct(dirPath); err != nil {
		log.Fatal(err)
	}

	srv.Start()
}
