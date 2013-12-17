// cmpout

// +build ignore

package main

import (
  "."
	"fmt"
)

func main() {
	cgroup.Init()
	fmt.Println(cgroup.DeleteIgnoreMigration)
	fmt.Println(cgroup.DeleteIgnoreMigration|cgroup.DeleteEmptyOnly)

  g := cgroup.NewCgroup("foo")
  c := g.AddController("bar")
  fmt.Printf("%#v\n", c)
  c = g.GetController("bar")
  fmt.Printf("%#v\n", c)

  g.SetPermissions(cgroup.Mode(0777), cgroup.Mode(0777), cgroup.Mode(0777))

	//fmt.Println(cgroup.GetSubSysMountPoint("cpu"))
	ctls, err := cgroup.GetAllControllers()
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := range ctls {
		fmt.Println(ctls[i])
	}
}
