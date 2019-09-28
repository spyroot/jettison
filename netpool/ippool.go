package netpool

import (
	"fmt"
	"net"
)

/**

 */
type ipPool struct {
	IpAddr string
	inuse  bool
}

/**

 */
type SimpleIpManager struct {
	cidr      string
	pool      []ipPool
	startAddr net.IP
}

func (p *SimpleIpManager) GetPoolCidr() string {
	return p.cidr
}

/**

 */
func (p *SimpleIpManager) Allocate() (string, error) {

	for i, v := range p.pool {
		if v.inuse == false {
			p.pool[i].inuse = true
			return v.IpAddr, nil
		}
	}

	return "", fmt.Errorf("no more free addresses")
}

/**

 */
func (p *SimpleIpManager) SetInUse(ipaddr string) {
	for i, v := range p.pool {
		if v.IpAddr == ipaddr && v.inuse == false {
			p.pool[i].inuse = true
		}
	}
}

/**

 */
func (p *SimpleIpManager) IsInUse(ipaddr string) bool {

	for _, v := range p.pool {
		if v.IpAddr == ipaddr && v.inuse == true {
			return true
		}
	}

	return false
}

/**

 */
func (p *SimpleIpManager) Release(ipaddr string) {

	for i, v := range p.pool {
		if v.IpAddr == ipaddr {
			p.pool[i].inuse = false
		}
	}
}

/**
Build ip pool
*/
func (p *SimpleIpManager) network(cidr string) ([]ipPool, error) {

	ip, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var pool []ipPool
	for ; subnet.Contains(ip); func() {
		for i := len(ip) - 1; i >= 0; i-- {
			ip[i]++
			if ip[i] > 0 {
				break
			}
		}
	}() {
		pool = append(pool, ipPool{ip.String(), false})
	}

	if len(pool) > 2 {
		pool = pool[1 : len(pool)-1]
	}

	return pool, nil
}

/**
Create new ip pool manager
*/
func NewPool(cidr string) (*SimpleIpManager, error) {

	var simple SimpleIpManager
	simple.cidr = cidr

	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	simple.startAddr = ip

	simple.pool, err = simple.network(cidr)
	if err != nil {
		return nil, err
	}

	return &simple, nil
}
