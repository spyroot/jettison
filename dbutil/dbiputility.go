package dbutil

import (
	"database/sql"
	"fmt"
	"github.com/juju/errors"
	"github.com/spyroot/jettison/jettypes"
	"github.com/spyroot/jettison/logging"
	"log"
	"net"
)

type Node interface {
	// a uuid of a node
	GetUuidName() string

	// a node type
	GetNodeTypeAsString() string
}

type SubnetAllocation interface {
	GetCidr() string
	GetAllocation() string
}

type PodSubnet struct {
	hostuuid  string
	ipblock   string
	cidrblock string
	podtype   jettypes.NodeType
}

func (s *PodSubnet) GetCidr() string {
	if s != nil {
		return s.cidrblock
	}
	return ""
}

func (s *PodSubnet) GetAllocation() string {
	if s != nil {
		return s.ipblock
	}
	return ""
}

func (s *PodSubnet) GetUuidName() string {
	if s != nil {
		return s.hostuuid
	}
	return jettypes.Unknown.String()
}

func (s *PodSubnet) GetNodeType() string {
	if s != nil {
		return s.podtype.String()
	}
	return ""
}

/**
  Function insert current IP block allocated to pod
  typically it is something like /24
*/
func MakeAllocation(db *sql.DB, node Node,
	projectName, ipblock, cidr string) (int64, error) {

	if db == nil {
		return 0, fmt.Errorf("database connector is nil")
	}
	if len(projectName) == 0 {
		return 0, fmt.Errorf("deployment name is empty")
	}
	if node == nil || len(node.GetUuidName()) == 0 {
		return 0, fmt.Errorf("node or node name is empty")
	}
	if len(cidr) == 0 || len(ipblock) == 0 {
		return 0, fmt.Errorf("cifr and ip block can't be empty")
	}
	if _, _, err := net.ParseCIDR(cidr); err != nil {
		err := fmt.Errorf("invalid cidr format")
		logging.ErrorLogging(err)
		return 0, err
	}
	// ip block stored with / mask
	if _, _, err := net.ParseCIDR(ipblock); err != nil {
		err := fmt.Errorf("invalid IP address format")
		logging.ErrorLogging(err)
		return 0, err
	}

	err := CreateTablesIfNeed(db)
	if err != nil {
		return 0, fmt.Errorf("failed create tables")
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, errors.Trace(err)
	}

	query := ` INSERT INTO main.podipblock(id, nodeid, ipblock, cidrblock, type)
                     SELECT deployment.id, nodes.nodeid, ?, ?, 1
               FROM deployment, nodes 
  	             WHERE deployment.DeploymentName = ? AND  nodes.JettisonUuid == ?`

	stmt, err := tx.Prepare(query)
	if err != nil {
		logging.ErrorLogging(err)
		return 0, errors.Trace(err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			log.Println("failed to close db smtm", err)
		}
	}()

	r, err := stmt.Exec(ipblock, cidr, projectName, node.GetUuidName())
	if err != nil {
		tx.Rollback()
		logging.ErrorLogging(err)
		return 0, errors.Trace(err)
	}
	depId, err := r.LastInsertId()
	if err != nil {
		tx.Rollback()
		logging.ErrorLogging(err)
		return 0, errors.Trace(err)
	}

	err = tx.Commit()
	if err != nil {
		logging.ErrorLogging(err)
		return 0, errors.Trace(err)
	}

	return depId, nil
}

/**
  Deletes current IP block allocation for a given node.
*/
func DeleteSubnetAllocation(db *sql.DB, nodeName string) error {

	if db == nil {
		return fmt.Errorf("database connector is nil")
	}

	query := `DELETE FROM podipblock WHERE nodeid = 
				(SELECT nodeid FROM nodes WHERE JettisonUuid is ?)`

	stmt, err := db.Prepare(query)
	if err != nil {
		logging.ErrorLogging(err)
		return errors.Trace(err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			log.Println("failed to close db smtm", err)
		}
	}()

	_, err = stmt.Exec(nodeName)
	if err != nil {
		logging.ErrorLogging(err)
		return errors.Trace(err)
	}

	return nil
}

/**
  Return a generic view about current allocation.
*/
func GetSubnetAllocation(db *sql.DB) ([]SubnetAllocation, bool, error) {

	models := make([]SubnetAllocation, 0)

	if db == nil {
		return nil, false, fmt.Errorf("database connector is nil")
	}

	query := `SELECT nodes.JettisonUuid, podipblock.cidrblock, podipblock.ipblock, podipblock.type
				FROM deployment, nodes, podipblock
			WHERE deployment.id = nodes.id AND podipblock.nodeid = nodes.nodeid;`

	rows, err := db.Query(query)
	if err != nil {
		return nil, false, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.Println("failed to close db smtm", err)
		}
	}()

	for rows.Next() {
		var (
			hostid   = ""
			cidr     = ""
			ipblock  = ""
			podtype  = ""
			nodetype = 0
		)
		err = rows.Scan(&hostid, &cidr, &ipblock, &podtype)
		if err != nil {
			log.Fatal(err)
		}

		log.Println(hostid)
		log.Println(cidr)
		log.Println(ipblock)
		log.Println(nodetype)

		r := &PodSubnet{}
		r.hostuuid = hostid
		r.cidrblock = cidr
		r.ipblock = ipblock
		r.podtype = jettypes.NodeType(nodetype)

		models = append(models, r)
	}

	err = rows.Err()
	if err != nil {
		return nil, false, err
	}

	return models, true, nil
}
