package nsxtapi

import (
	"github.com/vmware/go-vmware-nsxt/manager"
)

type RouterSearchHandler func(*manager.LogicalRouter, string) bool

/*
  Call back for seach term.
   TODO refactor move name to cosnt for easy consumption

*/
var RouterCallback = map[string]func(r *manager.LogicalRouter, val string) bool{
	"name": func(r *manager.LogicalRouter, name string) bool {
		// lookup by name
		if r.DisplayName == name {
			return true
		}
		return false
	},
	"uuid": func(r *manager.LogicalRouter, uuid string) bool {
		if r.Id == uuid {
			return true
		}
		return false
	},
	"edgeid": func(r *manager.LogicalRouter, edgeid string) bool {
		// edge cluster id
		if r.EdgeClusterId == edgeid {
			return true
		}
		return false
	},
	"type": func(r *manager.LogicalRouter, routerType string) bool {
		// keys TIER0 / TIER1
		if r.RouterType == routerType {
			return true
		}
		return false
	},
}

type EdgeCmpType string
type EdgeSearchHandler func(*manager.EdgeCluster, string) bool

const SearchByName EdgeCmpType = "name"
const SearchByUuid EdgeCmpType = "uuid"

/*
  Call back search term
*/
var EdgeClusterCallback = map[EdgeCmpType]func(r *manager.EdgeCluster, val string) bool{
	SearchByName: func(r *manager.EdgeCluster, name string) bool {
		if r.DisplayName == name {
			return true
		}
		return false
	},
	SearchByUuid: func(r *manager.EdgeCluster, uuid string) bool {
		if r.Id == uuid {
			return true
		}
		return false
	},
}
