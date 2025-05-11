package dataflow

import (
	"os"
	"path/filepath"
)

type DirFlow struct {
	rootDir    string
	recursive  bool
	files      []string
	currentIdx int
}

func Dir(path string, recursive bool) *Pipeline[string] {
	return New(&DirFlow{
		rootDir:    path,
		recursive:  recursive,
		currentIdx: -1,
	})
}

func (d *DirFlow) Next() bool {
	if d.files == nil {
		d.loadFiles()
	}

	d.currentIdx++
	return d.currentIdx < len(d.files)
}

func (d *DirFlow) Value() string {
	if d.currentIdx < 0 || d.currentIdx >= len(d.files) {
		return ""
	}
	return d.files[d.currentIdx]
}

func (d *DirFlow) Reset() {
	d.currentIdx = -1
}

func (d *DirFlow) loadFiles() {
	d.files = []string{}

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			d.files = append(d.files, path)
		} else if path != d.rootDir && !d.recursive {
			return filepath.SkipDir
		}

		return nil
	}

	filepath.Walk(d.rootDir, walkFunc)
}
