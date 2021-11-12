package repository

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"runtime"
	"time"
)

// outputs only commit id, original date and changed date
func (r *Repository) printOutputShort() error {
	for id, commit := range r.changes {
		fmt.Printf(`commit "%s" from "%s" to "%s"`+"\n",
			id.String(), r.time2String(&commit.when), r.time2String(&commit.new))
	}

	return nil
}

// long output printed, if the avalanche of changes is forward (in time) propagated
// the parent having a date changed (triggering new change) is outputted
//
// when processing git tree, only the necessary info was cached to avoid wasting of memory - rereading git
// information needed only for output (Author and Committer details)
func (r *Repository) printOutputLong() error {

	for _, change := range r.changes {
		// ignoring errors, i was able to read it before
		commitGitObj, _ := r.git.CommitObject(change.id)
		commitGitFrom, _ := r.git.CommitObject(change.from)

		var modifiedDT string

		// if the from was also changed, log the source of change
		if changeFrom, exists := r.changes[change.from]; exists {
			modifiedDT = fmt.Sprintf("%s (changed by commit %s)", r.time2String(&changeFrom.new), changeFrom.from)
		} else {
			var dt time.Time
			if TypeDTAuthor == r.typeDT {
				dt = commitGitFrom.Author.When
			} else {
				dt = commitGitFrom.Committer.When
			}
			modifiedDT = r.time2String(&dt)
		}

		fmt.Printf(
			`---------------------------------------------
commit:       %s
author:       %s <%s>
committer:    %s <%s>
original:     %s
changed:      %s
parent resulting the change:
  commit:     %s  
  current:    %s
  author:     %s <%s>
  committer:  %s <%s>
`,
			commitGitObj.Hash.String(),
			commitGitObj.Author.Name, commitGitObj.Author.Email,
			commitGitObj.Committer.Name, commitGitObj.Committer.Email,
			r.time2String(&change.when),
			r.time2String(&change.new),
			//--
			commitGitFrom.Hash.String(),
			modifiedDT,
			commitGitFrom.Author.Name, commitGitFrom.Author.Email,
			commitGitFrom.Committer.Name, commitGitFrom.Committer.Email)
	}

	return nil
}

// initializes git repository reader and iterator for traversing commits
// no close needed on reader
func (r *Repository) getCommitsIterator(pathRepository string) (*object.CommitIter, error) {
	// open repository
	var err error
	r.git, err = git.PlainOpen(pathRepository)
	if err != nil {
		r.errorLog(err)
		return nil, err
	}

	// get iterator over commits
	var it object.CommitIter
	it, err = r.git.CommitObjects()
	if err != nil {
		r.errorLog(err)
		return nil, err
	}

	return &it, err
}

// takes the git-go commit object and converts it into our commitStruct
// it also registers it into r.commit map to avoid visiting same commit chains multiple times
func (r *Repository) createCommitStruct(current *object.Commit) *commitStruct {
	commit := commitStruct{
		id:       current.Hash,
		children: make(map[plumbing.Hash]*commitStruct, 0),
		parents:  make(map[plumbing.Hash]*commitStruct, 0),
	}

	if r.typeDT == TypeDTAuthor {
		commit.when = current.Author.When
	} else {
		commit.when = current.Committer.When
	}

	r.commits[commit.id] = &commit

	return &commit
}

// outputs error and caller information
func (r *Repository) errorLog(err error) {
	_, file, line, _ := runtime.Caller(1)
	fmt.Printf(`#ERROR (file:"%s" line:"%d") "%s"`, file, line, err.Error())
}

// converts time to string; entry point for setting custom format
func (r *Repository) time2String(time *time.Time) string{
	return time.UTC().Format(timeFormat)
}
