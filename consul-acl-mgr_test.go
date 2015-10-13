package main

import (
	"testing"
)

func TestNewACL(t *testing.T) {

	// call this to get a new UUID
	// 	"github.com/pborman/uuid"
	// uid := uuid.New()

	if uid == "" {
		t.Log("Something's wrong. We didn't get back a UUID")
		t.Fail()
		return
	}
	t.Log("Got a uid ", uid)

	cparser := Config{}
	yamlfile := "/Users/202238/projects/go-projects/src/github.com/russellsimpkins/consul-acl-mgr/acls.yaml"
	err := cparser.ParseYaml(yamlfile)
	if err != nil {
		t.Log("Failure to parse yaml")
		t.Fail()
	}
	t.Log("Consul Cluster: ", cparser.Consul)
	t.Log("Consul acl token: ", cparser.Token)
	t.Log("Consul tokens: ", cparser.Tokens)

	what, _ := cparser.RulesString(cparser.Tokens[0])

	t.Log("json output: ")
	t.Log(what)
	err = cparser.SetConsulACL()
	t.Log(err)
}
