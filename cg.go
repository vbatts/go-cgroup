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
	g *C.struct_group
}

func NewCgroup(name string) Cgroup {
	cg := Cgroup{
		C.cgroup_new_cgroup(C.CString(name)),
	}
	runtime.SetFinalizer(cg, freeCgroupThings)
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

func freeCgroupThings(cg Cgroup) {
  freeCgroup(cg)
  freeControllers(cg)
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
*/
func CreateGroup(cg Cgroup, ignore_ownership bool) error {
  var i int = 0
  if ignore_ownership == true {
    i = 1
  }
  return _err(C.cgroup_create_cgroup(cg.g, C.int(i)))
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
		if err != nil && err != ECGEOF {
			return controllers, err
		}
		controllers = append(controllers, fromCControllerData(cd))
		if err == ECGEOF {
			break
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
	ECGEOF   = errors.New(C.GoString(C.cgroup_strerror(C.ECGEOF)))
	ECGOTHER = errors.New(C.GoString(C.cgroup_strerror(C.ECGOTHER)))
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
	}
	// There's a lot. We'll create them as they come
	return errors.New(C.GoString(C.cgroup_strerror(num)))
}
