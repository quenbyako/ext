// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

// Package fs defines basic interfaces to a file system.
// A file system can be provided by the host operating system
// but also by other packages.
package fs

import (
	stdfs "io/fs"
)

// Generic file system errors.
// Errors returned by file systems can be tested against these errors
// using [errors.Is].
var (
	ErrInvalid    = stdfs.ErrInvalid    // "invalid argument"
	ErrPermission = stdfs.ErrPermission // "permission denied"
	ErrExist      = stdfs.ErrExist      // "file already exists"
	ErrNotExist   = stdfs.ErrNotExist   // "file does not exist"
	ErrClosed     = stdfs.ErrClosed     // "file already closed"
)

// SkipDir is used as a return value from WalkDirFuncs to indicate that
// the directory named in the call is to be skipped. It is not returned
// as an error by any function.
var SkipDir = stdfs.SkipDir

// PathError records an error and the operation and file path that caused it.
type PathError = stdfs.PathError

// Glob returns the names of all files matching pattern or nil
// if there is no matching file. The syntax of patterns is the same
// as in [path.Match]. The pattern may describe hierarchical names such as
// usr/*/bin/ed.
//
// Glob ignores file system errors such as I/O errors reading directories.
// The only possible returned error is [path.ErrBadPattern], reporting that
// the pattern is malformed.
//
// If fs implements [GlobFS], Glob calls fs.Glob.
// Otherwise, Glob uses [ReadDir] to traverse the directory tree
// and look for matches for the pattern.
func Glob(fsys FS, pattern string) (matches []string, err error) { return stdfs.Glob(fsys, pattern) }

// ReadFile reads the named file from the file system fs and returns its contents.
// A successful call returns a nil error, not [io.EOF].
// (Because ReadFile reads the whole file, the expected EOF
// from the final Read is not treated as an error to be reported.)
//
// If fs implements [ReadFileFS], ReadFile calls fs.ReadFile.
// Otherwise ReadFile calls fs.Open and uses Read and Close
// on the returned [File].
func ReadFile(fsys FS, name string) ([]byte, error) { return stdfs.ReadFile(fsys, name) }

// ValidPath reports whether the given path name is valid for use in a call to
// Open.
//
// Path names passed to open are UTF-8-encoded, unrooted, slash-separated
// sequences of path elements, like “x/y/z”. Path names must not contain an
// element that is “.” or “..” or the empty string, except for the special case
// that the root directory is named “.”. Paths must not start or end with a
// slash: “/x” and “x/” are invalid.
//
// Note that paths are slash-separated on all systems, even Windows. Paths
// containing other characters such as backslash and colon are accepted as
// valid, but those characters must never be interpreted by an FS implementation
// as path element separators.
func ValidPath(name string) bool { return stdfs.ValidPath(name) }

// WalkDir walks the file tree rooted at root, calling fn for each file or
// directory in the tree, including root.
//
// All errors that arise visiting files and directories are filtered by fn: see
// the fs.WalkDirFunc documentation for details.
//
// The files are walked in lexical order, which makes the output deterministic
// but requires WalkDir to read an entire directory into memory before
// proceeding to walk that directory.
//
// WalkDir does not follow symbolic links found in directories, but if root
// itself is a symbolic link, its target will be walked.
func WalkDir(fsys FS, root string, fn WalkDirFunc) error { return stdfs.WalkDir(fsys, root, fn) }

// A DirEntry is an entry read from a directory
// (using the [ReadDir] function or a [ReadDirFile]'s ReadDir method).
type DirEntry = stdfs.DirEntry

// FileInfoToDirEntry returns a [DirEntry] that returns information from info.
// If info is nil, FileInfoToDirEntry returns nil.
func FileInfoToDirEntry(info FileInfo) DirEntry { return stdfs.FileInfoToDirEntry(info) }

// ReadDir reads the named directory
// and returns a list of directory entries sorted by filename.
//
// If fs implements [ReadDirFS], ReadDir calls fs.ReadDir.
// Otherwise ReadDir calls fs.Open and uses ReadDir and Close
// on the returned file.
func ReadDir(fsys FS, name string) ([]DirEntry, error) { return stdfs.ReadDir(fsys, name) }

// An FS provides access to a hierarchical file system.
//
// The FS interface is the minimum implementation required of the file system.
// A file system may implement additional interfaces, such as ReadFileFS, to
// provide additional or optimized functionality.
type FS = stdfs.FS

// Sub returns an FS corresponding to the subtree rooted at fsys's dir.
//
// If dir is ".", Sub returns fsys unchanged. Otherwise, if fs implements SubFS,
// Sub returns fsys.Sub(dir). Otherwise, Sub returns a new FS implementation sub
// that, in effect, implements sub.Open(name) as
// fsys.Open(path.Join(dir, name)). The implementation also translates calls to
// ReadDir, ReadFile, and Glob appropriately.
//
// Note that Sub(os.DirFS("/"), "prefix") is equivalent to os.DirFS("/prefix")
// and that neither of them guarantees to avoid operating system accesses
// outside "/prefix", because the implementation of os.DirFS does not check for
// symbolic links inside "/prefix" that point to other directories. That is,
// os.DirFS is not a general substitute for a chroot-style security mechanism,
// and Sub does not change that fact.
func Sub(fsys FS, dir string) (FS, error) { return stdfs.Sub(fsys, dir) }

// A File provides access to a single file. The File interface is the minimum
// implementation required of the file. Directory files should also implement
// ReadDirFile. A file may implement io.ReaderAt or io.Seeker as optimizations.
type File = stdfs.File

// A FileInfo describes a file and is returned by Stat.
type FileInfo = stdfs.FileInfo

// Stat returns a FileInfo describing the named file from the file system.
//
// If fs implements StatFS, Stat calls fs.Stat.
// Otherwise, Stat opens the file to stat it.
func Stat(fsys FS, name string) (FileInfo, error) { return stdfs.Stat(fsys, name) }

// A FileMode represents a file's mode and permission bits.
// The bits have the same definition on all systems, so that
// information about files can be moved from one system
// to another portably. Not all bits apply to all systems.
// The only required bit is ModeDir for directories.
type FileMode = stdfs.FileMode

// The defined file mode bits are the most significant bits of the FileMode.
// The nine least-significant bits are the standard Unix rwxrwxrwx permissions.
// The values of these bits should be considered part of the public API and
// may be used in wire protocols or disk representations: they must not be
// changed, although new bits might be added.
const (
	// The single letters are the abbreviations
	// used by the String method's formatting.

	ModeDir        FileMode = 1 << (32 - 1 - iota) // d: is a directory
	ModeAppend                                     // a: append-only
	ModeExclusive                                  // l: exclusive use
	ModeTemporary                                  // T: temporary file; Plan 9 only
	ModeSymlink                                    // L: symbolic link
	ModeDevice                                     // D: device file
	ModeNamedPipe                                  // p: named pipe (FIFO)
	ModeSocket                                     // S: Unix domain socket
	ModeSetuid                                     // u: setuid
	ModeSetgid                                     // g: setgid
	ModeCharDevice                                 // c: Unix character device, when ModeDevice is set
	ModeSticky                                     // t: sticky
	ModeIrregular                                  // ?: non-regular file; nothing else is known about this file

	// Mask for the type bits. For regular files, none will be set.
	ModeType = ModeDir | ModeSymlink | ModeNamedPipe | ModeSocket | ModeDevice | ModeCharDevice | ModeIrregular

	ModePerm FileMode = 0777 // Unix permission bits
)

// These constants are relative to permission bits, providing more deep
// information, which you can get from FileMode.
const (
	ModePermUserRead   FileMode = 1 << (8 - iota) // u:r
	ModePermUserWrite                             // u:w
	ModePermUserExec                              // u:x
	ModePermGroupRead                             // g:r
	ModePermGroupWrite                            // g:w
	ModePermGroupExec                             // g:x
	ModePermOtherRead                             // o:r
	ModePermOtherWrite                            // o:w
	ModePermOtherExec                             // o:x

	// for easier handling in code

	ModePermRead  = ModePermUserRead | ModePermGroupRead | ModePermOtherRead
	ModePermWrite = ModePermUserWrite | ModePermGroupWrite | ModePermOtherWrite
	ModePermExec  = ModePermUserExec | ModePermGroupExec | ModePermOtherExec

	ModePermUser  = ModePermUserRead | ModePermUserWrite | ModePermUserExec
	ModePermGroup = ModePermGroupRead | ModePermGroupWrite | ModePermGroupExec
	ModePermOther = ModePermOtherRead | ModePermOtherWrite | ModePermOtherExec

	// internal constants

	wx = ModePermWrite | ModePermExec
	rx = ModePermRead | ModePermExec
	// rw  = ModePermRead | ModePermWrite
	// rwx = ModePermRead | ModePermWrite | ModePermExec
)

type (
	// A GlobFS is a file system with a Glob method.
	GlobFS = stdfs.GlobFS

	// ReadDirFS is the interface implemented by a file system that provides an
	// optimized implementation of ReadDir.
	ReadDirFS = stdfs.ReadDirFS

	// A ReadDirFile is a directory file whose entries can be read with the
	// ReadDir method. Every directory file should implement this interface.
	// (It is permissible for any file to implement this interface, but if so
	// ReadDir should return an error for non-directories.)
	ReadDirFile = stdfs.ReadDirFile

	// ReadFileFS is the interface implemented by a file system that provides an
	// optimized implementation of ReadFile.
	ReadFileFS = stdfs.ReadFileFS

	// A StatFS is a file system with a Stat method.
	StatFS = stdfs.StatFS

	// A SubFS is a file system with a Sub method.
	SubFS = stdfs.SubFS

	// WalkDirFunc is the type of the function called by WalkDir to visit
	// each file or directory.
	//
	// The path argument contains the argument to WalkDir as a prefix.
	// That is, if WalkDir is called with root argument "dir" and finds a file
	// named "a" in that directory, the walk function will be called with
	// argument "dir/a".
	//
	// The d argument is the fs.DirEntry for the named path.
	//
	// The error result returned by the function controls how WalkDir
	// continues. If the function returns the special value SkipDir, WalkDir
	// skips the current directory (path if d.IsDir() is true, otherwise
	// path's parent directory). Otherwise, if the function returns a non-nil
	// error, WalkDir stops entirely and returns that error.
	//
	// The err argument reports an error related to path, signaling that
	// WalkDir will not walk into that directory. The function can decide how
	// to handle that error; as described earlier, returning the error will
	// cause WalkDir to stop walking the entire tree.
	//
	// WalkDir calls the function with a non-nil err argument in two cases.
	//
	// First, if the initial fs.Stat on the root directory fails, WalkDir
	// calls the function with path set to root, d set to nil, and err set to
	// the error from fs.Stat.
	//
	// Second, if a directory's ReadDir method fails, WalkDir calls the
	// function with path set to the directory's path, d set to an
	// fs.DirEntry describing the directory, and err set to the error from
	// ReadDir. In this second case, the function is called twice with the
	// path of the directory: the first call is before the directory read is
	// attempted and has err set to nil, giving the function a chance to
	// return SkipDir and avoid the ReadDir entirely. The second call is
	// after a failed ReadDir and reports the error from ReadDir.
	// (If ReadDir succeeds, there is no second call.)
	//
	// The differences between WalkDirFunc compared to filepath.WalkFunc are:
	//
	//   - The second argument has type fs.DirEntry instead of fs.FileInfo.
	//   - The function is called before reading a directory, to allow SkipDir
	//     to bypass the directory read entirely.
	//   - If a directory read fails, the function is called a second time
	//     for that directory to report the error.
	//
	WalkDirFunc = stdfs.WalkDirFunc
)
