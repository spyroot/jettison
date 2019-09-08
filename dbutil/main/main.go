package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/spyroot/jettison/dbutil"
	"github.com/spyroot/jettison/internal"
	"log"
	"os"
	"strings"
)

func CreateDeployments(db *sql.DB, depName string) error {

	nodes := dbutil.CreateSyntheticValidNodes(5)
	err := dbutil.CreateDeployment(db, nodes, depName)
	if err != nil {
		return err
	}

	id, numNodes, ok, err := dbutil.GetDeployment(db, depName)
	if err != nil {
		return err
	}

	if ok {
		log.Println("Deployment created. deployment ", id, " num nodes deployed ", numNodes)
	}

	return nil
}

/* Generic test */
func main() {

	db, err := dbutil.CreateDatabase()
	if err != nil {
		log.Fatal(err)
	}

	var depName string = "test2"
	_, numNodes, ok, err := dbutil.GetDeployment(db, depName)
	if err != nil {
		log.Fatal("error ", err)
	}

	if ok {
		log.Println("Deployment already in the system and num nodes deployed", numNodes)
		fmt.Print("Do you want kill existing deployment ?: (yes/no/show)")

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			// if exist delete and re-create
			if strings.Contains(scanner.Text(), "yes") || scanner.Text()[0] == 'y' {
				err = dbutil.DeleteDeployment(db, depName)
				err = CreateDeployments(db, depName)
				if err != nil {
					log.Fatal(err)
				}
				break
			} else if strings.Contains(scanner.Text(), "show") || scanner.Text()[0] == 's' {
				nodes, ok, err := dbutil.GetDeploymentNodes(db, depName)

				if err != nil {
					log.Fatal(err)
				}

				if ok {
					internal.DebugNodes(nodes)
				}

				return
			} else {
				return
			}
		}

		if scanner.Err() != nil {
			log.Fatal(scanner.Err())
		}
	}

	err = CreateDeployments(db, depName)
	if err != nil {
		log.Fatal(err)
	}
}
