package git

import (
	"path/filepath"
	"sort"
	"strings"
)

// BuildFileTree constructs a tree of all tracked and untracked files
// in the repository, annotated with git status.
func BuildFileTree(repoPath string) (*FileTreeNode, error) {
	// Get all tracked files
	trackedOutput, err := GitExec(repoPath, "ls-files")
	if err != nil {
		return nil, err
	}

	// Get status for changed files
	statusOutput, err := GitExec(repoPath, "status", "--porcelain=v2")
	if err != nil {
		return nil, err
	}

	// Parse status into maps
	stagedMap := make(map[string]FileStatus)
	unstagedMap := make(map[string]FileStatus)
	untrackedSet := make(map[string]bool)

	for _, line := range strings.Split(statusOutput, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "? ") {
			untrackedSet[line[2:]] = true
			continue
		}

		if strings.HasPrefix(line, "1 ") {
			parts := strings.SplitN(line, " ", 9)
			if len(parts) < 9 {
				continue
			}
			xy := parts[1]
			path := parts[8]
			if len(xy) >= 2 {
				if xy[0] != '.' {
					stagedMap[path] = charToStatus(xy[0])
				}
				if xy[1] != '.' {
					unstagedMap[path] = charToStatus(xy[1])
				}
			}
			continue
		}

		if strings.HasPrefix(line, "2 ") {
			parts := strings.SplitN(line, " ", 10)
			if len(parts) < 10 {
				continue
			}
			xy := parts[1]
			pathPart := parts[9]
			paths := strings.SplitN(pathPart, "\t", 2)
			path := paths[0]
			if len(xy) >= 2 {
				if xy[0] != '.' {
					stagedMap[path] = charToStatus(xy[0])
				}
				if xy[1] != '.' {
					unstagedMap[path] = charToStatus(xy[1])
				}
			}
			continue
		}
	}

	// Build set of all files
	allFiles := make(map[string]bool)
	for _, line := range strings.Split(trackedOutput, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			allFiles[line] = true
		}
	}
	for path := range untrackedSet {
		allFiles[path] = true
	}

	// Build tree
	root := &FileTreeNode{
		Name:       ".",
		Path:       ".",
		IsDir:      true,
		IsExpanded: true,
		Depth:      0,
	}

	for path := range allFiles {
		insertPath(root, path, stagedMap, unstagedMap, untrackedSet)
	}

	sortTree(root)
	return root, nil
}

func insertPath(root *FileTreeNode, filePath string, staged, unstaged map[string]FileStatus, untracked map[string]bool) {
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	current := root

	for i, part := range parts {
		isLast := i == len(parts)-1

		if isLast {
			// File node
			node := &FileTreeNode{
				Name:  part,
				Path:  filePath,
				IsDir: false,
				Depth: i + 1,
			}

			if s, ok := staged[filePath]; ok {
				node.Status = &s
				node.IsStaged = true
			}
			if s, ok := unstaged[filePath]; ok {
				node.Status = &s
				node.IsStaged = false
			}
			if untracked[filePath] {
				s := StatusUntracked
				node.Status = &s
			}

			current.Children = append(current.Children, node)
		} else {
			// Directory node — find or create
			dirPath := strings.Join(parts[:i+1], "/")
			found := false
			for _, child := range current.Children {
				if child.IsDir && child.Name == part {
					current = child
					found = true
					break
				}
			}
			if !found {
				dir := &FileTreeNode{
					Name:  part,
					Path:  dirPath,
					IsDir: true,
					Depth: i + 1,
				}
				current.Children = append(current.Children, dir)
				current = dir
			}
		}
	}
}

func sortTree(node *FileTreeNode) {
	if !node.IsDir {
		return
	}

	sort.SliceStable(node.Children, func(i, j int) bool {
		// Dirs first, then files, alphabetical within each group
		if node.Children[i].IsDir != node.Children[j].IsDir {
			return node.Children[i].IsDir
		}
		return node.Children[i].Name < node.Children[j].Name
	})

	for _, child := range node.Children {
		sortTree(child)
	}
}
