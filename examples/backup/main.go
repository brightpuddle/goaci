package main

import (
	"fmt"

	"github.com/brightpuddle/goaci"
)

func main() {
	client, err := goaci.NewBackup("config.tar.gz")
	if err != nil {
		panic(err)
	}

	res, _ := client.GetDn("uni/tn-infra")
	fmt.Println(res.Get("*.attributes.name"))
	// infra

	res, _ = client.GetClass("fvBD")
	fmt.Println(res.Get("0.fvBD.attributes|@pretty"))
	// {
	//   "dn": "uni/tn-..."
	//   "name": "...",
	//   ...
	// }
}
