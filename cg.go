package cgroup

// #include <libcgroup.h>
// #cgo LDFLAGS: -lcgroup
import "C"
import (
	"errors"
	"unsafe"
)

type ControllerData struct {
	Name       string
	Hierarchy  int
	NumCgroups int
	Enabled    int
}

func fromCControllerData(cData C.struct_controller_data) ControllerData {
	name := C.GoString(&cData.name[0])
	return ControllerData{
		Name:      name,
		Hierarchy: int(cData.hierarchy),
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
	ECGEOF = errors.New(C.GoString(C.cgroup_strerror(C.ECGEOF)))
)

func _err(num C.int) error {
	switch num {
	case 0:
		return nil
	case C.ECGEOF:
		return ECGEOF
	}
	return errors.New(C.GoString(C.cgroup_strerror(num)))
}

func Init() error {
	return _err(C.cgroup_init())
}
