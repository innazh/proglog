package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	enc = binary.BigEndian
)

const (
	recordLenBytes = 8 //8 of bytes for uint64
)

type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	size := uint64(fi.Size())
	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f),
	}, nil
}

// Append persists given bytes p to the store. Returns the num of bytes written, record's starting position and error.
func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pos = s.size
	//we're using buffer instead of writing to the file directly for performance reasons
	//this writes the length of data to the buffer:
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}
	//here we actually write p to the buffer
	bytesWritten, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}

	bytesWritten += recordLenBytes
	s.size += uint64(bytesWritten)
	return uint64(bytesWritten), pos, nil
}

// Read returns a sequence of bytes from the store, given its position
func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.buf.Flush(); err != nil {
		return nil, err
	}
	//Since we first store the length of record aka the number of bytes to read, we first need to retrieve how many bytes we need to read to get the record in requested.
	//get the size of the record:
	size := make([]byte, recordLenBytes)
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}
	//now read the record, which only starts after the position + the number of bytes to read (which is 8 bytes==recordLenBytes)
	recordBytes := make([]byte, enc.Uint64(size))
	if _, err := s.File.ReadAt(recordBytes, int64(pos+recordLenBytes)); err != nil {
		return nil, err
	}
	return recordBytes, nil
}

// ReadAt reads len(p) bytes into p, beginning at the off offset in the store's file. Returns the num of bytes read and err
func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.buf.Flush(); err != nil {
		return 0, err
	}

	return s.File.ReadAt(p, off)
}

// Close persists any buffered data beforee closing the file
func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.buf.Flush(); err != nil {
		return err
	}
	return s.File.Close()
}
