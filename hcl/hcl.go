package hcl

import "github.com/hashicorp/hcl/v2/hclsimple"

type HCL struct {
	Host     *Host      `hcl:"host,block" json:"host_config"`
	Products []*Product `hcl:"product,block" json:"products_config"`
}

type Blocks interface {
	*Host | *Product
}

type Host struct {
	Commands []Command `hcl:"command,block"`
	Shells   []Shell   `hcl:"shell,block"`
	GETs     []GET     `hcl:"GET,block"`
	Copies   []Copy    `hcl:"copy,block"`
	Excludes []string  `hcl:"excludes,optional"`
	Selects  []string  `hcl:"selects,optional"`
}

type Product struct {
	Name     string    `hcl:"name,label"`
	Commands []Command `hcl:"command,block"`
	Shells   []Shell   `hcl:"shell,block"`
	GETs     []GET     `hcl:"GET,block"`
	Copies   []Copy    `hcl:"copy,block"`
	Excludes []string  `hcl:"excludes,optional"`
	Selects  []string  `hcl:"selects,optional"`
}

type Command struct {
	Run    string `hcl:"run"`
	Format string `hcl:"format"`
}

type Shell struct {
	Run string `hcl:"run"`
}

type GET struct {
	Path string `hcl:"path"`
}

type Copy struct {
	Path  string `hcl:"path"`
	Since string `hcl:"since,optional"`
}

// Parse takes a file path and decodes the file from disk into HCL types.
func Parse(path string) (HCL, error) {
	var h HCL
	err := hclsimple.DecodeFile(path, nil, &h)
	if err != nil {
		return HCL{}, err
	}
	return h, nil
}
