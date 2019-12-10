package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/tidwall/gjson"
)

// Client is a ACI backup file client.
type Client struct {
	dns     map[string]*Res
	classes map[string][]*Res
}

func fmtRn(template string, record gjson.Result) (rn string) {
	// String templating state machine
	type State struct {
		inVariable  bool
		isBracketed bool
		varName     string
	}
	state := State{}

	// Iterate through characters and build the RN
	for _, c := range template {
		switch {
		case c == '{': // start of a variable
			state.inVariable = true
		case c == '[' || c == ']':
			state.isBracketed = true
		case c == '}': // end of a variable
			value := record.Get(state.varName).Str
			if state.isBracketed {
				value = "[" + value + "]"
			}
			rn += value
			// Reset variable state
			state = State{}
		case state.inVariable:
			state.varName += string(c)
		default:
			rn += string(c)
		}
	}
	return rn
}

func buildDn(record gjson.Result, parentDn []string, class string) ([]string, error) {
	// If record already has a DN just return it
	dn := record.Get("dn").Str
	if dn != "" {
		return strings.Split(dn, "/"), nil
	}

	// Get the RN template from the lookup table
	rnTemplate, ok := rnTemplates[class]
	if !ok {
		return []string{}, fmt.Errorf("rn template not found for %s", class)
	}

	// Parse the RN template
	rn := fmtRn(rnTemplate, record)

	return append(parentDn, rn), nil
}

// NewClient creates a new backup file client.
func NewClient(src string) (Client, error) {
	// Open backup file
	f, err := os.Open(src)
	if err != nil {
		return Client{}, err
	}
	defer f.Close()

	// Unzip backup tar.gz file
	gzf, err := gzip.NewReader(f)
	if err != nil {
		return Client{}, err
	}
	defer gzf.Close()

	// Initialize client
	client := Client{
		dns:     make(map[string]*Res),
		classes: make(map[string][]*Res),
	}

	// Untar backup tar file
	tarReader := tar.NewReader(gzf)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return Client{}, err
		}
		info := header.FileInfo()
		name := info.Name()
		if strings.HasSuffix(name, ".json") {
			data, err := ioutil.ReadAll(tarReader)
			if err != nil {
				return Client{}, err
			}
			client.addToDB(gjson.ParseBytes(data))
		}
	}
	return client, nil
}

func (client Client) addToDB(root gjson.Result) {
	type MO struct {
		object   gjson.Result
		parentDn []string
		class    string
	}
	// Create stack and populate root node
	stack := []MO{{object: root}}

	for len(stack) > 0 {
		// Pop item off stack j
		mo := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Get class/body for the current mo
		var moBody gjson.Result
		mo.object.ForEach(func(key, value gjson.Result) bool {
			mo.class = key.String()
			moBody = value
			return false
		})

		// Get DN of current object
		thisDn, err := buildDn(moBody.Get("attributes"), mo.parentDn, mo.class)
		if err == nil {
			dn := strings.Join(thisDn, "/")

			json := Body{}.
				SetRaw(mo.class+".attributes", moBody.Get("attributes").Raw). // Remove children
				Set(mo.class+".attributes.dn", dn).                           // Fix the DN
				gjson()

			client.dns[dn] = &json
			client.classes[mo.class] = append(client.classes[mo.class], &json)
		}

		// Add children of this MO to stack
		for _, child := range moBody.Get("children").Array() {
			stack = append(stack, MO{object: child, parentDn: thisDn})
		}
	}
}

// GetClass queries the backup file for an MO class.
func (client Client) GetClass(class string, mods ...func(*Req)) (Res, error) {
	res, ok := client.classes[class]
	if !ok {
		return Res{}, fmt.Errorf("%s not found", class)
	}
	return gjson.Parse(fmt.Sprintf("%v", res)), nil
}

// GetDn queries the backup for a specific DN.
// This returns a single object of the format:
//   { "moClass":
//       "attributes": {
//       ...
//       }
//   }
//
// For unknown class types, retrieve the attributes with a wildcard:
//   res.Get("*.attributes")
func (client Client) GetDn(dn string, mods ...func(*Req)) (Res, error) {
	res, ok := client.dns[dn]
	if !ok {
		return Res{}, fmt.Errorf("%s not fund", dn)
	}
	return *res, nil
}
