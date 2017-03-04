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
var interactive_mode bool

func init() {
	flag.StringVar(&baseCommit, "base", "master", "Base branch or commit")
	flag.BoolVar(&remove, "d", false, "Remove branch if it is safe")
	flag.BoolVar(&interactive_mode, "i", false, "Travers all local branch interactively")
}

// funtion returns all local branches which do not have upstream
func getLocalBranches() (branches []string, err error) {
	cmd := exec.Command("git", "symbolic-ref", "-q", "--short", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out

	curr_branch := ""
	err = cmd.Run()
	if err == nil {
		curr_branch = strings.Split(out.String(), "\n")[0]
	}

	// this super easy format was choosen to exclude any unexpected symbols in output
	cmd = exec.Command("git", "for-each-ref", "--format=%(refname:short)", "refs/heads/")
	cmd.Stdout = &out

	err = cmd.Run()
	if err != nil {
		return branches, err
	}

	all_branches := strings.Split(out.String(), "\n")

	for i := 0; i < len(all_branches); i++ {
		if curr_branch == all_branches[i] || 0 == len(all_branches[i]) {
			continue
		}
		cmd = exec.Command("git", "rev-parse", "--symbolic-full-name", all_branches[i]+"@{u}")
		loc_err := cmd.Run()
		if loc_err != nil { // means there is no upstream branch
			branches = append(branches, all_branches[i])
		}
	}

	return branches, err
}

func getMergeBase(c1 string, c2 string) (merge_base string, err error) {
	cmd := exec.Command("git", "merge-base", c1, c2)
	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Run()
	if err != nil {
		return merge_base, err
	}

	merge_base = strings.Trim(strings.SplitN(out.String(), "\n", 1)[0], " \n")
	if merge_base == "" {
		return "", errors.New("Cannot define merge_base")
	}

	return merge_base, err
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

func getPatchId(commit string) (patchId string, err error) {
	cmd_show := exec.Command("git", "show", commit)
	cmd_patch_id := exec.Command("git", "patch-id", "--stable")

	var out bytes.Buffer
	reader, writer := io.Pipe()
	cmd_show.Stdout = writer
	cmd_patch_id.Stdin = reader
	cmd_patch_id.Stdout = &out

	cmd_show.Start()
	cmd_patch_id.Start()
	cmd_show.Wait()
	writer.Close()
	err = cmd_patch_id.Wait()

	if err != nil {
		return patchId, err
	}

	patchId = strings.Split(out.String(), " ")[0]

	return patchId, err
}

func getPatchIds(commits []string) (patchId map[string]string, err error) {
	patchId = make(map[string]string, len(commits))

	for _, commit := range commits {
		pId, err := getPatchId(commit)
		if err != nil {
			return patchId, err
		}
		patchId[pId] = commit
	}

	return patchId, err
}

func removeBranch(branch string) (err error) {
	cmd := exec.Command("git", "branch", "-D", branch)

	err = cmd.Run()
	return err
}

func integrated(branch string, baseCommit string) (safeToRemove bool, err error) {
	mergeBase, err := getMergeBase(baseCommit, branch)
	if err != nil {
		return false, err
	}

	localPatchIds, err := getPatchIds(getCommits(mergeBase, branch))
	if err != nil {
		return false, err
	}
	if len(localPatchIds) == 0 {
		return true, nil
	}

	commits := getCommits(mergeBase, baseCommit)

	// walk from merge-base to HEAD, usually old branches are faster to
	// find this way
	for i := len(commits) - 1; i >= 0; i-- {
		commit := commits[i]
		pId, err := getPatchId(commit)
		if err != nil {
			return false, err
		}

		if _, ok := localPatchIds[pId]; ok {
			return true, nil
		}
	}

	return false, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()
	var argsTail = flag.Args()
	var brachToRemove string
	if len(argsTail) > 0 {
		brachToRemove = argsTail[0]
	}

	localBranches, err := getLocalBranches()
	if err != nil {
		log.Fatalln("ERROR:", err)
	}

	branchExists := false
	safeToRemove := false
	for _, b := range localBranches {
		if interactive_mode {
			safeToRemove, err = integrated(b, baseCommit)
			if err != nil {
				log.Fatalln("ERROR:", err)
			}
			if safeToRemove {
				fmt.Print("[" + b + "] is safe to remove. Remove? [Y/n] ")
				var input string
				fmt.Scanln(&input)
				if input == "Y" || input == "y" {
					err = removeBranch(b)
					if err != nil {
						log.Fatalln("ERROR:", err)
					}
					fmt.Println("[" + b + "] removed")
				}
			} else {
				fmt.Println("[" + b + "] is not safe to remove, skip it")
			}
		} else if b == brachToRemove {
			branchExists = true
			safeToRemove, err = integrated(brachToRemove, baseCommit)
			if err != nil {
				log.Fatalln("ERROR:", err)
			}
		}
	}

	if interactive_mode {
		os.Exit(0)
	}

	if !branchExists {
		log.Fatal("ERROR: branch doesn't exist")
	}

	if safeToRemove {
		fmt.Println("[" + brachToRemove + "] is safe to remove")
	} else {
		fmt.Println("[" + brachToRemove + "] is not in base")
	}

	if remove {
		if safeToRemove {
			err = removeBranch(brachToRemove)
			if err != nil {
				log.Fatalln("ERROR:", err)
			}
		} else {
			log.Fatalln("ERROR: branch '" + brachToRemove + "' is not safe to remove")
		}
	}
}
