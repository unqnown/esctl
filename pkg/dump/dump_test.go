package dump

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func Test_Writer(t *testing.T) {
	f, err := os.Create("test.json")
	if err != nil {
		t.Fatal(err)
	}
	w := NewWriter(f)

	if _, err := w.Write([]Doc{
		{
			ID:    "1",
			Index: "items",
		},
		{
			ID:    "2",
			Index: "items",
		},
	}...); err != nil {
		t.Fatal(err)
	}

	if _, err := w.Write([]Doc{
		{
			ID:    "3",
			Index: "items",
		},
		{
			ID:    "4",
			Index: "items",
		},
	}...); err != nil {
		t.Fatal(err)
	}

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

func Test_Reader(t *testing.T) {
	f, err := os.Open("test.json")
	if err != nil {
		t.Fatal(err)
	}

	r := NewReader(f)

	docs := make([]Doc, 1)

loop:
	for {
		switch _, err := r.Read(docs); err {
		case nil:
			fmt.Printf("%+v", docs)
		case io.EOF:
			fmt.Printf("received EOF")
			break loop
		default:
			t.Fatal(err)
		}
	}
}
