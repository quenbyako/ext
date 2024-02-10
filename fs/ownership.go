// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

package fs

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"syscall"
)

// ErrDifferentOwnership says to user that they is not able to make action with
// specific file or directory.
//
// You can throw this error, if you can't do anything with user/group settings,
// and only one way is to change file ownership.
type ErrDifferentOwnership struct {
	// WantID means that user/group must change ownership of some file to
	// this uid or gid to fix issue with ownership.
	//
	// Can be empty, if WantAs is ModePermOther. For other WantAs, WantID will
	// be root.
	//
	// E.g. if user is some_guy with groups [a, b, c], and issued file owned by
	// some_admin with group 'd', we can provide in these fields some_guy and
	// preferred (or random) group from user groups, let's choose 'b'. So error
	// hints user that they need to change file ownership to some_guy:c (or
	// without wanted group, if you don't need to change file owner)
	WantID string
	WantAs FileMode // ModePermUser, ModePermGroup, ModePermOther, can't combine

	// problematic path, which is needed to change (can be different from target
	// path)
	GotPath string
	// operation to run. for now supports only write and create only
	GotOp   Op
	GotUID  string
	GotGID  string
	GotMode FileMode
}

// targetPath which you wanted to make operation and got an error.
// wantAs:   ModePermUser, ModePermGroup or ModePermOther, DON'T COMBINE.
func HintErrPermission(err error, wantID string, wantAs FileMode) error {
	var pathErr *PathError
	if !errors.As(err, &pathErr) || !errors.Is(pathErr.Err, ErrPermission) {
		return err
	}

	// can't be pointer, cause it contains default empty value, we can cause a
	// panic if change to pointer
	var permErr ErrPermissionExtended
	errors.As(pathErr.Err, &permErr) // if fails, we have default empty value

	return ErrDifferentOwnership{
		WantID: wantID,
		WantAs: wantAs,

		GotPath: pathErr.Path,
		GotOp:   opFromString(pathErr.Op),
		GotUID:  permErr.Uid,
		GotGID:  permErr.Gid,
		GotMode: permErr.Mode,
	}
}

func chmodPrefix(wantAs FileMode) string {
	switch wantAs {
	case ModePermUser:
		return "u+"
	case ModePermGroup:
		return "g+"
	default:
		return "o+"
	}
}

func (e ErrDifferentOwnership) chmod() (string, bool) {
	var targetRWX FileMode
	currentRWX := swapSelectedPerms(e.GotMode, e.WantAs)

	// first of all, calculating expected permission bits, depend on operation
	switch e.GotOp {
	case OpReadDir, OpRead:
		targetRWX |= ModePermRead
	case OpExec:
		targetRWX |= ModePermExec
	case OpWrite, OpCreate, OpDelete:
		targetRWX |= ModePermWrite
	default:
		// as fallback we should propose user to set all flags. Why not?
		targetRWX = ModePermRead | ModePermWrite | ModePermExec
	}
	if e.GotMode.IsDir() && (targetRWX&(ModePermRead|ModePermWrite)) > 0 {
		// for dir, always setting exec bit, if read or write was setted.
		targetRWX |= ModePermExec
	}
	// using "other" perm bits for easier handling
	targetRWX &= lastThreeBits

	if targetRWX&currentRWX == targetRWX {
		return "", false
	}

	// now, we need to get all unset bits which we need to set
	// need? 0 already set? 0 => 0
	// need? 0 already set? 1 => 0
	// need? 1 already set? 0 => 1
	// need? 1 already set? 1 => 0
	totalRWX := (targetRWX ^ currentRWX) &^ currentRWX

	// ok, next need to string it, using stdlib method
	perms := chmodPrefix(e.WantAs) + permString(totalRWX)

	return fmt.Sprintf("sudo chmod %v %q", perms, "/"+e.GotPath), true
}

// as must be only ModeUser, ModeGroup or ModeOther
func swapSelectedPerms(mode, as FileMode) FileMode {
	switch as {
	case ModePermUser:
		return mode >> 6 & lastThreeBits
	case ModePermGroup:
		return mode >> 3 & lastThreeBits
	default:
		return mode & lastThreeBits
	}
}

func permString(m FileMode) (s string) {
	if m&ModePermRead > 0 {
		s += "r"
	}
	if m&ModePermWrite > 0 {
		s += "w"
	}
	if m&ModePermExec > 0 {
		s += "x"
	}
	return s
}

func (e ErrDifferentOwnership) Error() string { return e.Unwrap().Error() }

func (e ErrDifferentOwnership) Unwrap() error {
	return &PathError{Op: e.GotOp.String(), Path: e.GotPath, Err: ErrPermission}
}

// ErrPermissionExtended is an ErrPermission, but with more info, collected from
// target
type ErrPermissionExtended struct {
	Uid  string
	Gid  string
	Mode FileMode
}

func (e ErrPermissionExtended) Unwrap() error { return ErrPermission }
func (e ErrPermissionExtended) Error() string { return ErrPermission.Error() }

func permDenied(finfo FileInfo) ErrPermissionExtended {
	uid, gid, _ := FileOwner(finfo)
	return ErrPermissionExtended{Uid: uid, Gid: gid, Mode: finfo.Mode()}
}

// FileID returns user id and group id of file. It supports ONLY unix
// filesystems, if you run file stat under windows or in virtual fs (e.g.
// embed), it won't work.
//
// TODO: Find some good practice, how to fetch ownership from stat interfaces
func FileOwner(stat FileInfo) (uid, gid string, ok bool) { //cover:ignore // alias
	if stat, ok := stat.(FileInfoOwner); ok {
		uid, gid := stat.Owner()
		return uid, gid, true
	}

	// fallback for unix systems: for some reason go doesn't provide usable
	// interface to get file ownership. So in [os.DirFS] implementation, it
	// returns [os.File] type, and in unix, [os.File.Sys] response is always
	// [os.fileStat].
	if sys, ok := stat.Sys().(*syscall.Stat_t); ok {
		return strconv.Itoa(int(sys.Uid)), strconv.Itoa(int(sys.Gid)), true
	}

	// files without explicit permissions are always owned by root
	return "0", "0", false
}

const rootUuid = "0"

// CheckFileAbleToWrite checks that current os procces is able to write to
// specific file or directory, without additional permissions. It solves really
// complex problem of hinting user, what they need to do, to fix access issue.
//
// Most of programs requires sudo access to fix something, however sometimes
// it's not possible and better to say to user, that they needs to ask their
// administrator.
//
// Note: path must be absolute and fs-compatible: User [io/fs.ValidPath] to
// check, that your path won't cause any error.
//
// Note: CheckFileAbleToWrite is related ONLY FOR FILES! It means, if you want
// to create directory, probably you will fail later.
//
// TODO:
//
//   - provide fsys for test environments and for remote filesystems.
//     E.g. if we want to work with s3, or with anything else, it can be
//     implemented in [fs.FS]
func CheckFileAbleToWrite(fsys FS, path string, uid string, gids []string) (err error) {
	switch finfo, err := Stat(fsys, path); {
	// if we can stat file and either user is root or permission for user is valid
	case err == nil && (GetAllowedOperations(finfo, uid, gids)&OpWrite > 0 || uid == rootUuid):
		return nil
	// if we can stat file, but on previous case we failed with checking permissions
	case err == nil:
		return &PathError{Op: "write", Path: path, Err: permDenied(finfo)}
	// If file is even not exist
	case errors.Is(err, ErrNotExist):
		// makes no sense to fall in top dir, if path is even not exist
		return CheckDirAbleToCreateEntry(fsys, filepath.Dir(path), uid, gids)
	// any other stat error (including permission denied error)
	default:
		return err
	}
}

func CheckDirAbleToReadDir(fsys FS, path string, uid string, gids []string) error {
	switch finfo, err := Stat(fsys, path); {
	// if we can stat file and either user is root or permission for user is valid
	case err == nil && (GetAllowedOperations(finfo, uid, gids)&OpReadDir > 0 || uid == rootUuid):
		return nil
	// if we can stat file, but on previous case we failed with checking permissions
	case err == nil:
		return &PathError{Op: "readdir", Path: path, Err: permDenied(finfo)}
	// If file is even not exist
	case errors.Is(err, ErrNotExist):
		// makes no sense to fall in top dir, if path is even not exist
		return &PathError{Op: "readdir", Path: path, Err: ErrNotExist}
	// If fs says, that we can't even see metadata of file
	case errors.Is(err, ErrPermission):
		// fallback: if we can't stat this directory, go upper to check problematic path
		return CheckDirAbleToReadDir(fsys, filepath.Dir(path), uid, gids)
	// any other stat error
	default:
		return err
	}
}

func CheckDirAbleToCreateEntry(fsys FS, path string, uid string, gids []string) error {
	switch finfo, err := Stat(fsys, path); {
	// if we can stat file and either user is root or permission for user is valid
	case err == nil && (GetAllowedOperations(finfo, uid, gids)&OpCreate > 0 || uid == rootUuid):
		return nil
	// if we can stat file, but on previous case we failed with checking permissions
	case err == nil:
		return &PathError{Op: "create", Path: path, Err: permDenied(finfo)}
	// If file is even not exist
	case errors.Is(err, ErrNotExist):
		// lookin below until we will got existed directory, then there will be
		// no error
		return CheckDirAbleToCreateEntry(fsys, filepath.Dir(path), uid, gids)
	// If fs says, that we can't even see metadata of file
	case errors.Is(err, ErrPermission):
		// for permission denied to createdirectory it means that we are not
		// able to read top directory. E.g: if you want to create a/b/c/file.txt
		//
		// a     rwx
		// a/b   ---
		// a/b/c rwx
		//
		// directory a/b won't allow you to do this.
		// so in this case we'll definitely throw an error.
		//
		// if we wouldn't be able to stat dir, than stat function in start of
		// this function
		switch err := CheckDirAbleToReadDir(fsys, filepath.Dir(path), uid, gids).(type) {
		case nil:
			return &PathError{Op: "create", Path: path, Err: errors.New("unknown error")}
		case *PathError:
			return &PathError{Op: "create", Path: err.Path, Err: err.Err}
		default:
			return err
		}
	// any other stat error
	default:
		return err
	}
}

// Description of UNIX permission (to get more context)
//
// For files: you can only read them and write them, without looking on
// directory permissions. So for CRUD operations you have only UD.
//
// IMPORTANT: "for files" means for everything, except directories (symlinks,
// physical links, sockets, devices, or anything else)
//
// STICKY BITS:
//
// - SUID: user create process with owner access. Example:
//
//   if /bin/sudo has perms: --s--x---, root and sudoers can run this file,
//   others are not able to do it. If user is not root and ran file as sudoers,
//   this process will be owned by root:sudoers.
// - SGID: user create process with group access. Example:
//
//   if /bin/sudo has perms: --x--s---, root and super_sudoers can run this file,
//   others are not able to do it. If user is my_user, BUT! if my_user is not in
//   super_sudoers this process will be owned by my_user:super_sudoers. At the
//   same time, access to this process will be allowed to my_user, all my_user's
//   groups, AND for super_sudoers.
// - STICKY: makes no sense for files, only for directories
//
// For directories... Oh, gosh, let's look:
//
// Create entries:
//     Comment: Entry is creating  with user and group  ownership of created
//              process.
//     Required perms: Write + Execute
//     SGID: owner group of new  entry will be always a  group, which owns
//           parent  directory
//
// List entries:
//     Comment: For different operating systems it's really strange behavior: in
//              macos, you need both read + exec for reading dir entry, for
//              linux you can read only names with read only mode. Partial means
//              that you are able to read names only, Can't read dates, perm
//              bits,  ownership, etc
//     Required perms: Read + Execute
//
// Read entry content:
//     Comment: with --x bits you can access to inner directory with hiding it's
// 	         entries
//     Required perms: Read + Exectute
//
// Write entry content + modify file's permissions:
//     Comment: Edit entry = edit metadata. Exception: chown — allowed only to
//              root.
//     Required perms: Write + Execute
//
// Delete entries:
//     Comment: next steps for checking access:
//              1) user owns directory, where entry is located
//              2) If sticky bit set AND user owns entry to delete
//     Required perms: Write + Execute
//     STICKY: If set, entry can be deleted only by user owner, not group, or
//             others
//
// Notes:
// "rename entry" equals to read + delete + create
// "copy entry" equals to read + create
//
// How gorup permission gets:
// "user" is a process. You currently seeing this text through window manager,
// which is running under your username and your primary group.
// so when your process calling for some action, system checks process owner,
// then process group, if both of them not allowed to call this action, ONLY
// THEN system checks process owner groups and check them too.
//
// VERY IMPORTANT: this package doesn't provide mechanics of getting users and
// groups of user, it's working only with uid + gids. So, if you are trying to
// implement it, make sure that you provided process gid to gids list too!

// for file: Inapplicable
// for dir:  Create files/subdirs
//
// mode must be already normalized for user with getPermGroup
//
// SGID makes no sense here, it is a directive for kernel to setup uid and gid
func permCreate(mode FileMode) (op Op) {
	if mode.IsDir() && mode&wx == wx {
		return OpCreate
	}

	return
}

// for file: read content of file WITHOUT reading metadata
// for dir:  list entries without reading
//
// read       means that it's able at least to read entry names
// fullAccess means that it's able also to read metadata of entry: size, date
func permRead(mode FileMode) (op Op) {
	switch {
	case mode.IsDir() && mode&rx == rx:
		return OpReadDir
	case !mode.IsDir() && mode&ModePermRead > 0:
		return OpRead
	default:
		return
	}
}

// for file: Modify file content BUT without modifying metadata (date, perms)
// for dir:  modify entry content or entry metadata (but not file).
//
// WARNING: renaming file is a remove + create operation, this permissions
// affects only to modyfing content!!!
func permModify(mode FileMode) (op Op) {
	if !mode.IsDir() && mode&ModePermWrite > 0 ||
		mode.IsDir() && mode&wx == wx {
		return OpWrite
	}

	return
}

// for file: Inapplicable
// for dir:  delete entry
//
// owner:  uid of entry to delete
// caller: uid of operation requester
func permDeleteEntryOwner(mode FileMode, owner, caller string) (op Op) {
	if mode.IsDir() && (mode&wx == wx ||
		owner == caller && mode&wx|ModeSticky == wx|ModeSticky) {
		return OpDelete
	}

	return
}

type Op uint32

const (
	OpExec  Op = 1 << iota
	OpWrite    // for files: modify file. For dirs: create entry, modify entry metadata and delete files
	OpRead

	// permissions applied only to directories

	OpCreate
	OpReadDir
	OpDelete
)

func opFromString(str string) Op {
	switch str {
	case "exec":
		return OpExec
	case "write":
		return OpWrite
	case "read":
		return OpRead
	case "create":
		return OpCreate
	case "read_dir":
		return OpReadDir
	case "delete":
		return OpDelete
	default:
		return 0
	}
}

func (o Op) String() string {
	switch OpExec {
	case OpExec:
		return "exec"
	case OpWrite:
		return "write"
	case OpRead:
		return "read"
	case OpCreate:
		return "create"
	case OpReadDir:
		return "read_dir"
	case OpDelete:
		return "delete"
	default:
		return "unknown"
	}
}

func GetAllowedOperations(finfo FileInfo, uid string, gids []string) Op {
	ownerID, ownerGID, ok := FileOwner(finfo)
	if !ok {
		// fallback for fileinfo without ownership: allowing only if uid is root
		return 0
	}
	mode := getPermGroup(finfo.Mode(), ownerID, ownerGID, uid, gids)

	return permCreate(mode) |
		permRead(mode) |
		permModify(mode) |
		permDeleteEntryOwner(mode, ownerID, uid)
}

// getPermGroup detects, which rules of access must be applied, and puts this
// group of permission bits in all three bit groups.
func getPermGroup(mode FileMode, ownerID, ownerGID, uid string, gids []string) FileMode {
	p := callerPerms(mode, ownerID, ownerGID, uid, gids)
	return (mode &^ ModePerm) | // cutting perm bits
		p<<6 | p<<3 | p
}

const lastThreeBits = 0b111

// callerPerms returns 3 bits of access related to user who called
// operation
func callerPerms(mode FileMode, ownerID, ownerGID, uid string, gids []string) FileMode {
	switch {
	case uid == ownerID:
		return mode >> 6 & lastThreeBits
	case sliceContainsString(gids, ownerGID):
		return mode >> 3 & lastThreeBits
	default:
		return mode & lastThreeBits
	}
}

// TODO: replace to golang.org/x/exp/slices#Contains
func sliceContainsString(s []string, i string) bool { //cover:ignore
	for _, item := range s {
		if item == i {
			return true
		}
	}

	return false
}
