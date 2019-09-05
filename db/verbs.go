package db

import (
	"database/sql"
	"github.com/juju/errors"
	"github.com/spyroot/jettison/config"
	"log"
)

/**
  TODO option for different drivers
*/
func CreateDatabase() (*sql.DB, error) {

	database, err := sql.Open("sqlite3", "./jettison.db")
	if err != nil {
		return nil, errors.Trace(err)
	}

	query := `
		CREATE TABLE IF NOT EXISTS deployment(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		DeploymentName TEXT NOT NULL UNIQUE )`

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
		nodeid       int not null,
		id           int not null constraint nodes_deployment__fk references deployment,
		JettisonUuid TEXT not null,
		VimUuid      TEXT not null,
		VimName      TEXT not null,
		IPv4Addr     TEXT not null,
		MacAddr      TEXT not null,
		Type         TEXT not null
	);

	create unique index nodes_nodeid_uindex
	on nodes (nodeid);`

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
func deleteNodes(db *sql.DB, node *config.NodeTemplate, depName string) (int, bool, error) {

	var id int

	query := `DELETE FROM nodes WHERE id = 
				(SELECT id FROM deployment WHERE DeploymentName = ?`

	stmt, err := db.Prepare(query)
	if err != nil {
		return id, false, errors.Trace(err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			log.Println("failed to close db smtm", err)
		}
	}()

	err = stmt.QueryRow(depName).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		log.Println("no deployment", depName, "found")
		return id, false, nil
	case err != nil:
		return id, false, errors.Trace(err)
	default:
		log.Printf("deployment id %d\n deleted", id)
		return id, true, nil
	}
}

/**

 */
func DeleteDeployment(db *sql.DB, node *config.NodeTemplate, depName string) (int, bool, error) {

	id, ok, err := deleteNodes(db, node, depName)
	if err != nil {
		return id, false, errors.Trace(err)
	}
	if !ok {
		return id, false, err
	}

	query := `DELETE FROM deployment WHERE DeploymentName = 
				(SELECT id FROM deployment WHERE DeploymentName = ?`

	stmt, err := db.Prepare(query)
	if err != nil {
		return id, false, errors.Trace(err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			log.Println("failed to close db smtm", err)
		}
	}()

	err = stmt.QueryRow(depName).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		log.Println("no deployment with name", depName)
		return id, false, nil
	case err != nil:
		return id, false, errors.Trace(err)
	default:
		log.Printf("dep id is %d\n", id)
		return id, true, nil
	}
}

/**

 */
func GetDeployment(db *sql.DB, depName string) (int, bool, error) {

	var id int

	stmt, err := db.Prepare("SELECT id FROM deployment WHERE DeploymentName = ?")
	if err != nil {
		return id, false, errors.Trace(err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			log.Println("failed to close db smtm", err)
		}
	}()

	err = stmt.QueryRow(depName).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		log.Println("no deployment with name", depName)
		return id, false, nil
	case err != nil:
		return id, false, errors.Trace(err)
	default:
		log.Printf("dep id is %d\n", id)
		return id, true, nil
	}
}

/**

 */
func CreateDeployment(db *sql.DB, node *config.NodeTemplate, depName string) error {

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
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

	_, err = stmt.Exec(depName)
	if err != nil {
		log.Fatal(err)
	}

	// insert each node in
	query := `insert into nodes
	(
		id,
		JettisonUuid,
		VimUuid,
		VimName,
		IPv4Addr,
		MacAddr,
		Type
	) VALUE(?,?,?,?,?,?,?);`

	depId, ok, err := GetDeployment(db, depName)
	if err != nil {
		log.Fatal("error ", err)
	}

	if ok {
		stmt, err := tx.Prepare(query)
		if err != nil {
			return errors.Trace(err)
		}
		defer func() {
			if err := stmt.Close(); err != nil {
				log.Println("failed to close db smtm", err)
			}
		}()

		_, err = stmt.Exec(depId, node.Name, node.UUID, node.VimName,
			node.IPv4Addr.String(), node.Mac, node.Type.String())
		if err != nil {
			log.Fatal(err)
		}

		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := tx.Rollback()
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}
