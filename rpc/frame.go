package rpc

import (
	"encoding/binary"
	"errors"
	"io"
)

const frameBufSize = 4096

// FrameReader is a data frame reader.
type FrameReader struct {
	r io.Reader

	remaining int
}

// NewFrameReader returns a FrameReader with the given underlying reader.
func NewFrameReader(r io.Reader) *FrameReader {
	return &FrameReader{
		r: r,
	}
}

// Read reads the data from frames into p.
func (r *FrameReader) Read(p []byte) (int, error) {
	var read int

	for read < len(p) {
		if r.remaining == 0 {
			var s [4]byte
			n, err := r.r.Read(s[:])
			if err != nil {
				return read, err
			}
			if n < 4 {
				return read, errors.New("rpc: invalid frame header")
			}

			size := binary.BigEndian.Uint32(s[:])
			if size == 0 {
				return read, io.EOF
			}
			r.remaining = int(size)
		}

		l := len(p)
		if read+r.remaining < l {
			l = read + r.remaining
		}
		n, err := r.r.Read(p[read:l])
		if err != nil {
			return read + n, err
		}

		read += n
		r.remaining -= n
	}

	return read, nil
}

var emptyFrame = [4]byte{0x00, 0x00, 0x00, 0x00}

// FrameWriter is a data frame writer.
type FrameWriter struct {
	w io.Writer

	buf []byte
	l   int
	s   int
}

// NewFrameWriter returns a FrameWriter with the given underlying writer.
func NewFrameWriter(w io.Writer) *FrameWriter {
	return &FrameWriter{
		w:   w,
		buf: make([]byte, frameBufSize),
		l:   0,
		s:   frameBufSize,
	}
}

// Write writes p in a frame.
func (w *FrameWriter) Write(p []byte) (int, error) {
	l := len(p)
	if l > w.s-w.l {
		// The data is bigger then the remaining frame,
		// flush what we have and start the frame fresh
		if err := w.flush(); err != nil {
			return 0, err
		}
	}

	if l > w.s {
		// The data is bigger than a frame,
		// write the data into its own frame
		var s [4]byte
		binary.BigEndian.PutUint32(s[:], uint32(len(p)))
		if _, err := w.w.Write(s[:]); err != nil {
			return 0, err
		}
		return w.w.Write(p)
	}

	copy(w.buf[w.l:], p)
	w.l += l
	return l, nil
}

func (w *FrameWriter) flush() error {
	if w.l == 0 {
		return nil
	}

	var s [4]byte
	binary.BigEndian.PutUint32(s[:], uint32(w.l))

	_, err := w.w.Write(s[:])
	if err == nil {
		_, err = w.w.Write(w.buf[:w.l])
		w.l = 0
	}
	return err
}

// Flush flushes the current frame and writes an empty frame.
func (w *FrameWriter) Flush() error {
	err := w.flush()
	if err == nil {
		_, err = w.w.Write(emptyFrame[:])
	}
	return err
}
