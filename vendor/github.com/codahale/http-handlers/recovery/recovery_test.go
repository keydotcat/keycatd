package recovery

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRecoveryHandler(t *testing.T) {
	tmp := Swap(os.Stderr)
	defer tmp.Restore(os.Stderr)

	recovery := Wrap(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("aaaaugh")
		}),
		LogOnPanic,
	)

	server := httptest.NewServer(recovery)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	actual := string(b)
	if !strings.HasPrefix(actual, "Internal Server Error") {
		t.Errorf("Unexpected response: %#v", actual)
	}

	b, err = tmp.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	actual = string(b)
	t.Log("\n" + actual)
	if !strings.Contains(actual, "aaaaugh") {
		t.Errorf("Unexpected error output: %#v", actual)
	}
}

// extract this out into a reusable lib

type TmpFile struct {
	original os.File
	redir    *os.File
}

func Swap(f *os.File) *TmpFile {
	redir, err := ioutil.TempFile(os.TempDir(), "tmpfile")
	if err != nil {
		panic(err)
	}
	tmpFile := TmpFile{
		original: *f,
		redir:    redir,
	}
	*f = *redir
	return &tmpFile
}

func (tmp TmpFile) ReadAll() ([]byte, error) {
	if _, err := tmp.redir.Seek(0, 0); err != nil {
		return nil, err
	}
	return ioutil.ReadAll(tmp.redir)

}

func (tmp TmpFile) Restore(f *os.File) {
	*f = tmp.original
}
