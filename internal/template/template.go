package template

import (
	"html/template"
	"io"
)

type HTML = template.HTML

type Template[T any] struct {
	t *template.Template
}

func Must[T any](t *Template[T], err error) *Template[T] {
	return &Template[T]{template.Must(t.t, err)}
}

func ParseFiles[T any](filenames ...string) (*Template[T], error) {
	t, err := template.ParseFiles(filenames...)
	return &Template[T]{t}, err
}

func (t *Template[T]) Execute(w io.Writer, data T) error {
	return t.t.Execute(w, data)
}
