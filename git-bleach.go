package main

import "fmt"
import "flag"
import "os/exec"
import "bytes"
import "strings"
import "os"

var dryRun bool
var baseCommit string
var localBranch string

func init() {
    flag.BoolVar(&dryRun, "x", true, "dry run")
    flag.StringVar(&baseCommit, "base", "master", "Base branch or commit")
}

// funtion returns all local branches which do not have upstream
func getLocalBranches() (branches []string, err error) {
    // this super easy format was choosen to exclude any unexpected symbols in output
    cmd := exec.Command("git", "for-each-ref", "--format=%(refname:short)", "refs/heads/")
    var out bytes.Buffer
    cmd.Stdout = &out

    err = cmd.Run()
    if err != nil {
        return branches, err
    }

    all_branches := strings.Split(out.String(), "\n")

    for i := 0; i < len(all_branches); i++ {
        cmd = exec.Command("git", "rev-parse", "--symbolic-full-name", all_branches[i] + "@{u}")
        loc_err := cmd.Run()
        if loc_err != nil {     // means there is no upstream branch
            branches = append(branches, all_branches[i])
        }
    }

    return branches, err
}

func main() {
    flag.Parse()
    var argsTail = flag.Args()
    if len(argsTail) > 0 {
        localBranch = argsTail[0]
    }

    branches, err := getLocalBranches()
    if err != nil {
        fmt.Println("ERROR:", err)
        os.Exit(1)
    }

    fmt.Println("Dry run:", dryRun)
    fmt.Println("Base commit:", baseCommit)
    fmt.Println("Local branch:", localBranch)
    fmt.Println("Local branches:", branches)
}
