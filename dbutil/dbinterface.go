package dbutil

import (
	"database/sql"
	"fmt"
	"github.com/juju/errors"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net"

	"github.com/spyroot/jettison/jettypes"
	"github.com/spyroot/jettison/logging"
)

func CreateTablesIfNeed(db *sql.DB) error {

	query := `
		CREATE TABLE IF NOT EXISTS deployment(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		DeploymentName TEXT NOT NULL UNIQUE )`

	statement, err := db.Prepare(query)
	if err != nil {
		logging.ErrorLogging(err)
		return errors.Trace(err)
	}

	_, err = statement.Exec()
	if err != nil {
		logging.ErrorLogging(err)
		return errors.Trace(err)
	}

	query = `CREATE TABLE IF NOT EXISTS nodes
	(
		nodeid       INTEGER PRIMARY KEY AUTOINCREMENT,
		id           INTEGER not null constraint nodes_deployment__fk references deployment,
		JettisonUuid TEXT not null,
		VimUuid      TEXT not null,
		VimName      TEXT not null,
		IPv4Addr     TEXT not null,
		MacAddr      TEXT not null,
		VimFolder    TEXT not null,			
		SwitchUuid   TEXT not null,
		RouterUuid   TEXT not null,
		ClusterName  TEXT not null,
		DhcpUuid     TEXT not null,
		Type         TEXT not null
	);`

	// TODO add createifneed everwhere.

	statement, err = db.Prepare(query)
	if err != nil {
		logging.ErrorLogging(err)
		return errors.Trace(err)
	}

	_, err = statement.Exec()
	if err != nil {
		logging.ErrorLogging(err)
		return errors.Trace(err)
	}

	query = `CREATE TABLE IF NOT EXISTS podipblock
	(
		cidrid       INTEGER PRIMARY KEY AUTOINCREMENT,
		id           INTEGER not null constraint nodes_deployment__fk references deployment,
		nodeid       INTEGER not null constraint podblocks_nodes__fk references nodes,
		ipblock      TEXT not null,
		cidrblock    TEXT not null,
		type		 INTEGER not null
	)`

	statement, err = db.Prepare(query)
	if err != nil {
		logging.ErrorLogging(err)
		return errors.Trace(err)
	}

	_, err = statement.Exec()
	if err != nil {
		logging.ErrorLogging(err)
		return errors.Trace(err)
	}

	return nil
}

// clean all nodes
func CleanNodes(db *sql.DB) error {

	query := `DELETE FROM main.nodes`

	stmt, err := db.Prepare(query)
	if err != nil {
		log.Println(err)
		return errors.Trace(err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			log.Println("failed to close stmt", err)
		}
	}()

	_, err = stmt.Exec()
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

// clean all data in deployment table
func CleanDeployment(db *sql.DB) error {

	query := `DELETE FROM main.deployment`

	stmt, err := db.Prepare(query)
	if err != nil {
		log.Println(err)
		return errors.Trace(err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			log.Println("failed to close stmt", err)
		}
	}()

	_, err = stmt.Exec()
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func CleanUp(db *sql.DB) error {

	err := CleanNodes(db)
	if err != nil {
		return err
	}
	err = CleanDeployment(db)
	if err != nil {
		return err
	}

	return nil
}

// plain connect, caller need validate that path to database is valid
func Connect(dbpath string) (*sql.DB, error) {

	database, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return database, nil
}

/**
  TODO option for different drivers
*/
func CreateDatabase() (*sql.DB, error) {

	database, err := sql.Open("sqlite3", "/Users/spyroot/go/database/jettison.db")
	if err != nil {
		return nil, errors.Trace(err)
	}

	query := `CREATE TABLE IF NOT EXISTS deployment
		(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		DeploymentName TEXT NOT NULL UNIQUE 
		)`

	statement, err := database.Prepare(query)
	if err != nil {
		return nil, errors.Trace(err)
	}

	_, err = statement.Exec()
	if err != nil {
		return nil, errors.Trace(err)
	}

	query = `CREATE TABLE IF NOT EXISTS nodes
	(
		nodeid       INTEGER PRIMARY KEY AUTOINCREMENT,
		id           INTEGER not null constraint nodes_deployment__fk references deployment,
		JettisonUuid TEXT not null,
		VimUuid      TEXT not null,
		VimName      TEXT not null,
		IPv4Addr     TEXT not null,
		MacAddr      TEXT not null,
		VimFolder    TEXT not null,			
		SwitchUuid   TEXT not null,
		RouterUuid   TEXT not null,
		ClusterName  TEXT not null,
		DhcpUuid     TEXT not null,
		Type         TEXT not null
	);`

	statement, err = database.Prepare(query)
	if err != nil {
		return nil, errors.Trace(err)
	}

	_, err = statement.Exec()
	if err != nil {
		return nil, errors.Trace(err)
	}

	return database, nil
}

/**

 */
func deleteNodes(db *sql.DB, depName string) error {

	if db == nil {
		return fmt.Errorf("database connector is nil")
	}

	query := `DELETE FROM nodes WHERE id = 
				(SELECT id FROM deployment WHERE DeploymentName is ?)`

	stmt, err := db.Prepare(query)
	if err != nil {
		log.Println("db.Prepare error", err)
		return errors.Trace(err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			log.Println("failed to close db smtm", err)
		}
	}()

	_, err = stmt.Exec(depName)
	if err != nil {
		log.Println("Exec error", err)
		return errors.Trace(err)
	}

	return nil
}

/**
  Function delete deployment from database.
*/
func DeleteDeployment(db *sql.DB, projectName string) error {

	if db == nil {
		return fmt.Errorf("database connector is nil")
	}

	err := deleteNodes(db, projectName)
	if err != nil {
		logging.ErrorLogging(err)
		return errors.Trace(err)
	}

	// delete all nodes from deployment
	query := `DELETE FROM nodes WHERE id = (SELECT id FROM deployment WHERE DeploymentName = ?)`

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

	_, err = stmt.Exec(projectName)
	if err != nil {
		logging.ErrorLogging(err)
		return errors.Trace(err)
	}

	// delete all nodes from deployment
	query = `DELETE FROM podipblock WHERE id = (SELECT id FROM deployment WHERE DeploymentName = ?)`

	stmt, err = db.Prepare(query)
	if err != nil {
		logging.ErrorLogging(err)
		return errors.Trace(err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			log.Println("failed to close db smtm", err)
		}
	}()

	_, err = stmt.Exec(projectName)
	if err != nil {
		logging.ErrorLogging(err)
		return errors.Trace(err)
	}

	query = `DELETE FROM deployment WHERE DeploymentName = ?`

	stmt, err = db.Prepare(query)
	if err != nil {
		logging.ErrorLogging(err)
		return errors.Trace(err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			log.Println("failed to close db smtm", err)
		}
	}()

	_, err = stmt.Exec(projectName)
	if err != nil {
		logging.ErrorLogging(err)
		return errors.Trace(err)
	}
	return nil
}

type TemplateNode interface {
	Name() string
	MacAddress() string
	IPv4Address() string
	SwitchUuid() string
	RouterUuid() string
	DhcpServerUuid() string
	GetNodeType() string
	ClusterName() string
	VimFolder() string
}

/**

 */
func GetDeploymentNodes(db *sql.DB, depName string) ([]*jettypes.NodeTemplate, bool, error) {

	var nodes []*jettypes.NodeTemplate

	if db == nil {
		return nodes, false, fmt.Errorf("database connector is nil")
	}

	if depName == "" {
		return nodes, false, fmt.Errorf("empty deployment name")
	}

	query := `SELECT JettisonUuid, VimUuid, VimName, IPv4Addr, MacAddr, 
				VimFolder, SwitchUuid, RouterUuid, ClusterName, DhcpUuid, Type FROM nodes 
				WHERE id = (SELECT id FROM deployment WHERE DeploymentName is ?)`

	rows, err := db.Query(query, depName)
	if err != nil {
		return nodes, false, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.Println("failed to close db smtm", err)
		}
	}()

	for rows.Next() {
		var (
			node        = &jettypes.NodeTemplate{}
			mac         = ""
			nodeType    = ""
			vimName     = ""
			vimFolder   = ""
			switchUuid  = ""
			routerUuid  = ""
			clusterName = ""
			dhcpId      = ""
		)
		err = rows.Scan(&node.Name, &node.UUID, &vimName,
			&node.IPv4AddrStr, &mac, &vimFolder,
			&switchUuid, &routerUuid, &clusterName, &dhcpId, &nodeType)
		if err != nil {
			log.Fatal(err)
		}

		// add mac
		node.Mac = append(node.Mac, mac)
		node.Type = jettypes.GetNodeType(nodeType)
		node.IPv4Addr = net.ParseIP(node.IPv4AddrStr)
		node.SetVimName(vimName)
		node.SetFolderPath(vimFolder)

		// TODO add names
		lrSwitch := jettypes.NewGenericSwitch("", switchUuid, dhcpId, "")
		lrRouter := jettypes.NewGenericRouter("", routerUuid)

		node.SetGenericSwitch(lrSwitch)
		node.SetGenericRouter(lrRouter)
		node.VimCluster = clusterName

		nodes = append(nodes, node)
	}

	err = rows.Err()
	if err != nil {
		return nodes, false, err
	}

	return nodes, true, nil
}

/**
  Function return existing deployment stored in database and number of nodes.
  bool flag set to a true when
*/
func GetDeployment(db *sql.DB, depName string) (int, int, bool, error) {

	var id int
	var numNodes int

	if db == nil {
		return id, numNodes, false, fmt.Errorf("database connector is nil")
	}

	if depName == "" {
		return id, numNodes, false, fmt.Errorf("empty deployment name")
	}

	query := `SELECT deployment.id, count(*) AS numNodes
		FROM deployment, nodes WHERE DeploymentName is ?
    AND deployment.id = nodes.id
    GROUP BY deployment.id`

	stmt, err := db.Prepare(query)
	if err != nil {
		return id, numNodes, false, errors.Trace(err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			log.Println("failed to close db smtm", err)
		}
	}()

	err = stmt.QueryRow(depName).Scan(&id, &numNodes)
	switch {
	case err == sql.ErrNoRows:
		return id, numNodes, false, nil
	case err != nil:
		return id, numNodes, false, errors.Trace(err)
	default:
		return id, numNodes, true, nil
	}
}

/**

 */
func AddNode(db *sql.DB, node *jettypes.NodeTemplate, depId int) error {

	if db == nil {
		return fmt.Errorf("database connector is nil")
	}

	if node == nil {
		return fmt.Errorf("node template is nil")
	}

	err := CreateTablesIfNeed(db)
	if err != nil {
		return fmt.Errorf("failed create tables")
	}

	tx, err := db.Begin()
	if err != nil {
		return errors.Trace(err)
	}

	query := `insert into nodes
       (
		id,
		JettisonUuid,
		VimUuid,
		VimName,
		IPv4Addr,
		MacAddr,
		VimFolder,			
		SwitchUuid,
		RouterUuid,
		ClusterName,
		DhcpUuid,
		Type
  	) VALUES(?,?,?,?,?,?,?,?,?,?,?,?);`

	stmt2, err := tx.Prepare(query)
	if err != nil {
		_ = tx.Rollback()
		return errors.Trace(err)
	}

	defer func() {
		if err := stmt2.Close(); err != nil {
			log.Println("failed to close db smtm", err)
		}
	}()

	// insert each node to a table
	_, err = stmt2.Exec(depId,
		node.Name,
		node.UUID,
		node.GetVimName(),
		node.IPv4Addr.String(),
		node.Mac[0],
		node.GetFolderPath(),
		node.GenericSwitch().Uuid(),
		node.GenericRouter().Uuid(),
		node.VimCluster,
		node.GenericSwitch().DhcpUuid(),
		node.Type.String(),
	)
	if err != nil {
		_ = tx.Rollback()
		return errors.Trace(err)
	}

	// if ok commit
	err = tx.Commit()
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func createDeployment(db *sql.DB, node *jettypes.NodeTemplate, depName string) error {
	var nodes []*jettypes.NodeTemplate
	nodes = append(nodes, node)
	return CreateDeployment(db, nodes, depName)
}

/**
  Function take slice of nodes and deployment name and create respected
  record in database.

  Nodes must hold at least one node
*/
func CreateDeployment(db *sql.DB, nodes []*jettypes.NodeTemplate, depName string) error {

	if db == nil {
		return fmt.Errorf("database connector is nil")
	}

	if nodes == nil || len(nodes) == 0 {
		return fmt.Errorf("node template is nil or empty")
	}

	if len(depName) == 0 {
		return fmt.Errorf("deployment name is empty")
	}

	err := CreateTablesIfNeed(db)
	if err != nil {
		return fmt.Errorf("failed create tables")
	}

	tx, err := db.Begin()
	if err != nil {
		return errors.Trace(err)
	}

	stmt, err := tx.Prepare("INSERT INTO Deployment(DeploymentName) values(?)")
	if err != nil {
		return errors.Trace(err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			log.Println("failed to close db smtm", err)
		}
	}()

	r, err := stmt.Exec(depName)
	if err != nil {
		return errors.Trace(err)
	}
	depId, err := r.LastInsertId()
	if err != nil {
		return errors.Trace(err)
	}

	query := `insert into nodes
       (
		id,
		JettisonUuid,
		VimUuid,
		VimName,
		IPv4Addr,
		MacAddr,
		VimFolder,			
		SwitchUuid,
		RouterUuid,
		ClusterName,
		DhcpUuid,
		Type
  	) VALUES(?,?,?,?,?,?,?,?,?,?,?,?);`

	stmt2, err := tx.Prepare(query)
	if err != nil {
		_ = tx.Rollback()
		return errors.Trace(err)
	}

	defer func() {
		if err := stmt2.Close(); err != nil {
			log.Println("failed to close db smtm", err)
		}
	}()

	for _, node := range nodes {
		// insert each node to a table
		r, err = stmt2.Exec(depId, // 1
			node.Name,                       // 2
			node.UUID,                       // 3
			node.GetVimName(),               // 4
			node.IPv4Addr.String(),          // 5
			node.Mac[0],                     // 6
			node.GetFolderPath(),            // 7
			node.GenericSwitch().Uuid(),     // 8
			node.GenericRouter().Uuid(),     // 9
			node.VimCluster,                 // 10
			node.GenericSwitch().DhcpUuid(), // 11
			node.Type.String())              // 12
		if err != nil {
			_ = tx.Rollback()
			return errors.Trace(err)
		}
	}

	// if ok commit
	err = tx.Commit()
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}
