package main

import "fmt"
import "flag"

var dryRun bool
var baseCommit string
var localBranch string

func init() {
    flag.BoolVar(&dryRun, "x", true, "dry run")
    flag.StringVar(&baseCommit, "base", "master", "Base branch or commit")
}

func main() {
    flag.Parse()
    var argsTail = flag.Args()
    if len(argsTail) > 0 {
        localBranch = argsTail[0]
    }

    fmt.Println("Dry run:", dryRun)
    fmt.Println("Base commit:", baseCommit)
    fmt.Println("Local branch:", localBranch)
}

