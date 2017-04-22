package cgroup

/*
#include <libcgroup.h>
#cgo LDFLAGS: -lcgroup

// work around for the 'type' special word
enum cgroup_file_type type_from_file_info(struct cgroup_file_info fi) {
  return fi.type;
}
*/
import "C"
import (
	"errors"
	"runtime"
	"unsafe"
)

// Init initializes libcgroup. Information about mounted hierarchies are
// examined and cached internally (just what's mounted where, not the groups
// themselves).
func Init() error {
	return _err(C.cgroup_init())
}

// Cgroup is the structure describing one or more control groups. The structure
// is opaque to applications.
type Cgroup struct {
	g *C.struct_cgroup
}

// NewCgroup allocates a new cgroup structure. This function itself does not create new
// control group in kernel, only new <tt>struct cgroup</tt> inside libcgroup!
// The caller would still need to Create() or similar to create this group in the kernel.
//
// @param name Path to the group, relative from root group. Use @c "/" or @c "."
// 	for the root group itself and @c "/foo/bar/baz" or @c "foo/bar/baz" for
// 	subgroups.
func NewCgroup(name string) Cgroup {
	cg := Cgroup{
		C.cgroup_new_cgroup(C.CString(name)),
	}
	runtime.SetFinalizer(&cg, freeCgroupThings)
	return cg
}

// AddController attaches a new controller to cgroup. This function just
// modifies internal libcgroup structure, not the kernel control group.
func (cg *Cgroup) AddController(name string) *Controller {
	return &Controller{
		C.cgroup_add_controller(cg.g, C.CString(name)),
	}
}

// GetController returns appropriate controller from given group.
// The controller must be added before using AddController() or loaded
// from kernel using GetCgroup().
func (cg Cgroup) GetController(name string) *Controller {
	return &Controller{
		C.cgroup_get_controller(cg.g, C.CString(name)),
	}
}

func freeCgroupThings(cg *Cgroup) {
	C.cgroup_free(&cg.g)
	C.cgroup_free_controllers(cg.g)
}

/*
Create a control group in kernel. The group is created in all
hierarchies, which cover controllers added by Cgroup.AddController().

TODO correct docs for golang implementation

All parameters set by cgroup_add_value_* functions are written.
The created groups has owner which was set by cgroup_set_uid_gid() and
permissions set by cgroup_set_permissions.

  foo = cgroup.NewCgroup("foo)
  foo.Create()
*/
func (cg Cgroup) Create() error {
	return _err(C.cgroup_create_cgroup(cg.g, C.int(0)))
}

// CreateIgnoreOwnership is the same as Create(), but all errors are ignored
// when setting owner of the group and/or its tasks file.
func (cg Cgroup) CreateIgnoreOwnership() error {
	return _err(C.cgroup_create_cgroup(cg.g, C.int(1)))
}

/*
CreateFromParent creates new control group in kernel, with all parameters and
values copied from its parent group. The group is created in all hierarchies,
where the parent group exists. I.e. following code creates subgroup in all
hierarchies, because all of them have root (=parent) group.

  foo = cgroup.NewCgroup("foo)
  foo.CreateFromParent()

*/
func (cg Cgroup) CreateFromParent() error {
	return _err(C.cgroup_create_cgroup_from_parent(cg.g, C.int(0)))
}

// CreateFromParentIgnoreOwnership is the same as CreateFromParent(), but all
// errors are ignored when setting owner of the group and/or its tasks file.
func (cg Cgroup) CreateFromParentIgnoreOwnership() error {
	return _err(C.cgroup_create_cgroup_from_parent(cg.g, C.int(1)))
}

// Modify a control group in kernel. All parameters added by cgroup_add_value_
// or cgroup_set_value_ are written.  Currently it's not possible to change and
// owner of a group.
//
// TODO correct docs for golang implementation
func (cg Cgroup) Modify() error {
	return _err(C.cgroup_modify_cgroup(cg.g))
}

/*
Delete removes a control group from kernel. The group is removed from
all hierarchies,  which cover controllers added by Cgroup.AddController()
or GetCgroup(). All tasks inside the group are automatically moved
to parent group.

The group being removed must be empty, i.e. without subgroups. Use
cgroup_delete_cgroup_ext() for recursive delete.

TODO correct docs for golang implementation
*/
func (cg Cgroup) Delete() error {
	return _err(C.cgroup_delete_cgroup(cg.g, C.int(0)))
}

// DeleteIgnoreMigration is the same as Delete(), but ignores errors when
// migrating.
func (cg Cgroup) DeleteIgnoreMigration() error {
	return _err(C.cgroup_delete_cgroup(cg.g, C.int(1)))
}

/*
DeleteExt removes a control group from kernel.

All tasks are automatically moved to parent group.
If DeleteIgnoreMigration flag is used, the errors that occurred
during the task movement are ignored.

DeleteRecursive flag specifies that all subgroups should be removed
too. If root group is being removed with this flag specified, all subgroups
are removed but the root group itself is left undeleted.
*/
func (cg Cgroup) DeleteExt(flags DeleteFlag) error {
	return _err(C.cgroup_delete_cgroup_ext(cg.g, C.int(flags)))
}

/*
Get reads all information regarding the group from kernel.
Based on name of the group, list of controllers and all parameters and their
values are read from all hierarchies, where a group with given name exists.
All existing controllers are replaced. I.e. following code will fill root with
controllers from all hierarchies, because the root group is available in all of
them.

  root := cgroup.NewCgroup("/")
  err := root.Get()
  ...

*/
func (cg Cgroup) Get() error {
	return _err(C.cgroup_get_cgroup(cg.g))
}

type (
	// UID is the user ID type.
	//  cgroup.UID(0)
	UID C.uid_t

	// GID is the group ID type.
	//  cgroup.GID(0)
	GID C.gid_t
)

/*
SetUIDGID sets owner of the group control files and the @c tasks file. This function
modifies only libcgroup internal cgroup structure, use
Cgroup.Create() afterwards to create the group with given owners.

@param cgroup
@param tasksUID UID of the owner of group's @c tasks file.
@param tasksGID GID of the owner of group's @c tasks file.
@param controlUID UID of the owner of group's control files (i.e.
parameters).
@param controlGID GID of the owner of group's control files (i.e.
parameters).
*/
func (cg Cgroup) SetUIDGID(tasksUID UID, tasksGID GID,
	controlUID UID, controlGID GID) error {
	return _err(C.cgroup_set_uid_gid(cg.g,
		C.uid_t(tasksUID), C.gid_t(tasksGID),
		C.uid_t(controlUID), C.gid_t(controlGID)))

}

// GetUIDGID returns owners of the group's @c tasks file and control files.
// The data is read from libcgroup internal cgroup structure, use
// Cgroup.SetUIDGID() or Cgroup.Get() to fill it.
func (cg Cgroup) GetUIDGID() (tasksUID UID, tasksGID GID, controlUID UID, controlGID GID, err error) {
	var (
		cTU C.uid_t
		cTG C.gid_t
		cCU C.uid_t
		cCG C.gid_t
	)
	err = _err(C.cgroup_set_uid_gid(cg.g,
		cTU,
		cTG,
		cCU,
		cCG))
	return UID(cTU), GID(cTG), UID(cCU), GID(cCG), err

}

type (
	PID C.pid_t
)

func (cg Cgroup) AttachTaskPid(pid PID) error {
	return _err(C.cgroup_attach_task_pid(cg.g, pid))
}

const (
	// NoPerms is uninitialized file/directory permissions used for task/control files.
	NoPerms = C.NO_PERMS

	// NoUIDGID is uninitialized UID/GID used for task/control files.
	NoUIDGID = C.NO_UID_GID
)

// Mode is the file permissions. Like used in SetPermissions()
type Mode C.mode_t

/*
SetPermissions stores given file permissions of the group's control and tasks files
into the cgroup data structure. Use NoPerms if permissions shouldn't
be changed or a value which applicable to chmod(2). Please note that
the given permissions are masked with the file owner's permissions.
For example if a control file has permissions 640 and controlFilePerm is
471 the result will be 460.

controlDirPerm Directory permission for the group.
controlFilePerm File permission for the control files.
taskFilePerm File permissions for task file.

  g := cgroup.NewCgroup("foo")
  g.SetPermissions(cgroup.Mode(0777), cgroup.Mode(0777), cgroup.Mode(0777))
*/
func (cg Cgroup) SetPermissions(controlDirPerm, controlFilePerm, taskFilePerm Mode) {
	C.cgroup_set_permissions(cg.g, C.mode_t(controlDirPerm),
		C.mode_t(controlFilePerm), C.mode_t(taskFilePerm))
}

// CopyCgroup copies all controllers, parameters and their values. All existing
// controllers in the source group are discarded.
func CopyCgroup(src, dest Cgroup) error {
	return _err(C.cgroup_copy_cgroup(src.g, dest.g))
}

// CompareCgroup compares names, owners, controllers, parameters and values of two groups.
//
// Return value of:
//   * nil - a and b are equal
//   * ErrGroupNotEqual - groups are not equal
//   * ErrControllerNotEqual - controllers are not equal
func CompareCgroup(a, b Cgroup) error {
	return _err(C.cgroup_compare_cgroup(a.g, b.g))
}

// Controller is the structure describing a controller attached to one struct
// @c cgroup, including parameters of the group and their values. The structure
// is opaque to applications.
type Controller struct {
	c *C.struct_cgroup_controller
}

// AddValueString adds parameter and its value to internal libcgroup
// structures.  Use Cgroup.Modify() or Cgroup.Create() to write it to kernel.
//
// Name of the parameter and its value
func (c Controller) AddValueString(name, value string) error {
	return _err(C.cgroup_add_value_string(c.c, C.CString(name), C.CString(value)))
}

func (c Controller) AddValueInt64(name string, value int64) error {
	return _err(C.cgroup_add_value_int64(c.c, C.CString(name), C.int64_t(value)))
}

func (c Controller) AddValueBool(name string, value bool) error {
	return _err(C.cgroup_add_value_bool(c.c, C.CString(name), C.bool(value)))
}

// GetValueString fetches the values from the controller.  Use Cgroup.Get() to
// get the names available to fetch values from the kernel.
func (c Controller) GetValueString(name string) (value string, err error) {
	var v *C.char
	err = _err(C.cgroup_get_value_string(c.c, C.CString(name), &v))
	return C.GoString(v), err
}

func (c Controller) GetValueInt64(name string) (value int64, err error) {
	var v C.int64_t
	err = _err(C.cgroup_get_value_int64(c.c, C.CString(name), &v))
	return int64(v), err
}

func (c Controller) GetValueBool(name string) (value bool, err error) {
	var v C.bool
	err = _err(C.cgroup_get_value_bool(c.c, C.CString(name), &v))
	return bool(v), err
}

// SetValueString sets a parameter value in @c libcgroup internal structures.
// Use Cgroup.Modify() or Cgroup.Create() to write it to kernel.
func (c Controller) SetValueString(name, value string) error {
	return _err(C.cgroup_set_value_string(c.c, C.CString(name), C.CString(value)))
}

func (c Controller) SetValueInt64(name string, value int64) error {
	return _err(C.cgroup_set_value_int64(c.c, C.CString(name), C.int64_t(value)))
}

func (c Controller) SetValueUint64(name string, value uint64) error {
	return _err(C.cgroup_set_value_uint64(c.c, C.CString(name), C.u_int64_t(value)))
}

func (c Controller) SetValueBool(name string, value bool) error {
	return _err(C.cgroup_set_value_bool(c.c, C.CString(name), C.bool(value)))
}

/*
CompareControllers compares names, parameters and values of two controllers.

Return value of:
 * nil - a and b are equal
 * ErrControllerNotEqual - controllers are not equal
*/
func CompareControllers(a, b Controller) error {
	return _err(C.cgroup_compare_controllers(a.c, b.c))
}

// LoadConfig file and mount and create control groups described there.
// See cgconfig.conf(5) man page for format of the file.
func LoadConfig(filename string) error {
	return _err(C.cgroup_config_load_config(C.CString(filename)))
}

// Unload deletes all control groups and unmount all hierarchies.
func Unload() error {
	return _err(C.cgroup_unload_cgroups())
}

type DeleteFlag int

const (
	// DeleteIgnoreMigration ignore errors caused by migration of tasks to parent group.
	DeleteIgnoreMigration = DeleteFlag(C.CGFLAG_DELETE_IGNORE_MIGRATION)

	// DeleteRecursive recursively delete all child groups.
	DeleteRecursive = DeleteFlag(C.CGFLAG_DELETE_RECURSIVE)

	// DeleteEmptyOnly deletes the cgroup only if it is empty, i.e. it has no
	// subgroups and no processes inside. This flag cannot be used with
	// DeleteRecursive
	DeleteEmptyOnly = DeleteFlag(C.CGFLAG_DELETE_EMPTY_ONLY)
)

/*
UnloadFromConfig deletes all cgroups and unmount all mount points defined in
specified config file.

The groups are either removed recursively or only the empty ones, based on
given flags. Mount point are always umounted only if they are empty, regardless
of any flags.

The groups are sorted before they are removed, so the removal of empty ones
actually works (i.e. subgroups are removed first).
*/
func UnloadFromConfig(filename string, flags DeleteFlag) error {
	return _err(C.cgroup_config_unload_config(C.CString(filename), C.int(flags)))
}

/*
SetDefault permissions of groups created by subsequent
cgroup_config_load_config() calls. If a config file contains a 'default {}'
section, the default permissions from the config file is then used.

Use cgroup_new_cgroup() to create a dummy group and cgroup_set_uid_gid() and
cgroup_set_permissions() to set its permissions. Use NoUIDGID instead of
GID/UID and NoPerms instead of file/directory permissions to let kernel
decide the default permissions where you don't want specific user and/or
permissions. Kernel then uses current user/group and permissions from umask
then.

New default permissions from this group are copied to libcgroup internal
structures.
*/
func SetDefault(cg Cgroup) error {
	return _err(C.cgroup_config_set_default(cg.g))
}

// FileInfo is the Information about found cgroup directory (= a control group).
type FileInfo struct {
	fileType FileType
	path     string
	parent   string
	fullPath string
	depth    int8
}

// FileType of this cgroup's
func (fi FileInfo) FileType() FileType {
	return fi.fileType
}

// Path to this cgroup
func (fi FileInfo) Path() string {
	return fi.path
}

// Parent of this cgroup
func (fi FileInfo) Parent() string {
	return fi.parent
}

// FullPath to this cgroup
func (fi FileInfo) FullPath() string {
	return fi.fullPath
}

// Depth of this cgroup
func (fi FileInfo) Depth() int8 {
	return fi.depth
}

func fromCFileInfo(cData C.struct_cgroup_file_info) FileInfo {
	return FileInfo{
		fileType: FileType(C.type_from_file_info(cData)),
		path:     C.GoString(cData.path),
		parent:   C.GoString(cData.parent),
		fullPath: C.GoString(cData.full_path),
		depth:    int8(cData.depth),
	}
}

type FileType int

const (
	FileTypeFile  = FileType(C.CGROUP_FILE_TYPE_FILE)
	FileTypeDir   = FileType(C.CGROUP_FILE_TYPE_DIR)
	FileTypeOther = FileType(C.CGROUP_FILE_TYPE_OTHER)
)

/*
int cgroup_walk_tree_begin(const char *controller, const char *base_path, int depth,
				void **handle, struct cgroup_file_info *info,
				int *base_level);
*/

// ControllerData is the information model for controllers available
type ControllerData struct {
	name       string
	hierarchy  int
	numCgroups int
	enabled    int
}

// Name of the this controller
func (cd ControllerData) Name() string {
	return cd.name
}

// Hierarchy is the identification of the controller. Controllers with the same
// hierarchy ID are mounted together as one hierarchy. Controllers with ID 0
// are not currently mounted anywhere.
func (cd ControllerData) Hierarchy() int {
	return cd.hierarchy
}

// NumCgroups is the number of cgroups
func (cd ControllerData) NumCgroups() int {
	return cd.numCgroups
}

// Enabled indicates whether or not this controller is enabled
func (cd ControllerData) Enabled() int {
	return cd.enabled
}

func fromCControllerData(cData C.struct_controller_data) ControllerData {
	return ControllerData{
		name:       C.GoString(&cData.name[0]),
		hierarchy:  int(cData.hierarchy),
		numCgroups: int(cData.num_cgroups),
		enabled:    int(cData.enabled),
	}
}

func GetAllControllers() (controllers []ControllerData, err error) {
	var (
		cd     C.struct_controller_data
		handle unsafe.Pointer
	)
	err = _err(C.cgroup_get_all_controller_begin(&handle, &cd))
	if err != nil {
		return controllers, err
	}
	defer C.cgroup_get_all_controller_end(&handle)

	controllers = append(controllers, fromCControllerData(cd))
	for {
		err = _err(C.cgroup_get_all_controller_next(&handle, &cd))
		if err != nil {
			if err == ErrEOF {
				break
			}

			return controllers, err
		}
		controllers = append(controllers, fromCControllerData(cd))
	}
	return controllers, nil
}

func GetSubSysMountPoint(controller string) (string, error) {
	var mp *C.char
	err := _err(C.cgroup_get_subsys_mount_point(C.CString(controller), &mp))
	if err != nil {
		return "", err
	}
	return C.GoString(mp), nil
}

// Various errors
var (
	ErrEOF                = errors.New(C.GoString(C.cgroup_strerror(C.ECGEOF)))
	ErrOTHER              = errors.New(C.GoString(C.cgroup_strerror(C.ECGOTHER)))
	ErrGroupNotEqual      = errors.New(C.GoString(C.cgroup_strerror(C.ECGROUPNOTEQUAL)))
	ErrControllerNotEqual = errors.New(C.GoString(C.cgroup_strerror(C.ECGCONTROLLERNOTEQUAL)))
)

// LastError returns last errno, which caused ErrOTHER error.
func LastError() error {
	return _err(C.cgroup_get_last_errno())
}

func _err(num C.int) error {
	switch num {
	case 0:
		return nil
	case C.ECGEOF:
		return ErrEOF
	case C.ECGOTHER:
		return ErrOTHER
	case C.ECGROUPNOTEQUAL:
		return ErrGroupNotEqual
	case C.ECGCONTROLLERNOTEQUAL:
		return ErrControllerNotEqual
	}
	// There's a lot. We'll create them as they come
	return errors.New(C.GoString(C.cgroup_strerror(num)))
}
