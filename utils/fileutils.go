/*
 * MIT License
 *
 * Copyright (c) 2020 Alexey Edelev <semlanik@gmail.com>
 *
 * This file is part of gostfix project https://git.semlanik.org/semlanik/gostfix
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this
 * software and associated documentation files (the "Software"), to deal in the Software
 * without restriction, including without limitation the rights to use, copy, modify,
 * merge, publish, distribute, sublicense, and/or sell copies of the Software, and
 * to permit persons to whom the Software is furnished to do so, subject to the following
 * conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies
 * or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
 * INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR
 * PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE
 * FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
 * OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
 * DEALINGS IN THE SOFTWARE.
 */

package utils

import (
	"os"

	unix "golang.org/x/sys/unix"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && !info.IsDir() && info != nil
}

func DirectoryExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && info.IsDir() && info != nil
}

type LockedFile struct {
	file *os.File
	lock *unix.Flock_t
}

func OpenAndLockWait(path string) (file *LockedFile, err error) {
	file = &LockedFile{}
	file.file, err = os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	file.lock = &unix.Flock_t{
		Type: unix.F_WRLCK,
	}
	err = unix.FcntlFlock(file.file.Fd(), unix.F_SETLKW, file.lock)
	file.lock.Type = unix.F_UNLCK

	if err != nil {
		return nil, err
	}

	return
}

func (f *LockedFile) Read(p []byte) (n int, err error) {
	return f.file.Read(p)
}

func (f *LockedFile) Truncate(size int64) error {
	return f.file.Truncate(size)
}

func (f *LockedFile) CloseAndUnlock() error {
	err1 := unix.FcntlFlock(f.file.Fd(), unix.F_SETLKW, f.lock)
	err2 := f.file.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
