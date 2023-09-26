package main

import (
	_ "embed"
	"github.com/rixocz/azure-law-sender/cmd"
	"github.com/rixocz/azure-law-sender/version"
	"log"
)

//go:embed version.txt
var versionFile string

func main() {
	version.AppVersion = versionFile
	if err := cmd.NewRootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
