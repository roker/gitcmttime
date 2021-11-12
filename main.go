package main

import (
	"flag"
	"fmt"
	"gitcmttime/repository"
	"os"
	"strings"
)

func main() {
	pathRepository := flag.String("repo", "", "path to git repository")
	dttype := flag.String("type", "author", `datetime to use, possible values "author", "committer"`)
	outtype := flag.String("output", "short", `type of output, possible values "short", "long", "errors"`)
	flag.Parse()

	if "" == *pathRepository {
		fmt.Println("Usage: gitcmttime --repo path [--type=<type>] [--output=<output type>]  ")
		flag.PrintDefaults()
		os.Exit(1)
	}

	r, err := repository.OpenRepository(*pathRepository, getDTType(*dttype))
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(2)
	}

	r.PrintOutput(getOutputType(*outtype))
}

func getOutputType(outtype string) repository.TypeOutput {
	outtype = strings.ToLower(outtype)

	switch outtype {
	case "short":
		return repository.TypeOutShort
	case "long":
		return repository.TypeOutLong
	case "errors":
		return repository.TypeOutErrors
	}

	fmt.Println(`--type must be either "short", "long" or "errors"`)
	os.Exit(1)
	return 0
}

func getDTType(dttype string) repository.TypeDateTime {
	dttype = strings.ToLower(dttype)

	switch dttype {
	case "author":
		return repository.TypeDTAuthor
	case "committer":
		return repository.TypeDTCommitter
	}

	fmt.Println(`--type must be either "author" or "committer"`)
	os.Exit(1)
	return 0
}
