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

/*
Structure describing one or more control groups. The structure is opaque to
applications.
*/
type Cgroup struct {
	g *C.struct_cgroup
}

func NewCgroup(name string) Cgroup {
	cg := Cgroup{
		C.cgroup_new_cgroup(C.CString(name)),
	}
	runtime.SetFinalizer(&cg, freeCgroupThings)
	return cg
}

func (cg *Cgroup) AddController(name string) Controller {
	return Controller{
		C.cgroup_add_controller(cg.g, C.CString(name)),
	}
}
func (cg Cgroup) GetController(name string) Controller {
	return Controller{
		C.cgroup_get_controller(cg.g, C.CString(name)),
	}
}

func freeCgroupThings(cg *Cgroup) {

	freeCgroup(*cg)
	freeControllers(*cg)
}

func freeCgroup(cg Cgroup) {
	C.cgroup_free(&cg.g)
}

func freeControllers(cg Cgroup) {
	C.cgroup_free_controllers(cg.g)
}

/*
Physically create a control group in kernel. The group is created in all
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

/*
Same as Create(), but all errors are ignored when setting
owner of the group and/or its tasks file.
*/
func (cg Cgroup) CreateIgnoreOwnership() error {
	return _err(C.cgroup_create_cgroup(cg.g, C.int(1)))
}

/*
Physically create new control group in kernel, with all parameters and values
copied from its parent group. The group is created in all hierarchies, where
the parent group exists. I.e. following code creates subgroup in all
hierarchies, because all of them have root (=parent) group.

  foo = cgroup.NewCgroup("foo)
  foo.CreateFromParent()

*/
func (cg Cgroup) CreateFromParent() error {
	return _err(C.cgroup_create_cgroup_from_parent(cg.g, C.int(0)))
}

/*
Same as CreateFromParent(), but all errors are ignored when setting
owner of the group and/or its tasks file.
*/
func (cg Cgroup) CreateFromParentIgnoreOwnership() error {
	return _err(C.cgroup_create_cgroup_from_parent(cg.g, C.int(1)))
}

/*
Physically modify a control group in kernel. All parameters added by
cgroup_add_value_ or cgroup_set_value_ are written.
Currently it's not possible to change and owner of a group.

TODO correct docs for golang implementation
*/
func (cg Cgroup) Modify() error {
	return _err(C.cgroup_modify_cgroup(cg.g))
}

/*
Physically remove a control group from kernel. The group is removed from
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

/*
Same as Delete(), but ignores errors when migrating.
*/
func (cg Cgroup) DeleteIgnoreMigration() error {
	return _err(C.cgroup_delete_cgroup(cg.g, C.int(1)))
}

/*
Physically remove a control group from kernel.
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
Read all information regarding the group from kernel.
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

type UID C.uid_t
type GID C.gid_t

/*
Set owner of the group control files and the @c tasks file. This function
modifies only libcgroup internal cgroup structure, use
Cgroup.Create() afterwards to create the group with given owners.

@param cgroup
@param tasks_uid UID of the owner of group's @c tasks file.
@param tasks_gid GID of the owner of group's @c tasks file.
@param control_uid UID of the owner of group's control files (i.e.
parameters).
@param control_gid GID of the owner of group's control files (i.e.
parameters).
*/
func (cg Cgroup) SetUidGid(tasks_uid UID, tasks_gid GID,
	control_uid UID, control_gid GID) error {
	return _err(C.cgroup_set_uid_gid(cg.g,
		C.uid_t(tasks_uid), C.gid_t(tasks_gid),
		C.uid_t(control_uid), C.gid_t(control_gid)))

}

/*
Return owners of the group's @c tasks file and control files.
The data is read from libcgroup internal cgroup structure, use
Cgroup.SetUidGid() or Cgroup.Get() to fill it.
*/
func (cg Cgroup) GetUidGid() (tasks_uid UID, tasks_gid GID, control_uid UID, control_gid GID, err error) {
	var (
		c_t_u C.uid_t
		c_t_g C.gid_t
		c_c_u C.uid_t
		c_c_g C.gid_t
	)
	err = _err(C.cgroup_set_uid_gid(cg.g,
		c_t_u,
		c_t_g,
		c_c_u,
		c_c_g))
	return UID(c_t_u), GID(c_t_g), UID(c_c_u), GID(c_c_g), err

}

const (
	// Uninitialized file/directory permissions used for task/control files.
	NO_PERMS = C.NO_PERMS

	// Uninitialized UID/GID used for task/control files.
	NO_UID_GID = C.NO_UID_GID
)

type Mode C.mode_t

/*
Stores given file permissions of the group's control and tasks files
into the cgroup data structure. Use NO_PERMS if permissions shouldn't
be changed or a value which applicable to chmod(2). Please note that
the given permissions are masked with the file owner's permissions.
For example if a control file has permissions 640 and control_fperm is
471 the result will be 460.

control_dperm Directory permission for the group.
control_fperm File permission for the control files.
task_fperm File permissions for task file.

  g := cgroup.NewCgroup("foo")
  g.SetPermissions(cgroup.Mode(0777), cgroup.Mode(0777), cgroup.Mode(0777))
*/
func (cg Cgroup) SetPermissions(control_dperm, control_fperm, task_fperm Mode) {
	C.cgroup_set_permissions(cg.g, C.mode_t(control_dperm),
		C.mode_t(control_fperm), C.mode_t(task_fperm))
}

/*
Copy all controllers, parameters and their values. All existing controllers
in the source group are discarded.
*/
func CopyCgroup(src, dest Cgroup) error {
	return _err(C.cgroup_copy_cgroup(src.g, dest.g))
}

/*
Compare names, owners, controllers, parameters and values of two groups.

Return value of:
 * nil - a and b are equal
 * ECGROUPNOTEQUAL - groups are not equal
 * ECGCONTROLLERNOTEQUAL - controllers are not equal

*/
func CompareCgroup(a, b Cgroup) error {
	return _err(C.cgroup_compare_cgroup(a.g, b.g))
}

/*
Structure describing a controller attached to one struct @c cgroup, including
parameters of the group and their values. The structure is opaque to
applications.
*/
type Controller struct {
	c *C.struct_cgroup_controller
}

/*
Add parameter and its value to internal libcgroup structures.
Use Cgroup.Modify() or Cgroup.Create() to write it to kernel.

Name of the parameter and its value
*/
func (c Controller) AddValueString(name, value string) error {
	return _err(C.cgroup_add_value_string(c.c, C.CString(name), C.CString(value)))
}

func (c Controller) AddValueInt64(name string, value int64) error {
	return _err(C.cgroup_add_value_int64(c.c, C.CString(name), C.int64_t(value)))
}

func (c Controller) AddValueBool(name string, value bool) error {
	return _err(C.cgroup_add_value_bool(c.c, C.CString(name), C.bool(value)))
}

/*
Use Cgroup.Get() to fill these values with data from the kernel
*/
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

/*
Set a parameter value in @c libcgroup internal structures.
Use Cgroup.Modify() or Cgroup.Create() to write it to kernel.
*/
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
Compare names, parameters and values of two controllers.

Return value of:
 * nil - a and b are equal
 * ECGCONTROLLERNOTEQUAL - controllers are not equal
*/
func CompareControllers(a, b Controller) error {
	return _err(C.cgroup_compare_controllers(a.c, b.c))
}

/*
Initialize libcgroup. Information about mounted hierarchies are examined
and cached internally (just what's mounted where, not the groups themselves).
*/
func Init() error {
	return _err(C.cgroup_init())
}

/*
Load configuration file and mount and create control groups described there.
See cgconfig.conf man page for format of the file.
*/
func LoadConfig(filename string) error {
	return _err(C.cgroup_config_load_config(C.CString(filename)))
}

/*
Delete all control groups and unmount all hierarchies.
*/
func Unload() error {
	return _err(C.cgroup_unload_cgroups())
}

type DeleteFlag int

const (
	// Ignore errors caused by migration of tasks to parent group.
	DeleteIgnoreMigration = DeleteFlag(C.CGFLAG_DELETE_IGNORE_MIGRATION)

	// Recursively delete all child groups.
	DeleteRecursive = DeleteFlag(C.CGFLAG_DELETE_RECURSIVE)

	/*
		Delete the cgroup only if it is empty, i.e. it has no subgroups and
		no processes inside. This flag cannot be used with
		DeleteRecursive
	*/
	DeleteEmptyOnly = DeleteFlag(C.CGFLAG_DELETE_EMPTY_ONLY)
)

/*
Delete all cgroups and unmount all mount points defined in specified config
file.

The groups are either removed recursively or only the empty ones, based
on given flags. Mount point are always umounted only if they are empty,
regardless of any flags.

The groups are sorted before they are removed, so the removal of empty ones
actually works (i.e. subgroups are removed first).
*/
func UnloadFromConfig(filename string, flags DeleteFlag) error {
	return _err(C.cgroup_config_unload_config(C.CString(filename), C.int(flags)))
}

/*
Sets default permissions of groups created by subsequent
cgroup_config_load_config() calls. If a config file contains a 'default {}'
section, the default permissions from the config file is then used.

Use cgroup_new_cgroup() to create a dummy group and cgroup_set_uid_gid() and
cgroup_set_permissions() to set its permissions. Use NO_UID_GID instead of
GID/UID and NO_PERMS instead of file/directory permissions to let kernel
decide the default permissions where you don't want specific user and/or
permissions. Kernel then uses current user/group and permissions from umask
then.

New default permissions from this group are copied to libcgroup internal
structures.
*/
func SetDefault(cg Cgroup) error {
	return _err(C.cgroup_config_set_default(cg.g))
}

type FileInfo struct {
	Type     FileType
	Path     string
	Parent   string
	FullPath string
	Depth    int8
}

func fromCFileInfo(cData C.struct_cgroup_file_info) FileInfo {
	return FileInfo{
		Type:     FileType(C.type_from_file_info(cData)),
		Path:     C.GoString(cData.path),
		Parent:   C.GoString(cData.parent),
		FullPath: C.GoString(cData.full_path),
		Depth:    int8(cData.depth),
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

/*
Information model for Controllers available
*/
type ControllerData struct {
	Name       string
	Hierarchy  int
	NumCgroups int
	Enabled    int
}

func fromCControllerData(cData C.struct_controller_data) ControllerData {
	return ControllerData{
		Name:       C.GoString(&cData.name[0]),
		Hierarchy:  int(cData.hierarchy),
		NumCgroups: int(cData.num_cgroups),
		Enabled:    int(cData.enabled),
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
			if err == ECGEOF {
				break
			}

			return controllers, err
		} else {
			controllers = append(controllers, fromCControllerData(cd))
		}
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

var (
	// End-of-file for iterators
	ECGEOF                = errors.New(C.GoString(C.cgroup_strerror(C.ECGEOF)))
	ECGOTHER              = errors.New(C.GoString(C.cgroup_strerror(C.ECGOTHER)))
	ECGROUPNOTEQUAL       = errors.New(C.GoString(C.cgroup_strerror(C.ECGROUPNOTEQUAL)))
	ECGCONTROLLERNOTEQUAL = errors.New(C.GoString(C.cgroup_strerror(C.ECGCONTROLLERNOTEQUAL)))
)

/*
Return last errno, which caused ECGOTHER error.
*/
func LastError() error {
	return _err(C.cgroup_get_last_errno())
}

func _err(num C.int) error {
	switch num {
	case 0:
		return nil
	case C.ECGEOF:
		return ECGEOF
	case C.ECGOTHER:
		return ECGOTHER
	case C.ECGROUPNOTEQUAL:
		return ECGROUPNOTEQUAL
	case C.ECGCONTROLLERNOTEQUAL:
		return ECGCONTROLLERNOTEQUAL
	}
	// There's a lot. We'll create them as they come
	return errors.New(C.GoString(C.cgroup_strerror(num)))
}
