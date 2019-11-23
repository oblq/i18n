package sprbox

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// gitRepository represent a git repository
type gitRepository struct {
	path                           string
	branchName, commit, build, tag string

	Error error
	mutex sync.Mutex
}

// NewGitRepository return a new gitRepository instance for the given path
func newGitRepository(path string) *gitRepository {
	repo := &gitRepository{path: path}
	repo.updateInfo()
	return repo
}

// Info return Git repository info.
func (g *gitRepository) Info() string {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	gitLog := kvLogger{ValuePainter: magenta}
	return fmt.Sprintf("%s\n%s\n%s\n%s\n",
		gitLog.sprint("Git Branch:", g.branchName),
		gitLog.sprint("Git Commit:", g.commit),
		gitLog.sprint("Git Tag:", g.tag),
		gitLog.sprint("Git Build:", g.build))
}

// updateInfo grab git info and set 'Error' var eventually.
func (g *gitRepository) updateInfo() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.branchName = g.git("rev-parse", "--abbrev-ref", "HEAD")
	g.commit = g.git("rev-parse", "--short", "HEAD")
	g.build = g.git("rev-list", "--all", "--count")
	g.tag = g.git("describe", "--abbrev=0", "--tags", "--always")
}

// Git is the bash git command.
func (g *gitRepository) git(params ...string) string {
	cmd := exec.Command("git", params...)
	if len(g.path) > 0 {
		cmd.Dir = g.path
	}

	output, err := cmd.Output()
	if err != nil {
		gitErrString := err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			gitErrString = string(exitError.Stderr)
		}
		gitErrString = strings.TrimPrefix(gitErrString, "fatal: ")
		gitErrString = strings.TrimSuffix(gitErrString, "\n")
		gitErrString = strings.TrimSuffix(gitErrString, ": .git")
		g.Error = errors.New(gitErrString)
		return gitErrString
	}

	out := strings.TrimSuffix(string(output), "\n")
	return out
}
