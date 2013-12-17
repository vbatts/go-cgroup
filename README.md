go-cgroup
=========

Bindings to the libcgroup library


Installing
==========

	go get github.com/vbatts/go-cgroup
 

Sample
======


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


Contributing
============
Fork and Pull Request please!

