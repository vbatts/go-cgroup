// cmpout
// +build ignore

package main

import (
	"."
	"fmt"
)

func main() {
	cgroup.Init()

	g := cgroup.NewCgroup("foo")
	c := g.AddController("bar")
	fmt.Printf("%#v\n", c)
	c = g.GetController("bar")
	fmt.Printf("%#v\n", c)

	g.SetPermissions(cgroup.Mode(0777), cgroup.Mode(0777), cgroup.Mode(0777))

	ctls, err := cgroup.GetAllControllers()
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := range ctls {
		fmt.Printf("Hierarchy=%d Enabled=%d NumCgroups=%d Name=%s\n", ctls[i].Hierarchy, ctls[i].Enabled, ctls[i].NumCgroups, ctls[i].Name)
	}
}
