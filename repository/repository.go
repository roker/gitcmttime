package repository

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"os"
	"strings"
	"time"
)

const timeFormat = time.RFC3339
const TypeDTAuthor = TypeDateTime(0)
const TypeDTCommitter = TypeDateTime(1)
const TypeOutShort = TypeOutput(0)
const TypeOutLong = TypeOutput(1)
const TypeOutErrors = TypeOutput(1)

type TypeDateTime int8
type TypeOutput int8

var epoch = time.Unix(0, 0)

type Repository struct {
	git     *git.Repository
	commits map[plumbing.Hash]*commitStruct
	changes map[plumbing.Hash]*changeStruct
	typeDT  TypeDateTime
}

// opens repository and starts processing it, typedt defines which time to take, committer(TypeDTAuthor)
// or author(TypeDTCommitter)
func OpenRepository(pathRepository string, typedt TypeDateTime) (*Repository, error) {
	r := Repository{
		commits: make(map[plumbing.Hash]*commitStruct, 0),
		changes: make(map[plumbing.Hash]*changeStruct, 0),
	}

	var err error
	var it *object.CommitIter

	// parameters cleanup
	pathRepository = strings.TrimSpace(pathRepository)
	pathRepository = strings.TrimRight(pathRepository, string(os.PathSeparator)) + string(os.PathSeparator)

	if typedt != TypeDTAuthor && typedt != TypeDTCommitter {
		err := errors.New("unknown datetime type, use typeDTAuthor or typeDTCommitter")
		r.errorLog(err)
		return nil, err
	}
	r.typeDT = typedt

	// get iterator to git commits
	if it, err = r.getCommitsIterator(pathRepository + `.git` + string(os.PathSeparator)); err != nil {
		r.errorLog(err)
		return nil, err
	}

	// create structure with all chains connected and times adjusted by parent time
	(*it).ForEach(func(commitGitObject *object.Commit) error {
		r.addCommit(nil, commitGitObject)
		return nil
	})

	return &r, nil
}

func (r *Repository) PrintOutput(typeout TypeOutput) error {
	switch typeout {
	case TypeOutShort:
		return r.printOutputShort()
	case TypeOutLong:
		return r.printOutputLong()
	}

	err := errors.New("unknown datetime type, use TypeOutShort or TypeOutLong")
	r.errorLog(err)
	return err
}

// recursive function that is adding commit structures and creating a chain
// - if the commit id has already been visited it registers caller as child and returns
// - if the commit id is unknown it creates new commit structure and recursively visits all parents
// 	 - parent returns back its time and if it is newer that current commit time it is updated to
//     returned time + 1 second
// - the current commit time is returned
func (r *Repository) addCommit(child *commitStruct, currentGitObject *object.Commit) (*time.Time, error) {
	if current, exists := r.commits[currentGitObject.Hash]; exists {
		// this commit was already visited, but the child might be unknown
		current.addChild(child)
		return &current.when, nil
	}

	// create and register commit to registry
	currentCommit := r.createCommitStruct(currentGitObject)

	// register caller as a child
	currentCommit.addChild(child)

	// recurse all parents, adjust current time if needed
	var newestParentDT *time.Time = &epoch
	var newestParentId plumbing.Hash
	for _, parentid := range currentGitObject.ParentHashes {

		// obtain parent git object
		var err error
		var parentGitObject *object.Commit
		if parentGitObject, err = r.git.CommitObject(parentid); err != nil {
			r.errorLog(err)
			return &epoch, err
		}

		// recurse into parent
		var parentDT *time.Time
		if parentDT, err = r.addCommit(currentCommit, parentGitObject); err != nil {
			r.errorLog(err)
			continue
		}

		// tracking newest datetime and its commit id
		if parentDT.After(*newestParentDT) {
			newestParentDT = parentDT
			newestParentId = parentid
		}
	}

	// (#1) verify if current commit time is older than the newest parent datetime and store the change
	if currentCommit.when.Before(*newestParentDT) {
		change := r.createChange(currentCommit, newestParentDT, &newestParentId)
		return &change.new, nil
	}

	// return current date to caller to adjust its time if needed (check #1)
	return &currentCommit.when, nil
}

func (r *Repository) createChange(cs *commitStruct, new *time.Time, source *plumbing.Hash) *changeStruct {
	change := changeStruct{
		commitStruct: cs,
		new:          new.Add(time.Second),
		from:         *source,
	}

	r.changes[cs.id] = &change

	return &change
}
