package main

import (
	"fmt"

	"github.com/brightpuddle/goaci"
)

func main() {
	backup, err := goaci.NewBackup("config.tar.gz")
	if err != nil {
		panic(err)
	}

	res, _ := backup.GetDn("uni/tn-infra")
	fmt.Println(res.Get("*.attributes.name"))
	// infra

	res, _ = backup.GetClass("fvBD")
	fmt.Println(res.Get("0.fvBD.attributes|@pretty"))
	// {
	//   "dn": "uni/tn-..."
	//   "name": "...",
	//   ...
	// }
}
