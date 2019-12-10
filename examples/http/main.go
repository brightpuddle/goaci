package main

import (
	"fmt"

	"github.com/brightpuddle/goaci"
)

const (
	host     = "1.1.1.1"
	user     = "admin"
	password = "cisco"
)

func main() {
	client, _ := goaci.NewClient(host, user, password)
	err := client.Login()
	if err != nil {
		panic(err)
	}

	// Get request
	res, _ := client.Get("/api/mo/uni/tn-infra")
	name := res.Get("imdata.0.*.attributes.name")
	fmt.Println(name)
	// infra

	// Get an MO by DN
	res, _ = client.GetDn("uni/tn-infra")
	tenantRecord := res.Get("*.attributes|@pretty")
	fmt.Println(tenantRecord)
	// {
	//   "dn": "uni/tn-infra",
	//   "name" "infra",
	//   ...
	// }

	// Get by class
	res, _ = client.GetClass("fvTenant")
	for _, tnName := range res.Get("#.*.attributes.name").Array() {
		fmt.Println(tnName)
	}
	// infra
	// common
	// mgmt
	// ...

	// Query parameters

	queryInfra := goaci.Query("query-target-filter", `eq(fvTenant.name,"infra")`)
	res, _ = client.GetClass("fvTenant", queryInfra)
	name = res.Get("imdata.0.fvTenant.attributes.name")
	fmt.Println(name)

	// Create a new tenant
	body := goaci.Body{}.Set("fvTenant.attributes.name", "goaci-example").Str
	res, err = client.Post("/api/mo/uni/tn-goaci-example", body)
	if err != nil {
		panic(err)
	}
	fmt.Println(res.Get("@pretty"))
}
