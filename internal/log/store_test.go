package log

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testRecord   = []byte("hello world")
	testRecWidth = uint64(len(testRecord)) + recordLenBytes
)

func TestStoreAppendRead(t *testing.T) {
	f, err := os.CreateTemp(os.TempDir(), "index_test")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	s, err := newStore(f)
	require.NoError(t, err)

	testAppend(t, s)
	testRead(t, s)
	testReadAt(t, s)

	s, err = newStore(f)
	require.NoError(t, err)
	testRead(t, s)
}

func testAppend(t *testing.T, s *store) {
	t.Helper()
	for i := uint64(1); i < 4; i++ {
		n, pos, err := s.Append(testRecord)
		require.NoError(t, err)
		require.Equal(t, pos+n, testRecWidth*i)
	}
}

func testRead(t *testing.T, s *store) {
	t.Helper()

	var pos uint64
	for i := uint64(1); i < 4; i++ {
		read, err := s.Read(pos)
		require.NoError(t, err)
		require.Equal(t, testRecord, read)
		pos += testRecWidth
	}
}

func testReadAt(t *testing.T, s *store) {
	t.Helper()

	for i, off := uint64(1), int64(0); i < 4; i++ {
		b := make([]byte, recordLenBytes)
		n, err := s.ReadAt(b, off)
		require.NoError(t, err)
		require.Equal(t, recordLenBytes, n)

		off += int64(n)
		size := enc.Uint64(b)

		b = make([]byte, size)
		n, err = s.ReadAt(b, off)
		require.NoError(t, err)
		require.Equal(t, testRecord, b)
		require.Equal(t, int(size), n)
		off += int64(n)
	}
}

func TestStoreClose(t *testing.T) {
	f, err := os.CreateTemp(os.TempDir(), "store_close_test")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	s, err := newStore(f)
	require.NoError(t, err)

	_, _, err = s.Append(testRecord)
	require.NoError(t, err)

	f, beforeSize, err := openFile(f.Name())
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	_, afterSize, err := openFile(f.Name())
	require.NoError(t, err)
	require.True(t, afterSize > beforeSize)
}

func openFile(name string) (file *os.File, size int64, err error) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, 0, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}
	return f, fi.Size(), nil
}
