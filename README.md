<p align="center">
<img src="logo.png" width="800" height="321" border="0" alt="goACI">
<br/>
ACI client library for Go
<p>
<hr/>

goACI is a client library for Cisco ACI. Itfeatures an HTTP client and a backup
client for running queries directly against backup files.

# Getting Started

## Installing

To start using GoACI, install Go and `go get`:

`$ go get -u github.com/brightpuddle/goaci`

## Initiate the APIC client

```go
package main

import "github.com/brightpuddle/goaci"

apic, _ := goaci.NewAPIC("1.2.3.4", "username", "secretpwd")

err := apic.Login()
if err != nil {
    panic(err)
}
```

## Get an object
Get queries by full path. The `Result` objects is a
[gjson.Result](https://github.com/tidwall/gjson), and has extensive query capabilities.

```go
res, _ = apic.GetDn("uni/tn-infra")
tenantName := res.Get("*.attributes.name")

fmt.Println(tenantRecord)
```

Prints:

`infra`


## Backup client
goACI features a "backup" client that can query ACI `.tar.gz` backup files.
As much as possible this client mirrors the external interface of the HTTP
client facilitating tool development that can run against both the APIC and
the backup.

## Inititate the Backup client
```go
backup, _ := goaci.NewBackup("config.tar.gz")
```


## Get an object by DN
```go
res, _ := backup.GetByDn("uni/tn-infra")
name := res.Get("*.attributes.name")

fmt.Println(name)
```

## Usage

