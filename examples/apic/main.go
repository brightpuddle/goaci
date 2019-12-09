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
	apic, _ := goaci.NewAPIC(host, user, password)
	err := apic.Login()
	if err != nil {
		panic(err)
	}

	// Get request
	res, _ := apic.Get("/api/mo/uni/tn-infra")
	name := res.Get("imdata.0.*.attributes.name")
	fmt.Println(name)
	// infra

	// Get an MO by DN
	res, _ = apic.GetDn("uni/tn-infra")
	tenantRecord := res.Get("*.attributes|@pretty")
	fmt.Println(tenantRecord)
	// {
	//   "dn": "uni/tn-infra",
	//   "name" "infra",
	//   ...
	// }

	// Get by class
	res, _ = apic.GetClass("fvTenant")
	for _, tnName := range res.Get("#.*.attributes.name").Array() {
		fmt.Println(tnName)
	}
	// infra
	// common
	// mgmt
	// ...

	// Query parameters

	queryInfra := goaci.Query("query-target-filter", `eq(fvTenant.name,"infra")`)
	res, _ = apic.GetClass("fvTenant", queryInfra)
	name = res.Get("imdata.0.fvTenant.attributes.name")
	fmt.Println(name)

	// Create a new tenant
	body := goaci.Body{}.Set("fvTenant.attributes.name", "goaci-example").Str
	res, err = apic.Post("/api/mo/uni/tn-goaci-example", body)
	if err != nil {
		panic(err)
	}
	fmt.Println(res.Get("@pretty"))
}
