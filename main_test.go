package main

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetenv(t *testing.T) {
	var e string
	// ---------
	e = Getenv("YOLO", "nope")
	assert.Equal(t, e, "nope")
	// ----------
	os.Setenv("YOLO", "woohoo")
	e = Getenv("YOLO", "nope")
	assert.Equal(t, e, "woohoo")
}

func TestGoroutines(t *testing.T) {
	var r interface{}
	r = Goroutines()

	switch r := r.(type) {
	case int:
		t.Log(int(r))
	default:
		t.Fatal("wrong type")
	}
}

func TestUptime(t *testing.T) {
	var u interface{}
	u = Uptime()
	switch u := u.(type) {
	case int64:
		t.Log(int64(u))
	default:
		t.Fatal("wrong type")
	}
}

type FakeConfigReader struct {
	text string
	err  error
}

func (r FakeConfigReader) Read() (string, error) {
	return r.text, r.err
}

func TestReadSecretFromFile(t *testing.T) {
	cfgReader := ConfigFileReader{
		path: "/dev/null",
	}

	str, err := ReadSecretFromFile(cfgReader)

	assert.Equal(t, nil, err)
	assert.Equal(t, "", str)

	var cfgReaderTests = []struct {
		name   string
		reader FakeConfigReader
		out    string
		err    error
	}{
		{
			name: "yolo_nil",
			reader: FakeConfigReader{
				text: "yolo",
				err:  nil,
			},
			out: "yolo",
			err: nil,
		},
		{
			name: "yolo_err_EOF",
			reader: FakeConfigReader{
				text: "",
				err:  io.EOF,
			},
			out: "",
			err: io.EOF,
		},
		{
			name: "yolo_err_ErrUnexpectedEOF",
			reader: FakeConfigReader{
				text: "aabbcc",
				err:  io.ErrUnexpectedEOF,
			},
			out: "aabbcc",
			err: io.ErrUnexpectedEOF,
		},
		{
			name: "yolo_err_trailingNewLine",
			reader: FakeConfigReader{
				text: `hello
`,
				err: nil,
			},
			out: "hello",
			err: nil,
		},
		{
			name: "yolo_err_trailingNewLineMultiLine",
			reader: FakeConfigReader{
				text: `hello
  
						
`,
				err: nil,
			},
			out: "hello\n  \n\t\t\t\t\t\t", // only remove the final new line
			err: nil,
		},
	}

	for _, tt := range cfgReaderTests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := ReadSecretFromFile(tt.reader)
			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.out, value)
		})
	}

}
