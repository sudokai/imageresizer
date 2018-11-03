package store

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestTwoTier_Get(t *testing.T) {
	testFilename := "/300x300/crop/s/natasha-kasim-708827-unsplash.jpg"
	tmpdir, err := ioutil.TempDir("../testdata", "TestTwoTier_Get")
	if err != nil {
		t.Errorf("Error creating temp dir")
		return
	}
	defer os.RemoveAll(tmpdir)
	inbuf, err := ioutil.ReadFile("../testdata" + testFilename)
	if inbuf == nil || err != nil {
		t.Errorf("Could not read test file")
	}
	fs := NewFileStore(tmpdir)
	fs.Put(testFilename, inbuf)
	twotier := &TwoTier{
		Store: fs,
		Cache: nil,
	}

	outbuf, err := twotier.Get(testFilename)
	if bytes.Compare(inbuf, outbuf) != 0 {
		t.Errorf("Input and output buffers differ")
	}
}
