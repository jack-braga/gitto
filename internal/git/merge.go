package git

// Merge merges the given branch into the current branch.
func Merge(repoPath, branch string) (string, error) {
	return GitExec(repoPath, "merge", branch)
}

// Rebase rebases the current branch onto the given branch.
func Rebase(repoPath, branch string) (string, error) {
	return GitExec(repoPath, "rebase", branch)
}

// CherryPick applies the given commit to the current branch.
func CherryPick(repoPath, hash string) (string, error) {
	return GitExec(repoPath, "cherry-pick", hash)
}

// RevertCommit reverts the given commit.
func RevertCommit(repoPath, hash string) (string, error) {
	return GitExec(repoPath, "revert", "--no-edit", hash)
}

// AbortMerge aborts an in-progress merge.
func AbortMerge(repoPath string) error {
	_, err := GitExec(repoPath, "merge", "--abort")
	return err
}

// AbortRebase aborts an in-progress rebase.
func AbortRebase(repoPath string) error {
	_, err := GitExec(repoPath, "rebase", "--abort")
	return err
}
