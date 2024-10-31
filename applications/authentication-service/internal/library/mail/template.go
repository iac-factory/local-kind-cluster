package mail

import (
	"bytes"
	"embed"
	_ "embed"
	"errors"
	html "html/template"
	"io"
	text "text/template"
)

type Metadata struct {
	Expiration int    // e.g. 5
	Duration   string // e.g. "days"
	URL        string // verification url link
}

type Implementation interface {
	*html.Template | *text.Template
}

type Template[T Implementation] struct {
	t      string
	name   string
	buffer bytes.Buffer

	functions map[string]interface{}
	template  T
}

func (t Template[T]) Template() T {
	var value interface{} = t.template

	return value.(T)
}

func (t Template[T]) Execute(w io.Writer, metadata Metadata) error {
	var value interface{} = t.template

	switch t.t {
	case "html":
		return value.(*html.Template).Execute(w, metadata)
	case "text":
		return value.(*text.Template).Execute(w, metadata)
	default:
		return errors.New("invalid template type and value")
	}
}

var (
	//go:embed *.go.template
	directive embed.FS

	Text = Template[*text.Template]{
		t:         "text",
		name:      "email.text.go.template",
		buffer:    bytes.Buffer{},
		functions: text.FuncMap{},
		template:  &text.Template{},
	}

	HTML = Template[*html.Template]{
		t:         "html",
		name:      "email.html.go.template",
		buffer:    bytes.Buffer{},
		functions: text.FuncMap{},
		template:  &html.Template{},
	}
)

func init() {
	{
		buffer, e := directive.ReadFile(Text.name)
		if e != nil {
			panic(e)
		}

		if _, e := Text.buffer.Write(buffer); e != nil {
			panic(e)
		}

		Text.template = text.Must(text.New(Text.name).Option("missingkey=error").Parse(Text.buffer.String()))
	}

	{
		buffer, e := directive.ReadFile(HTML.name)
		if e != nil {
			panic(e)
		}

		if _, e := HTML.buffer.Write(buffer); e != nil {
			panic(e)
		}

		HTML.template = html.Must(html.New(HTML.name).Option("missingkey=error").Parse(HTML.buffer.String()))
	}
}
