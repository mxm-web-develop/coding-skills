package main

import (
	"errors"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "version", "--version", "-v":
		fmt.Printf("flowctl %s\n", packVersion)
		return
	case "doctor":
		err = runDoctor(os.Args[2:])
	case "status":
		err = runStatus(os.Args[2:])
	case "work":
		err = runWork(os.Args[2:])
	case "checkpoint":
		err = runCheckpoint(os.Args[2:])
	case "evidence":
		err = runEvidence(os.Args[2:])
	case "cleanup":
		err = runCleanup(os.Args[2:])
	case "validate":
		err = runValidate(os.Args[2:])
	case "project":
		if len(os.Args) < 3 || os.Args[2] != "init" {
			err = errors.New("usage: flowctl project init [--root PATH] --mode greenfield|existing --name NAME")
		} else {
			err = runProjectInit(os.Args[3:])
		}
	case "render-board":
		err = runRenderBoard(os.Args[2:])
	case "lint-message":
		err = runLintMessage(os.Args[2:])
	case "help", "--help", "-h":
		usage()
		return
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "flowctl: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`flowctl manages the deterministic state of an AI Flow project.

Usage:
  flowctl version
  flowctl doctor [--root PATH] [--json]
  flowctl project init [--root PATH] --mode greenfield|existing --name NAME
  flowctl status [--root PATH] [--json]
  flowctl work <create|list|show|ready|start|block|review-ready|complete|cancel>
  flowctl checkpoint <save|list|show|latest|resume>
  flowctl evidence <run|record|list|show|verify>
  flowctl cleanup digest [--root PATH] --plan PATH
  flowctl validate [--root PATH] [--json]
  flowctl render-board [--root PATH]
  flowctl lint-message [--file PATH]`)
}
