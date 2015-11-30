package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	log "github.com/russellsimpkins/consul-acl-mgr/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	yaml "github.com/russellsimpkins/consul-acl-mgr/Godeps/_workspace/src/gopkg.in/yaml.v2"
)

// I need something that can read in YAML files and generate my Consul ACLs with known UUIDs
// so that I can hand out the UUIDs to my internal clients and have tight control over the ACLs
// to support revoking ACLs, updating ACLs and changing the ACL token at given intervals.
// The code here will read one or more YAML files and update the consul using the HTTP API.
// Then I can store the ACL file using git-crypt.

// for name value pairs we use in the keys and services for ACLs
type Pair struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// an ACL entry in our yaml file
type ACL struct {
	Department string
	Team       string
	Token      string
	Name       string
	Type       string
	Remove     bool
	Keys       []Pair
	Services   []Pair
}

// Consul's ACLs are a little odd, in that the rules are string encoded json
type ConsulACL struct {
	ID    string
	Name  string
	Type  string
	Rules string
}

// Consul service definition
type ConsulService struct {
	Id      string
	Service string
	Address string
	Port    uint64
	Tags    []string
}

// Consul Node definition
type ConsulNode struct {
	Node       string
	Address    string
	Datacenter string
	Service    ConsulService
}

// the base configurion in our YAML file
type Config struct {
	Consul     string `yaml:"consul_cluster"` // which IP/DNS address should we be talking to
	Token      string `yaml:"token"`          // we need the master token to keep things up to date
	UpdateAcl  bool   `yaml:"update_acl"`
	AddNodes   bool   `yaml:"add_nodes"`
	AddKeys    bool   `yaml:"add_keys"`
	Datacenter string
	Nodes      []ConsulNode
	Tokens     []ACL
	KeyValues  []Pair `yaml:"keys"`
}

// struct to output Consul ACL Rules as JSON
// e.g. {"key":{ "some/key": {"policy": "write}}}
type ACLRule struct {
	Key     map[string]map[string]string `json:"key"`
	Service map[string]map[string]string `json:"service"`
}

// the functions we need to update consul
type ACLParser interface {
	ParseYaml(yamlFile string) (err error) // parse the yaml file in the directory to read in all ACL information
	SetConsulACL() (err error)             // update consul via the API - should we remove all ACLs, default is false
	RulesString(token ACL) (json string, err error)
	AddConsulNodes() (err error)
	AddConsulKvPairs() (err error)
}

// this function reads a yaml file to populate our Config struct
func (csl *Config) ParseYaml(yamlFile string) (err error) {

	file, err := os.Open(yamlFile)
	if err != nil {
		return err
	}
	_ = file.Close()
	data, err := ioutil.ReadFile(yamlFile)
	if nil != err {
		panic(err)
	}

	err = yaml.Unmarshal(data, csl)
	if err != nil {
		panic(err)
	}

	return
}

// Output consul ACL rules string
func (c *Config) RulesString(token ACL) (result string, err error) {

	data := ACLRule{}
	var body []byte

	data.Key = make(map[string]map[string]string)
	data.Service = make(map[string]map[string]string)

	for _, pair := range token.Keys {
		item := make(map[string]string)
		item["policy"] = pair.Value
		data.Key[pair.Name] = item
	}

	for _, pair := range token.Services {
		item := make(map[string]string)
		item["policy"] = pair.Value
		data.Service[pair.Name] = item
	}

	body, err = json.MarshalIndent(data, "", "  ")

	if err != nil {
		return
	}
	result = string(body)
	return
}

// call this fuction to add nodes
func (c *Config) AddConsulNodes() (err error) {

	// iterate over all nodes and take appropriate action
	for _, node := range c.Nodes {
		var (
			request *http.Request
			client  *http.Client
			resp    *http.Response
			url     string
			putdata []byte
		)
		client = &http.Client{
			CheckRedirect: nil,
		}
		url = fmt.Sprintf("http://%s/v1/catalog/register?token=%s", c.Consul, c.Token)
		log.Debugf("Attempting to register the service using URL: %s", url)
		node.Datacenter = c.Datacenter
		//putdata, err := json.MarshalIndent(node, "", "	")
		putdata, err := json.Marshal(node)
		if err != nil {
			log.Fatalf("Unable to create JSON from our structure. This could only be caused by a bug in your yaml: %s", err)
			return err
		}

		log.Debugf("The token data to PUT: %s", string(putdata))
		request, err = http.NewRequest("PUT", url, nil)
		if err != nil {
			log.Warnf("Unable to register the node. Err: %s", err)
			return err
		}
		reqdata := strings.NewReader(string(putdata))
		request, err = http.NewRequest("PUT", url, reqdata)

		resp, err = client.Do(request)

		if err != nil {
			log.Fatalf("There was a problem calling the Consul server: %s", err)
			return err
		}
		if resp.StatusCode != 200 {
			log.Fatalf("There was a problem calling the Consul server. Response Code %i", resp.StatusCode)
			return err
		}

	}
	return err
}

// call this fuction to add nodes
func (c *Config) AddConsulKvPairs() (err error) {
	// iterate over all nodes and take appropriate action
	for _, pair := range c.KeyValues {
		var (
			request *http.Request
			client  *http.Client
			resp    *http.Response
			url     string
		)
		client = &http.Client{
			CheckRedirect: nil,
		}
		url = fmt.Sprintf("http://%s/v1/kv/%s?token=%s", c.Consul, pair.Name, c.Token)
		log.Debugf("Attempting to register the service using URL: %s", url)
		log.Debugf("The token data to PUT: %s", pair.Value)
		request, err = http.NewRequest("PUT", url, nil)
		if err != nil {
			log.Warnf("Unable to register the node. Err: %s", err)
			return err
		}
		reqdata := strings.NewReader(pair.Value)
		request, err = http.NewRequest("PUT", url, reqdata)

		resp, err = client.Do(request)

		if err != nil {
			log.Fatalf("There was a problem calling the Consul server: %s", err)
			return err
		}
		if resp.StatusCode != 200 {
			log.Fatalf("There was a problem calling the Consul server. Response Code %i", resp.StatusCode)
			return err
		}

	}
	return err
}

// does the logic to set the tokens on the Consul cluster
// if a token is set to be removed, it does the logic to delete
// the token.
func (c *Config) SetConsulACL() (err error) {

	// iterate over all tokens and take appropriate action
	for _, token := range c.Tokens {

		var (
			request *http.Request
			client  *http.Client
			resp    *http.Response
			url     string
			tok     ConsulACL
		)
		tok = ConsulACL{}
		client = &http.Client{
			CheckRedirect: nil,
		}
		tok.ID = token.Token

		if token.Remove {
			url = fmt.Sprintf("http://%s/v1/acl/destroy/%s?token=%s", c.Consul, tok.ID, c.Token)
			log.Debugf("Attempting to destroy an ACL. Using URL: %s", url)
			request, err = http.NewRequest("PUT", url, nil)
			if err != nil {
				log.Warnf("Unable to destroy the ACL, it's possible it was already deleted. %s", err)
			}

		} else {
			url = fmt.Sprintf("http://%s/v1/acl/destroy/%s?token=%s", c.Consul, tok.ID, c.Token)
			log.Debugf("Cancel URL: %s", url)
			request, err = http.NewRequest("PUT", url, nil)

			if err != nil {
				log.Fatalf("Unable to issue PUT request to desroy an existing token: %s", url)
				return err
			}

			url = fmt.Sprintf("http://%s/v1/acl/create?token=%s", c.Consul, c.Token)
			log.Debugf("Create URL: %s", url)

			tok.Name = token.Name
			tok.Type = token.Type
			tok.Rules, err = c.RulesString(token)

			if err != nil {
				log.Fatalf("Unable to create JSON string of the rules: %s", err)
				return err
			}

			tokdata, err := json.MarshalIndent(tok, "", "  ")

			if err != nil {
				log.Fatalf("Unable to create JSON from our structure. This could only be caused by a bug in your code: %s", err)
				return err
			}

			log.Debugf("The token data to PUT: %s", string(tokdata))
			reqdata := strings.NewReader(string(tokdata))
			request, err = http.NewRequest("PUT", url, reqdata)
		}

		resp, err = client.Do(request)

		if err != nil {
			log.Fatalf("There was a problem calling the Consul server: %s", err)
			return err
		}
		if resp.StatusCode != 200 {
			log.Fatalf("There was a problem calling the Consul server. Response Code %i", resp.StatusCode)
			return
		}
	}
	log.Info("Ran all updates against Consul at: ", c.Consul)
	return
}

// Manage your Consul ACLs using a YAML file. There are 2 flags you
// can run with -f and -v. -f specify the YAML file location and -v
// set's the log level
func main() {

	yamlFile := flag.String("f", "", "/abs/path/to/yaml/file.yml that has your Consul ACL data.")
	logLevel := flag.String("v", "", "set your log level, v: Warn vv: Info vvv: Debug")
	flag.Parse()

	switch *logLevel {
	case "":
		log.SetLevel(log.ErrorLevel)
	case "v":
		log.SetLevel(log.WarnLevel)
		break
	case "vv":
		log.SetLevel(log.InfoLevel)
		break
	case "vvv":
		log.SetLevel(log.DebugLevel)
	}

	if *yamlFile == "" {
		log.Error("You need to specify a valid yaml file with the -f flag.\nconsul-acl  -f /path/to/acl.yaml")
		return
	}

	cparser := Config{}

	err := cparser.ParseYaml(*yamlFile)
	if err != nil {
		log.Error(err)
		return
	}

	if cparser.UpdateAcl {
		log.Debug("Set Consul ACL")
		err = cparser.SetConsulACL()

		if err != nil {
			log.Error("There were problems updating the consul server: ", err)
		}
	}

	if cparser.AddNodes {
		log.Debug("Add node(s) to Consul.")
		err = cparser.AddConsulNodes()

		if err != nil {
			log.Error("There were problems adding nodes to consul")
		}
	}
	if cparser.AddKeys {
		log.Debug("Adding key value pair(s) to Consul.")

		err = cparser.AddConsulKvPairs()
		if err != nil {
			log.Error("There were problems adding nodes to consul")
		}
	}
	log.Debug("All done.")
}
