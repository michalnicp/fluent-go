package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/michalnicp/fluent-go/syntax"
)

var version = "0.0.0"

var usage = `Usage: fluent [options] [file]...

Options:
  -h, -help      Print this message and exit.
  -v, -version   Print the version and exit.`

func main() {
	var code int
	defer func() { os.Exit(code) }()

	var (
		helpRequested    bool
		versionRequested bool
	)

	flag.BoolVar(&helpRequested, "help", false, "")
	flag.BoolVar(&helpRequested, "h", false, "")
	flag.BoolVar(&versionRequested, "version", false, "")
	flag.BoolVar(&versionRequested, "v", false, "")
	flag.Parse()

	if helpRequested {
		fmt.Println(usage)
		return
	}

	if versionRequested {
		fmt.Println(version)
		return
	}

	if len(flag.Args()) == 0 {
		fmt.Println(usage)
		code = 1
		return
	}

	for _, file := range flag.Args() {
		input, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("read %s: %v\n", file, err)
			code = 1
		}

		if _, err := syntax.Parse(input); err != nil {
			fmt.Printf("parse %s:\n%+v\n", file, err)
			code = 1
		}
	}
}
