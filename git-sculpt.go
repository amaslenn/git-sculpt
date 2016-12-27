package main

import "fmt"
import "flag"
import "os/exec"
import "bytes"
import "strings"
import "errors"
import "io"
import "log"

var baseCommit string
var localBranch string
var remove bool

func init() {
	flag.StringVar(&baseCommit, "base", "master", "Base branch or commit")
	flag.BoolVar(&remove, "d", false, "Remove branch if it is safe")
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

func main() {
	flag.Parse()
	var argsTail = flag.Args()
	if len(argsTail) > 0 {
		localBranch = argsTail[0]
	}

	branches, err := getLocalBranches()
	if err != nil {
		log.Fatalln("ERROR:", err)
	}

	var mergeBase string
	branchExists := false
	safeToRemove := false
	for _, b := range branches {
		if b == localBranch {
			branchExists = true
			mergeBase, err = getMergeBase(baseCommit, localBranch)
			if err != nil {
				log.Fatalln("ERROR:", err)
			}

			localPatchIds, err := getPatchIds(getCommits(mergeBase, localBranch))
			if err != nil {
				log.Fatalln("ERROR:", err)
			}
			if len(localPatchIds) == 0 {
				safeToRemove = true
				break
			}

			commits := getCommits(mergeBase, baseCommit)

			found := 0
			// walk from merge-base to HEAD, usually old branches are faster to
			// find this way
			for i := len(commits) - 1; i >= 0; i-- {
				commit := commits[i]
				pId, err := getPatchId(commit)
				if err != nil {
					log.Fatalln("ERROR:", err)
				}

				if _, ok := localPatchIds[pId]; ok {
					found++
				}

				if found == len(localPatchIds) {
					safeToRemove = true
					break
				}
			}
			break
		}
	}

	if !branchExists {
		log.Fatal("ERROR: branch doesn't exist")
	}

	if safeToRemove {
		fmt.Println("[" + localBranch + "] is safe to remove")
	} else {
		fmt.Println("[" + localBranch + "] is not in base")
	}

	if safeToRemove && remove {
		err = removeBranch(localBranch)
		if err != nil {
			log.Fatalln("ERROR:", err)
		} else {
			fmt.Println("[" + localBranch + "] removed")
		}
	}
}
