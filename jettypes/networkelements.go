package jettypes

import (
	"encoding/json"
	"fmt"
)

/*
 A generic switch required for jettison.
 we need be able identify a a switch and vim need provide some sort of
 uuid.
*/
type GenericSwitch struct {
	//// name
	name string

	// switch uuid
	uuid string

	// dhcp attached to this logical switch
	dhcpUuid string `yaml:"dhcpUuid"`

	// a router attached to this logical switch
	routerUuid string `yaml:"routerUuid"`

	// a port attached to this logical switch
	routerPortUuid string

	// indicate weather vm attached or not
	attached bool
}

func (s *GenericSwitch) Attached() bool {
	if s != nil {
		return s.attached
	}
	return false
}

func (s *GenericSwitch) SetAttached(attached bool) {
	if s != nil {
		s.attached = attached
	}
}

func (s *GenericSwitch) PrintAsJson() {
	var p []byte
	p, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s \n", p)
}

func (s *GenericSwitch) RouterUuid() string {
	return s.routerUuid
}

func (s *GenericSwitch) SetRouterUuid(routerUuid string) {
	s.routerUuid = routerUuid
}

func (s *GenericSwitch) RouterPortUuid() string {
	return s.routerPortUuid
}

func (s *GenericSwitch) SetRouterPortUuid(routerPortUuid string) {
	s.routerPortUuid = routerPortUuid
}

//
func NewGenericSwitch(name string, uuid string, dhcpUuid string, routerUuid string) *GenericSwitch {

	//	&GenericSwitch{&genericSwitchJson{name: name, uuid: uuid, dhcpUuid: dhcpUuid, routerUuid: routerUuid}}

	//	s := genericSwitchJson{name: name, uuid: uuid, dhcpUuid: dhcpUuid, routerUuid: routerUuid}

	return &GenericSwitch{name: name, uuid: uuid, dhcpUuid: dhcpUuid, routerUuid: routerUuid}
}

func (s *GenericSwitch) DhcpUuid() string {
	return s.dhcpUuid
}

func (s *GenericSwitch) SetDhcpUuid(dhcpUuid string) {
	s.dhcpUuid = dhcpUuid
}

func (s *GenericSwitch) Name() string {
	return s.name
}

func (s *GenericSwitch) SetName(name string) {
	s.name = name
}

func (s *GenericSwitch) Uuid() string {
	return s.uuid
}

func (s *GenericSwitch) SetUuid(uuid string) {
	s.uuid = uuid
}

func (s *GenericSwitch) String() string {
	return s.name + ":" + s.uuid + " dhcp:" + s.dhcpUuid
}

func (w GenericSwitch) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			SomeField  string `json:"dhcp-uuid"`
			SomeField2 string `json:"switch-uuid"`
			SomeField3 bool   `json:"attached"`
			SomeField4 string `json:"router-uuid"`
		}{
			SomeField: w.dhcpUuid, SomeField2: w.uuid, SomeField3: w.attached, SomeField4: w.routerUuid,
		})
}

/*
  Generic router need to have a identifier uuid and name
*/
type GenericRouter struct {
	name string `yaml:"logicalRouter"`

	uuid string `yaml:"logicalRouterUuid"`

	switchPortUuid string // switch port uuid
}

func (g *GenericRouter) SwitchPortUuid() string {
	return g.switchPortUuid
}

func (g *GenericRouter) SetSwitchPortUuid(switchPortUuid string) {
	g.switchPortUuid = switchPortUuid
}

func NewGenericRouter(name string, uuid string) *GenericRouter {
	return &GenericRouter{name: name, uuid: uuid}
}

func (g *GenericRouter) Uuid() string {
	return g.uuid
}

func (g *GenericRouter) SetUuid(uuid string) {
	g.uuid = uuid
}

func (g *GenericRouter) Name() string {
	return g.name
}

func (g *GenericRouter) SetName(name string) {
	g.name = name
}

func (g *GenericRouter) String() string {
	return g.name + ":" + g.uuid
}
