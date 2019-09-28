package netpool

import (
	"fmt"
	"log"
	"net"
)

/**

 */
type subnetPool struct {
	IpAddr net.IP
	inuse  bool
}

/**

 */
type SimpleSubnetPool struct {
	cidr      string
	startAddr net.IP
	current   net.IP
	subnet    *net.IPNet

	// in case if we need pre-allocate
	subnets      []subnetPool
	allocateSize uint
}

func NextIP(ip net.IP, inc uint) net.IP {
	i := ip.To4()

	// extract each octet
	v := uint(i[0])<<24 + uint(i[1])<<16 + uint(i[2])<<8 + uint(i[3])

	// increment
	v += inc

	// put back
	v3 := byte(v & 0xFF)
	v2 := byte((v >> 8) & 0xFF)
	v1 := byte((v >> 16) & 0xFF)
	v0 := byte((v >> 24) & 0xFF)

	return net.IPv4(v0, v1, v2, v3)
}

func (p *SimpleSubnetPool) nextSubnets(pool string, size uint) {

	offset := uint(0xFFFFFFFF>>size + 1)
	//	ip, subnet, _ := net.ParseCIDR(pool)
	addr := p.current
	for index := 0; p.subnet.Contains(addr) != false; index++ {
		next := NextIP(addr, offset)
		if p.subnet.Contains(next) {
			log.Println(next)
		}
		addr = next
	}
}

/**

 */
func (p *SimpleSubnetPool) AllocateSubnet() (net.IP, error) {

	offset := uint(0xFFFFFFFF>>p.allocateSize + 1)

	if p.current == nil {

		t := net.IPNet{
			IP:   net.ParseIP(p.startAddr.String()),
			Mask: net.CIDRMask(int(p.allocateSize), 32),
		}

		if p.subnet.Contains(t.IP) {
			p.current = t.IP
			// add to inuse list
			p.subnets = append(p.subnets, subnetPool{p.current, true})
			return p.current, nil
		}
	}

	addr := p.current
	for index := 0; p.subnet.Contains(addr) != false; index++ {
		next := NextIP(addr, offset)
		if p.subnet.Contains(next) {
			p.current = next
			// add to in use list
			p.subnets = append(p.subnets, subnetPool{next, true})
			return next, nil
		}
		addr = next
	}

	return nil, fmt.Errorf("no more block left in the CIDR")
}

// TODO Fix me
func (p *SimpleSubnetPool) generateSubnets(size uint) {

	offset := uint(0xFFFFFFFF>>size + 1)
	//	ip, subnet, _ := net.ParseCIDR(pool)
	addr := p.current
	for index := 0; p.subnet.Contains(addr) != false; index++ {
		next := NextIP(addr, offset)
		if p.subnet.Contains(next) {
			p.subnets = append(p.subnets, subnetPool{next, false})
			log.Println(next)
		}
		addr = next
	}
}

/**
  Create new ip subnet manager. It provide simple naive allocation strategy.
  Caller indicate valid range as cidr block and indicate block size each allocation
  need provide.

  For example 10.1.1.0/16 blockSize 24
  upon a first call
  10.1.1.0
  second block
  10.1.2.0
   etc
*/
func NewSubnetPool(cidr string, blockSize uint) (*SimpleSubnetPool, error) {

	var simple SimpleSubnetPool
	simple.cidr = cidr

	cidrIP, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	cidrLen, v := subnet.Mask.Size()
	log.Println(cidrLen, v)

	if int(blockSize) <= cidrLen {
		return nil, fmt.Errorf("the block size must be large than CIDR mask")
	}

	if blockSize > 32 {
		return nil, fmt.Errorf("the block size can't be large than 32 bit")
	}

	simple.subnet = subnet
	simple.allocateSize = blockSize
	simple.startAddr = cidrIP
	simple.current = nil

	return &simple, nil
}
