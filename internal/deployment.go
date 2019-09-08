package internal

import (
	"fmt"
	"github.com/spyroot/jettison/config"
	"github.com/vmware/govmomi/vim25/types"
	"log"
	"net"
)

type TaskMessage struct {
	VmName string
	Status types.TaskInfoState
	Err    error
}

type DeployTaskList struct {
	TaskList *[]config.NodeTemplate
}

/*

 */
type Deployment struct {

	// deployment name
	DeploymentName string

	// deployment name
	ClusterName string

	// list of workers node
	Workers []config.NodeTemplate

	// list of ingress controllers
	Ingress []config.NodeTemplate

	// list of controller nodes
	Controllers []config.NodeTemplate

	AddressPools map[string]SimpleIpManager
}

/**

 */
func (d *Deployment) DeployTaskList() (*[]*config.NodeTemplate, error) {

	tmpSlice := make([]*config.NodeTemplate, 0)

	if len(d.Controllers) < 1 {
		return nil, fmt.Errorf("less than minimum number of controllers")
	}

	if len(d.Workers) < 1 {
		return nil, fmt.Errorf("less than minimum number of controllers")
	}

	if len(d.Ingress) < 1 {
		return nil, fmt.Errorf("less than minimum number of controllers")
	}

	for i := 0; i < len(d.Controllers); i++ {
		tmpSlice = append(tmpSlice, &(d.Controllers[i]))
	}

	for i := 0; i < len(d.Workers); i++ {
		tmpSlice = append(tmpSlice, &(d.Workers[i]))
	}

	for i := 0; i < len(d.Ingress); i++ {
		tmpSlice = append(tmpSlice, &(d.Ingress[i]))
	}

	return &tmpSlice, nil
}

func (d *Deployment) FindController(name string) (*config.NodeTemplate, error) {

	for _, v := range d.Controllers {
		if v.Name == name {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("controller node not found")
}

func (d *Deployment) FindWorker(name string) (*config.NodeTemplate, error) {

	for _, v := range d.Workers {
		if v.Name == name {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("workers node not found")
}

func (d *Deployment) FindIngress(name string) (*config.NodeTemplate, error) {

	for _, v := range d.Ingress {
		if v.Name == name {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("ingress node not found")
}

func (d *Deployment) allocateAddress(poolName string) (string, error) {

	// get the pool
	pool, allocated := d.AddressPools[poolName]
	if !allocated {
		return "", fmt.Errorf("ip pool not found")
	}

	ipAddr, err := pool.Allocate()
	if err != nil {
		return "", fmt.Errorf("no free IP address")
	}

	return ipAddr, nil
}

func (d *Deployment) buildPools(wPool string, worksSubnet string,
	cPool string, controllerSubnet string, p string, ingressSubnet string) error {

	_, ok := d.AddressPools[worksSubnet]
	if !ok {
		newPool, err := NewPool(wPool)
		if err != nil {
			return err
		}
		log.Println("Creating pool for works", wPool)
		d.AddressPools[worksSubnet] = *newPool
	}

	_, ok = d.AddressPools[controllerSubnet]
	if !ok {
		controllersPool, err := NewPool(cPool)
		if err != nil {
			return err
		}
		log.Println("Creating pool for controller", cPool)
		d.AddressPools[controllerSubnet] = *controllersPool
	}

	_, ok = d.AddressPools[ingressSubnet]
	if !ok {
		ingressPool, err := NewPool(p)
		if err != nil {
			return err
		}
		log.Println("Creating pool for ingress", p)
		d.AddressPools[ingressSubnet] = *ingressPool
	}

	return nil
}

/**
  Function create new ip pool manager
*/
func NewDeployment(workersTemplate *config.NodeTemplate, controllerTemplate *config.NodeTemplate,
	ingressTemplate *config.NodeTemplate, depName string) (*Deployment, error) {

	if workersTemplate == nil || controllerTemplate == nil || ingressTemplate == nil || depName == "" {
		return nil, fmt.Errorf("nil argument")
	}

	var d Deployment
	d.DeploymentName = depName

	d.AddressPools = make(map[string]SimpleIpManager)
	err := d.buildPools(workersTemplate.DesiredAddress, workersTemplate.IPv4Net.String(),
		controllerTemplate.DesiredAddress, controllerTemplate.IPv4Net.String(),
		ingressTemplate.DesiredAddress, ingressTemplate.IPv4Net.String())
	if err != nil {
		return nil, fmt.Errorf("failed initilize ip pools %s", err)
	}

	ingPool, ok := d.AddressPools[ingressTemplate.IPv4Net.String()]
	if ok {
		log.Println("Ingress controller uses same pool as controller/workers", ingressTemplate.IPv4Net.String())
		ingPool.SetInUse(ingressTemplate.IPv4Addr.String())
	}

	newNode := ingressTemplate.Clone()
	newNode.GenerateName()

	newNode.IPv4Addr = ingressTemplate.IPv4Addr
	newNode.IPv4AddrStr = ingressTemplate.IPv4Addr.String()
	newNode.Type = config.IngressType

	d.Ingress = append(d.Ingress, *newNode)

	log.Println("Creating vm profile for controller nodes:", controllerTemplate.DesiredCount)
	for i := 0; i < controllerTemplate.DesiredCount; i++ {
		newNode := controllerTemplate.Clone()
		newNode.GenerateName()
		ipAddr, err := d.allocateAddress(controllerTemplate.IPv4Net.String())

		// allocate address
		if err != nil {
			return nil, fmt.Errorf("failed allocate address for controller")
		}
		newNode.IPv4Addr = net.ParseIP(ipAddr)
		newNode.IPv4AddrStr = ipAddr
		newNode.Type = config.ControlType

		// add to a list
		d.Controllers = append(d.Controllers, *newNode)
	}

	log.Println("Creating vm profile for workers node, desired vm count:", workersTemplate.DesiredCount)
	for i := 0; i < workersTemplate.DesiredCount; i++ {

		// clone from template
		newNode := workersTemplate.Clone()
		newNode.GenerateName()

		// allocate address
		ipAddr, err := d.allocateAddress(workersTemplate.IPv4Net.String())
		if err != nil {
			return nil, fmt.Errorf("failed allocate address for worker")
		}
		newNode.IPv4Addr = net.ParseIP(ipAddr)
		newNode.IPv4AddrStr = ipAddr
		newNode.Type = config.WorkerType
		// add to a list
		d.Workers = append(d.Workers, *newNode)
	}

	return &d, nil
}
