// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs

import (
	"bytes"
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

type androidNode struct {
	mtpNodeImpl

	// If set, the backing file was changed.
	write     bool
	start     time.Time
	byteCount int64
}

func (n *androidNode) startEdit() bool {
	if n.write {
		return true
	}

	n.start = time.Now()
	n.byteCount = 0
	if err := n.fs.dev.AndroidBeginEditObject(n.Handle()); err != nil {
		log.Println("AndroidBeginEditObject failed:", err)
		return false
	}
	n.write = true
	return true
}

func (n *androidNode) endEdit() bool {
	if !n.write {
		return true
	}

	dt := time.Now().Sub(n.start)
	log.Printf("%d bytes in %v: %d mb/s",
		n.byteCount, dt, (1e3*n.byteCount)/(dt.Nanoseconds()))

	if err := n.fs.dev.AndroidEndEditObject(n.Handle()); err != nil {
		log.Println("AndroidEndEditObject failed:", err)
		return false
	}
	n.write = false
	return true
}

func (n *androidNode) Open(flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	return &androidFile{
		node: n,
		File: nodefs.NewDefaultFile(),
	}, fuse.OK
}

func (n *androidNode) Truncate(file nodefs.File, size uint64, context *fuse.Context) (code fuse.Status) {
	w := n.write
	if !n.startEdit() {
		return fuse.EIO
	}
	if err := n.fs.dev.AndroidTruncate(n.Handle(), int64(size)); err != nil {
		log.Println("AndroidTruncate failed:", err)
		return fuse.EIO
	}
	n.Size = int64(size)

	if !w {
		if !n.endEdit() {
			return fuse.EIO
		}
	}
	return fuse.OK
}

var _ = mtpNode((*androidNode)(nil))

type androidFile struct {
	nodefs.File
	node *androidNode
}

func (f *androidFile) Read(dest []byte, off int64) (fuse.ReadResult, fuse.Status) {
	if off > f.node.Size {
		// ENXIO = no such address.
		return nil, fuse.Status(int(syscall.ENXIO))
	}

	if off+int64(len(dest)) > f.node.Size {
		dest = dest[:f.node.Size-off]
	}
	b := bytes.NewBuffer(dest[:0])
	err := f.node.fs.dev.AndroidGetPartialObject64(f.node.Handle(), b, off, uint32(len(dest)))
	if err != nil {
		log.Println("AndroidGetPartialObject64 failed:", err)
		return nil, fuse.EIO
	}

	return fuse.ReadResultData(dest[:b.Len()]), fuse.OK
}

func (f *androidFile) String() string {
	return fmt.Sprintf("androidFile h=0x%x", f.node.Handle())
}

func (f *androidFile) Write(dest []byte, off int64) (written uint32, status fuse.Status) {
	if !f.node.startEdit() {
		return 0, fuse.EIO
	}
	f.node.byteCount += int64(len(dest))
	b := bytes.NewBuffer(dest)
	err := f.node.fs.dev.AndroidSendPartialObject(f.node.Handle(), off, uint32(len(dest)), b)
	if err != nil {
		log.Println("AndroidSendPartialObject failed:", err)
		return 0, fuse.EIO
	}
	written = uint32(len(dest) - b.Len())
	if off+int64(written) > f.node.Size {
		f.node.Size = off + int64(written)
	}
	return written, fuse.OK
}

func (f *androidFile) Flush() fuse.Status {
	if !f.node.endEdit() {
		return fuse.EIO
	}
	return fuse.OK
}
