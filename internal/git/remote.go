package git

// Push pushes the current branch to its upstream remote.
func Push(repoPath string) (string, error) {
	return GitExecLong(repoPath, "push")
}

// Pull pulls from the upstream remote.
func Pull(repoPath string) (string, error) {
	return GitExecLong(repoPath, "pull")
}

// Fetch fetches from all remotes.
func Fetch(repoPath string) (string, error) {
	return GitExecLong(repoPath, "fetch", "--all")
}
