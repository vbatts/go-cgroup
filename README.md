go-cgroup
=========

Bindings to the libcgroup library


Installing
==========

	go get github.com/vbatts/go-cgroup
 

on debian, you'll need packages: golang, libcgroup-dev
on fedora, you'll need packages: golang, libcgroup-devel

Sample
======

```
package main

import "github.com/vbatts/go-cgroup"
import "fmt"

func main() {
  cgroup.Init()

  fmt.Println(cgroup.GetSubSysMountPoint("cpu"))

  ctls, err := cgroup.GetAllControllers()
  if err != nil {
    fmt.Println(err)
    return
  }
  for i := range ctls {
    fmt.Println(ctls[i])
  }

}
```

Contributing
============
Fork and Pull Request please!

