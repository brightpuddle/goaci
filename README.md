<p align="center">
<img src="logo.png" width="800" height="321" border="0" alt="goACI">
<br/>
ACI client library for Go
<p>
<hr/>

GoACI is a Go client library for Cisco ACI. It features a simple, extensible API, advanced JSON manipulation, and a [backup
client](#backup-client) for running queries against the backup file.

# Getting Started

## Installing

To start using GoACI, install Go and `go get`:

`$ go get -u github.com/brightpuddle/goaci`

## Basic Usage

```go
package main

import "github.com/brightpuddle/goaci"

func main() {
    apic, _ := goaci.NewAPIC("1.1.1.1", "user", "pwd")
    if err := apic.Login(); err != nil {
        panic(err)
    }

    res, _ = apic.Get("/api/mo/uni/tn-infra")
    println(res.Get("imdata.0.*.attributes.name"))
}
```
This will print:
```
infra
```

### Result manipulation
`goaci.Result` is a [gjson.Result](https://github.com/tidwall/gjson) object, which provide advanced query capabilities:
```go
res, _ := GetClass("fvBD")
res.Get("0.fvBD.attributes.name").Str // name of first BD
res.Get("0.*.attributes.name").Str // name of first BD (if you don't know the class)

for _, bd := range res.Array() {
    println(res.Get("*.attributes|@pretty")) // pretty print BD attributes
}

for _, bd := range res.Get("#.fvBD.attributes").Array() {
    println(res.Get("@pretty") // pretty print BD attributes
}
```
See the [GJSON](https://github.com/tidwall/gjson) documentation for more detail on query options.

### Helpers for common patterns
```go
res, _ := GetDn("uni/tn-infra")
res, _ := GetClass("fvTenant")
```

### Query parameters
Pass the `goaci.Query` object to the `Get` request:

```go
queryInfra := goaci.Query("query-target-filter", `eq(fvTenant.name,"infra")`)
res, _ := apic.GetClass("fvTenant", queryInfra)
```

Pass as many paramters as needed:
```go
res, _ := apic.GetClass("isisRoute",
    goaci.Query("rsp-subtree-include", "relations"),
    goaci.Query("query-target-filter", `eq(isisRoute.pfx,"10.66.0.1/32")`,
)
```

### POST data creation
`goaci.Body` is a wrapper for [SJSON](https://github.com/tidwall/sjson). SJSON supports an advanced path syntax simplifying JSON creation.

```go
exampleTenant := goaci.Body{}.Set("fvTenant.attributes.name", "goaci-example").Str
apic.Post("/api/mo/uni/tn-goaci-example", exampleTenant)
```

These can be chained:
```go
tenantA := goaci.Body{}.
    Set("fvTenant.attributes.name", "goaci-example-a").
    Set("fvTenant.attributes.descr", "Example tenant A")
```

...or nested:
```go
attrs := goaci.Body{}.
    Set("name", goaci-example-b").
    Set("descr", "Example tenant B").
    Str
tenantB := goaaci.Body{}.SetRaw("fvTenant.attributes", attrs).Str
```

### Token refresh
Token refresh is handled automatically. The APIC client keeps a timer and checks elapsed time on each request, refreshing the token every 8 minutes. This can be handled manually if desired:
```go
apic, _ := goaci.NewAPIC("1.1.1.1", "user", "pwd", goaci.NoRefresh)
apic.Login()

// Do some stuff...
apic.Refresh()
```

### Backup client
goACI also features a "backup" client for querying ACI `.tar.gz` backup files. As much as possible this client mirrors the API of the HTTP client.

```go
backup, _ := goaci.NewBackup("config.tar.gz")

// Get the record for the infra tenantj
res, _ := backup.GetDn("uni/tn-infra")
res, _ = backup.GetClass("fvBD")
```
