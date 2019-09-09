package dbutil

import (
	"database/sql"
	"github.com/spyroot/jettison/config"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

var DB *sql.DB

func setupTestCase(t *testing.T) func(t *testing.T) {
	db, err := Connect("/Users/spyroot/go/database/jettison.db")
	if err != nil {
		log.Fatal(err)
	}

	DB = db
	return func(t *testing.T) {
		t.Log("cleanup tables")
		//err := CleanUp(DB)
		//if err != nil {
		//	t.Fatal("Failed drop tables", err)
		//}
		defer DB.Close()
	}
}

func TestCreateDatabase(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    " create database",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := CreateDatabase()
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDatabase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				defer db.Close()
			}
		})
	}
}

//
//func createSyntheticValidNode() *config.NodeTemplate {
//
//	var node0 = config.NodeTemplate{}
//	node0.Name=uuid.New().String()
//	node0.DesiredAddress = "172.16.81.0/24"
//	node0.IPv4Addr = net.ParseIP("172.16.81.1")
//	node0.IPv4Addr = net.ParseIP("172.16.81.1")
//	node0.VimCluster = uuid.New().String()
//	node0.VmTemplateName =  uuid.New().String()
//	node0.Mac = append(node0.Mac,  "abcd:abcd:abc")
//	node0.VimName = uuid.New().String()
//	node0.LogicalSwitchRef = "test-lds"
//	node0.Type = config.ControlType
//
//	return &node0
//}

//
//type NodeTemplate struct {
//	Name				string
//	Prefix         		string `yaml:"prefix"`
//	DomainSuffix   		string `yaml:"domainSuffix"`
//	DesiredCount   		int    `yaml:"desiredCount"`
//	DesiredAddress 		string `yaml:"desiredAddress"`
//	IPv4AddrStr			string `yaml:"IPv4address"`
//	IPv4Addr			net.IP
//	IPv4Net             *net.IPNet
//	Gateway     		string `yaml:"gateway"`
//	VmTemplateName 		string `yaml:"vmTemplateName"`
//	UUID           		string `yaml:"uuid"`
//	VimCluster     		string `yaml:"clusterName"`
//	NetworksRef    		[]string
//	Mac            		[]string
//	VimName        		string
//	LogicalSwitchRef	string	`yaml:"logicalSwitch"`
//	LogicalRouterRef 	string
//	Type				NodeType
//	VimState			Status
//	DhcpStatus     		Status
//	NetworkStatus	 	Status
//	AnsibleStatus		Status
//}

func TestCreateDeployment(t *testing.T) {

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	type args struct {
		db      *sql.DB
		node    *[]*config.NodeTemplate
		depName string
	}

	testNode01 := CreateSyntheticValidNodes(t, 1)

	args01 := args{nil, testNode01, "test"}
	args02 := args{DB, nil, "test"}
	args03 := args{DB, testNode01, ""}
	args04 := args{DB, testNode01, "test"}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "nil db",
			args:    args01,
			wantErr: true,
		},
		{
			name:    "nil node struct",
			args:    args02,
			wantErr: true,
		},
		{
			name:    "empty name",
			args:    args03,
			wantErr: true,
		},
		{
			name:    "basic insert",
			args:    args04,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := func() error {
				createErr := CreateDeployment(tt.args.db, tt.args.node, tt.args.depName)
				if createErr != nil {
					return createErr
				}

				argNodes := *tt.args.node
				nodes, ok, err := GetDeploymentNodes(tt.args.db, tt.args.depName)
				if err != nil {
					return err
				}

				if ok {
					for _, n := range nodes {
						// expect vs actual
						assert.Equal(t, n.Name, argNodes[0].Name, " name mismatch")
						assert.Equal(t, n.IPv4Addr.String(), argNodes[0].IPv4Addr.String(), " ip addr mismatch")
						assert.Equal(t, n.Mac[0], argNodes[0].Mac[0], "mac mismatch")
						assert.Equal(t, n.GetVimName(), argNodes[0].GetVimName(), "vim name mismatch")
						assert.Equal(t, n.GetFolderPath(), argNodes[0].GetFolderPath(), "folder mismatch")
						//assert.Equal(t, n.GetSwitchUuid(), argNodes[0].GetSwitchUuid(), "switch mismatch")
						assert.Equal(t, n.GetDhcpId(), argNodes[0].GetDhcpId(), "dhcp id mismatch")
						assert.Equal(t, n.GetRouterUuid(), argNodes[0].GetRouterUuid(), "router uuid mismatch")
						assert.Equal(t, n.VimCluster, argNodes[0].VimCluster, "cluster name mismatch")
						assert.Equal(t, n.Type, argNodes[0].Type, "type mismatch")
					}
				}

				// delete everything
				return DeleteDeployment(tt.args.db, tt.args.depName)
			}()

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDeployment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
			}

			//if err := CreateDeployment(tt.args.db, tt.args.node, tt.args.depName); (err != nil) != tt.wantErr {
			//	t.Errorf("CreateDeployment() error = %v, wantErr %v", err, tt.wantErr)
			//} else {
			//
			//	log.Println("")
			//	if (err != nil) != tt.wantErr {
			//		t.Errorf("CreateDeployment() error = %v, wantErr %v", err, tt.wantErr)
			//	}
			//	//if tt.wantErr == true && ok == true {
			//	//	t.Errorf("CreateDeployment() error = %v, wantErr %v", err, tt.wantErr)
			//	//}
			//	//
			//	//if tt.wantErr == false && ok == false {
			//	//	t.Errorf("CreateDeployment() error = %v, wantErr %v", err, tt.wantErr)
			//	//}
			//	//
			//	//
			//	//argNodes := *tt.args.node
			//	//
			//	//for _, n := range nodes {
			//	//	if n.Name != argNodes[0].Name {
			//	//		t.Log("Name mismatched")
			//	//	} else {
			//	//		t.Log("Name matched")
			//	//	}
			//	//}
			//}
		})
	}
}

func TestDeleteDeployment(t *testing.T) {

	type args struct {
		db      *sql.DB
		node    *config.NodeTemplate
		depName string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		want1   bool
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DeleteDeployment(tt.args.db, tt.args.depName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteDeployment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGetDeployment(t *testing.T) {
	type args struct {
		db      *sql.DB
		depName string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		want1   bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, _ = GetDeployment(tt.args.db, tt.args.depName)
		})
	}
}

func Test_deleteNodes(t *testing.T) {

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testNode01 := CreateSyntheticValidNodes(t, 1)

	type args struct {
		db      *sql.DB
		depName string
	}

	args01 := args{DB, "test"}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "nil db",
			args:    args01,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// create deployment
			if err := CreateDeployment(DB, testNode01, tt.args.depName); (err != nil) != tt.wantErr {
				t.Errorf("Test_deleteNodes() error = %v, wantErr %v", err, tt.wantErr)
			}

			_, _, ok, err := GetDeployment(tt.args.db, tt.args.depName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Test_deleteNodes() error = %v, wantErr %v", err, tt.wantErr)
			}

			if ok {
				// delete and re-check
				if err := deleteNodes(tt.args.db, tt.args.depName); (err != nil) != tt.wantErr {
					t.Errorf("Test_deleteNodes() error = %v, wantErr %v", err, tt.wantErr)
				}
				_, AfterDelete, ok, err := GetDeployment(tt.args.db, tt.args.depName)
				if (err != nil) != tt.wantErr {
					t.Errorf("Test_deleteNodes() error = %v, wantErr %v", err, tt.wantErr)
				}
				if (ok && AfterDelete == 0) != tt.wantErr {
					t.Errorf("Test_deleteNodes() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
