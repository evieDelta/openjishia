/*Package helputil .

  A utility for putting help pages in md files and importing them via go embed

  to use add a directory called `topic` and go embed it via
  ```
  //go:embed topics/*
  var helpTopics embed.FS

  func init() {
  	htopic.MustAddFS(helpTopics)
  }
  ```

  this does not include an implementation of how to get or identify the right file
  implementations of a help command will need to figure that out themselves
*/
package helputil

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
)

var list = make(map[string]string)

// Get a help topic, returns true if a page exists
func Get(page string) (string, bool) {
	x, ok := list[strings.ToLower(page)]
	return x, ok
}

// List gets a list of topics
func List(prefix, suffix, contains string) []string {
	keys := make([]string, 0, len(list))
	for k := range list {
		lk := strings.ToLower(k)
		if !strings.HasPrefix(lk, prefix) {
			continue
		}
		if !strings.HasSuffix(lk, suffix) {
			continue
		}
		if !strings.Contains(lk, contains) {
			continue
		}
		keys = append(keys, k)
	}

	return keys
}

// Add a help topic, will not overwrite an existing help topic
func Add(page, text string) bool {
	if list[page] != "" {
		return false
	}
	list[page] = text
	return true
}

// MustAdd is Add but it panics if it fails
func MustAdd(page, text string) {
	ok := Add(page, text)
	if !ok {
		panic("Could not add help page " + page + " because a page already exists with that name")
	}
}

// AddFS adds a list of help topics from a fs.FS
func AddFS(f fs.FS) error {
	const prefix = "topics"

	l, err := fs.ReadDir(f, prefix)
	if err != nil {
		return err
	}
	for _, de := range l {
		if de.IsDir() {
			return fmt.Errorf("topics directory cannot contain directories, problem: %v is dir", de.Name())
		}
		if !strings.HasSuffix(de.Name(), ".md") {
			return fmt.Errorf("topics directory can only contain .md files, problem: %v is not .md", de.Name())
		}
		dat, err := fs.ReadFile(f, path.Join(prefix, de.Name()))
		if err != nil {
			return err
		}
		list[de.Name()[:len(de.Name())-3]] = string(dat)
	}

	return nil
}

// MustAddFS is AddFS but it panics if an error occours
func MustAddFS(f fs.FS) {
	err := AddFS(f)
	if err != nil {
		panic(err)
	}
}
