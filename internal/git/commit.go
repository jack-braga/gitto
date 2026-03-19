package git

// Commit creates a new commit with the given message.
func Commit(repoPath, message string) error {
	_, err := GitExec(repoPath, "commit", "-m", message)
	return err
}

// CommitAmend amends the last commit with a new message.
func CommitAmend(repoPath, message string) error {
	_, err := GitExec(repoPath, "commit", "--amend", "-m", message)
	return err
}
