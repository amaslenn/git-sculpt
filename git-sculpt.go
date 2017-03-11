package main

import "fmt"
import "flag"
import "os/exec"
import "bytes"
import "strings"
import "errors"
import "io"
import "os"
import "log"

var baseCommit string
var remove bool
var interactiveMode bool
var removeAll bool

func init() {
	flag.StringVar(&baseCommit, "base", "master", "Base branch or commit")
	flag.BoolVar(&remove, "d", false, "Remove branch if it is safe")
	flag.BoolVar(&interactiveMode, "i", false, "Travers all local branch interactively")
	flag.BoolVar(&removeAll, "all", false, "Remove all branches if it is safe")
}

// funtion returns all local branches which do not have upstream
func getLocalBranches() (branches []string, err error) {
	cmd := exec.Command("git", "symbolic-ref", "-q", "--short", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out

	currBranch := ""
	err = cmd.Run()
	if err == nil {
		currBranch = strings.Split(out.String(), "\n")[0]
	}

	// this super easy format was choosen to exclude any unexpected symbols in output
	cmd = exec.Command("git", "for-each-ref", "--format=%(refname:short)", "refs/heads/")
	cmd.Stdout = &out

	err = cmd.Run()
	if err != nil {
		return branches, errors.New("error getting local branches")
	}

	allBranches := strings.Split(out.String(), "\n")

	for i := 0; i < len(allBranches); i++ {
		if currBranch == allBranches[i] || 0 == len(allBranches[i]) {
			continue
		}
		cmd = exec.Command("git", "rev-parse", "--symbolic-full-name", allBranches[i]+"@{u}")
		locErr := cmd.Run()
		if locErr != nil { // means there is no upstream branch
			branches = append(branches, allBranches[i])
		}
	}

	return branches, err
}

func getMergeBase(c1 string, c2 string) (mergeBase string, err error) {
	cmd := exec.Command("git", "merge-base", c1, c2)
	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Run()
	if err != nil {
		return mergeBase, errors.New("error getting merge-base")
	}

	mergeBase = strings.Trim(strings.SplitN(out.String(), "\n", 1)[0], " \n")
	if mergeBase == "" {
		return "", errors.New("Cannot define mergeBase")
	}

	return mergeBase, err
}

func getCommits(c1 string, c2 string) (commits []string) {
	cmd := exec.Command("git", "log", "--format=%H", c1+".."+c2)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return commits
	}

	if out.String() == "" {
		return commits
	}

	for _, commit := range strings.Split(out.String(), "\n") {
		commit = strings.Trim(commit, "\n")
		if commit != "" {
			commits = append(commits, commit)
		}
	}

	return commits
}

func getPatchID(commit string) (patchID string, err error) {
	cmdShow := exec.Command("git", "show", commit)
	cmdPatchID := exec.Command("git", "patch-id", "--stable")

	var out bytes.Buffer
	reader, writer := io.Pipe()
	cmdShow.Stdout = writer
	cmdPatchID.Stdin = reader
	cmdPatchID.Stdout = &out

	cmdShow.Start()
	cmdPatchID.Start()
	cmdShow.Wait()
	writer.Close()
	err = cmdPatchID.Wait()

	if err != nil {
		return patchID, errors.New("error calculating patch-id")
	}

	patchID = strings.Split(out.String(), " ")[0]

	return patchID, err
}

func getPatchIDs(commits []string) (patchID map[string]string, err error) {
	patchID = make(map[string]string, len(commits))

	for _, commit := range commits {
		pID, err := getPatchID(commit)
		if err != nil {
			return patchID, err
		}
		patchID[pID] = commit
	}

	return patchID, err
}

func removeBranch(branch string) (err error) {
	cmd := exec.Command("git", "branch", "-D", branch)

	err = cmd.Run()
	if err != nil {
		return errors.New("error deleting branch `" + branch + "'")
	}

	return nil
}

func integrated(branch string, baseCommit string) (safeToRemove bool, err error) {
	mergeBase, err := getMergeBase(baseCommit, branch)
	if err != nil {
		return false, err
	}

	localPatchIDs, err := getPatchIDs(getCommits(mergeBase, branch))
	if err != nil {
		return false, err
	}
	if len(localPatchIDs) == 0 {
		return true, nil
	}

	commits := getCommits(mergeBase, baseCommit)

	// walk from merge-base to HEAD, usually old branches are faster to
	// find this way
	for i := len(commits) - 1; i >= 0; i-- {
		commit := commits[i]
		pID, err := getPatchID(commit)
		if err != nil {
			return false, err
		}

		if _, ok := localPatchIDs[pID]; ok {
			return true, nil
		}
	}

	return false, nil
}

func removeSingleBranch(branch string, base string) (err error) {
	safeToRemove, err := integrated(branch, base)
	if err != nil {
		return err
	}

	if safeToRemove {
		fmt.Println("[" + branch + "] is safe to remove")
	} else {
		fmt.Println("[" + branch + "] is not in base")
	}

	if remove {
		if safeToRemove {
			err = removeBranch(branch)
			if err != nil {
				return err
			}
		} else {
			return errors.New("branch '" + branch + "' is not safe to remove")
		}
	}

	return nil
}

func interactiveRemove(localBranches []string, baseCommit string) (err error) {
	safeToRemove := false
	for _, b := range localBranches {
		safeToRemove, err = integrated(b, baseCommit)
		if err != nil {
			return err
		}
		if safeToRemove {
			fmt.Print("[" + b + "] is safe to remove. Remove? [Y/n] ")
			var input string
			fmt.Scanln(&input)
			if input == "Y" || input == "y" {
				err = removeBranch(b)
				if err != nil {
					return err
				}
				fmt.Println("[" + b + "] removed")
			}
		} else {
			fmt.Println("[" + b + "] is not safe to remove, skip it")
		}
	}

	return nil
}

func removeAllBranches(localBranches []string, baseCommit string) (err error) {
	var branchesToRemove []string
	var branchesToKeep []string

	for _, b := range localBranches {
		safeToRemove, err := integrated(b, baseCommit)
		if err != nil {
			return err
		}
		if safeToRemove {
			branchesToRemove = append(branchesToRemove, b)
		} else {
			branchesToKeep = append(branchesToKeep, b)
		}
	}

	fmt.Println("Branches to be removed:", strings.Join(branchesToRemove, ", "))
	fmt.Println("Branches to keep:", strings.Join(branchesToKeep, ", "))

	for _, b := range branchesToRemove {
		err = removeBranch(b)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	log.SetFlags(log.Lshortfile)

	flag.Parse()
	var argsTail = flag.Args()
	var branchToRemove string
	if len(argsTail) > 0 {
		branchToRemove = argsTail[0]
	}

	if (removeAll || interactiveMode) && branchToRemove == "" {
		localBranches, err := getLocalBranches()
		if err != nil {
			log.Fatalln("ERROR:", err)
		}

		if removeAll {
			err = removeAllBranches(localBranches, baseCommit)
		} else {
			err = interactiveRemove(localBranches, baseCommit)
		}

		if err != nil {
			log.Fatalln("ERROR:", err)
		}
	} else if branchToRemove != "" {
		err := removeSingleBranch(branchToRemove, baseCommit)
		if err != nil {
			log.Fatalln("ERROR:", err)
		}
	} else if !interactiveMode {
		fmt.Println("Nothing to do")
	}

	os.Exit(0)
}
