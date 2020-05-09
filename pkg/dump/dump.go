package dump

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type Writer interface {
	Write(...Doc) (int, error)
}

type Closer interface {
	Close() error
}

type WriteCloser interface {
	Writer
	Closer
}

type Reader interface {
	Read([]Doc) (int, error)
}

type ReadCloser interface {
	Reader
	Closer
}

type Sizer interface {
	Size() int64
}

type Doc struct {
	ID    string      `json:"_id"`
	Index string      `json:"index"`
	Body  interface{} `json:"body"`
}

type fileWriter struct {
	file io.Closer
	WriteCloser
}

func NewFileWriter(path string) (*fileWriter, error) {
	dir, _ := filepath.Split(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return &fileWriter{
		file:        f,
		WriteCloser: NewWriter(f),
	}, nil
}

func (w *fileWriter) Close() error {
	if err := w.WriteCloser.Close(); err != nil {
		return err
	}
	return w.file.Close()
}

type writer struct {
	w io.Writer

	mu     sync.Mutex
	opened bool
	closed bool
}

func NewWriter(w io.Writer) *writer {
	return &writer{w: w}
}

func (w *writer) Write(docs ...Doc) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	var buf bytes.Buffer
	leadComma := w.opened
	if err := w.open(); err != nil {
		return n, err
	}
	if leadComma {
		buf.Write([]byte{',', '\n'})
	}
	size := len(docs)
	for i, doc := range docs {
		buf.WriteByte('\t')
		data, err := json.MarshalIndent(doc, "\t", "\t")
		if err != nil {
			return n, err
		}
		buf.Write(data)
		if i < size-1 {
			buf.Write([]byte{',', '\n'})
		}
		if _, err := buf.WriteTo(w.w); err != nil {
			return n, err
		}
		n++
		buf.Reset()
	}
	return n, nil
}

func (w *writer) open() error {
	if w.opened {
		return nil
	}
	if _, err := w.w.Write([]byte{'[', '\n'}); err != nil {
		return err
	}
	w.opened = true
	return nil
}

func (w *writer) Close() error { return w.close() }
func (w *writer) close() error {
	if w.closed {
		return nil
	}
	if _, err := w.w.Write([]byte{'\n', ']'}); err != nil {
		return err
	}
	w.closed = true
	return nil
}

type fileReader struct {
	Closer
	Reader
	size int64
}

func NewFileReader(name string) (*fileReader, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	s, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return &fileReader{
		Closer: f,
		Reader: NewReader(f),
		size:   s.Size(),
	}, nil
}

func (r *fileReader) Size() int64 { return r.size }

type readCloser struct {
	Reader
	Closer
}

func NewReadCloser(rc io.ReadCloser) *readCloser {
	return &readCloser{
		Reader: NewReader(rc),
		Closer: rc,
	}
}

type reader struct {
	d *json.Decoder

	mu     sync.Mutex
	opened bool
	closed bool
}

func NewReader(r io.Reader) *reader {
	return &reader{d: json.NewDecoder(r)}
}

func (r *reader) Read(docs []Doc) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.open(); err != nil {
		return n, err
	}
	for i := range docs {
		if !r.d.More() {
			if err := r.close(); err != nil {
				return n, err
			}
			return n, io.EOF
		}
		if err := r.d.Decode(&docs[i]); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func (r *reader) open() error {
	if r.opened {
		return nil
	}
	if _, err := r.d.Token(); err != nil {
		return err
	}
	r.opened = true
	return nil
}

func (r *reader) close() error {
	if r.closed {
		return nil
	}
	if _, err := r.d.Token(); err != nil {
		return err
	}
	r.closed = true
	return nil
}
