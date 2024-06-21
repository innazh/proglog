package log

import (
	"io"
	"os"

	"github.com/tysonmote/gommap"
)

// These constants define the num of bytes that makeup each index entry
// position in the file would be entWidth * offset
var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

// Our index entries contain two fields: the record's offset and its position in the store file.
type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64 //the size of the index and the starting point of the next index entry
}

// newIndex creates an index for the given file. Once max index size is reached, we memory-map the file and return the index to the caller.
func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{file: f}

	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}

	idx.size = uint64(fi.Size())
	err = os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes))
	if err != nil {
		return nil, err
	}

	idx.mmap, err = gommap.Map(idx.file.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	return idx, nil
}

/*
Close makes sure the memory-maped files has synced its data to the persisted file and that the persisted file has flushed its contents to stable storage.
Then it truncates the persisted file to the amount of data that's actually in it and closes the file.
*/
func (i *index) Close() error {
	if err := i.mmap.Sync(gommap.MS_SYNC); err != nil { //let any remaining changes to persist before unmapping
		return err
	}
	if err := i.mmap.UnsafeUnmap(); err != nil { //unload the file from memory, free up the space
		return err
	}
	if err := i.file.Sync(); err != nil { //commits the current contents of the file to stable storage. In-memory->Disk.
		return err
	}
	if err := i.file.Truncate(int64(i.size)); err != nil { //truncate any empty space, keep only the size od the index
		return err
	}
	return i.file.Close()
}

func (i *index) Read(in int64) (out uint32, pos uint64, err error) {
	if i.size == 0 {
		return 0, 0, io.EOF
	}
	if in == -1 {
		out = uint32((i.size / entWidth) - 1)
	} else {
		out = uint32(in)
	}
	pos = uint64(out) * entWidth
	if i.size < pos+entWidth {
		return 0, 0, io.EOF
	}
	out = enc.Uint32(i.mmap[pos : pos+offWidth])
	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])
	return out, pos, nil
}

func (i *index) Write(off uint32, pos uint64) error {
	if i.isMaxed() {
		return io.EOF
	}
	enc.PutUint32(i.mmap[i.size:i.size+offWidth], off)
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)
	i.size += uint64(entWidth)
	return nil
}

func (i *index) isMaxed() bool {
	return uint64(len(i.mmap)) < i.size+entWidth
}

func (i *index) Name() string {
	return i.file.Name()
}
