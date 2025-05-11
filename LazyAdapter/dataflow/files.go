package dataflow

import (
	"io/ioutil"
)

type FileContent struct {
	Path    string
	Content []byte
}

type OpenFilesFlow struct {
	source  DataFlow[string]
	current FileContent
}

func OpenFiles() func(DataFlow[string]) DataFlow[FileContent] {
	return func(source DataFlow[string]) DataFlow[FileContent] {
		return &OpenFilesFlow{
			source: source,
		}
	}
}

func (f *OpenFilesFlow) Next() bool {
	if !f.source.Next() {
		return false
	}

	path := f.source.Value()
	content, err := ioutil.ReadFile(path)
	if err != nil {

		return f.Next()
	}

	f.current = FileContent{
		Path:    path,
		Content: content,
	}
	return true
}

func (f *OpenFilesFlow) Value() FileContent {
	return f.current
}

func (f *OpenFilesFlow) Reset() {
	f.source.Reset()
}
