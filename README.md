
###Consul-acl-mgr

The consul-acl-mgr is a simple utility to manage your Consul ACLs with a YAML file.

Basic usage:
```
./consul-acl-mgr -f /path/to/acl.yaml -v vvv
```

* -v set's logging output
  * v: warn
  * vv: info
  * vvv:debug
* -f should be a valid yaml file

An example YAML file is acls.yaml and I've added the basics of the YAML file below

```
# The IP or DNS:PORT combination where your consul master is running
consul_cluster: 192.168.33.11:8500

# the acl master token - without this you can't manage ACLs
# DON'T SIMPLY USE THIS VALUE - GENERATE A NEW UUID !!!
acl_master_token: 3f53b9dc-a577-4b07-873c-0216bd9b8696

# Tokens to create. You can generate new IDs with the command line "uuidgen" program
# See https://www.consul.io/docs/internals/acl.html
tokens:
  - # handle setting access for the default token
    department: Common
    team: Default
    token: anonymous 
    name: Anonymous
    type: client
    # IF set to true, the code only issues a delete
    remove: false  
    keys:
      - {name: "", value: deny}
    services:
      - {name: "_rexec", value: deny}
      - {name: "", value: deny}
      - {name: "consul", value: deny}
  -
    department: XPS
    team: DU
    token: 7EBBC145-7475-404B-ABBF-C6C3846B051C
    name: xps/du
    type: client
    # IF set to true, the code only issues a delete
    remove: false
    keys:
      - {name: "xps/du", value: write}
      - {name: "", value: deny}
    services:
      - {name: "xps-du-", value: write}
      - {name: "", value: read}

package main
```

For this to work, you need to have a consul cluster configured with ACL enabled. Imagine a simple acl.json file in your configuration directory e.g.

```
{
  "acl_datacenter": "dc1",
  "acl_default_policy": "deny",
  "acl_down_policy": "allow",
  "acl_master_token": "3f53b9dc-a577-4b07-873c-0216bd9b8696"
}
```

Whatever value you use for the acl_master_token needs to reside in your yaml file.

