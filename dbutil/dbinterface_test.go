package dbutil

import (
	"database/sql"
	"github.com/spyroot/jettison/jettypes"
	"github.com/spyroot/jettison/vcenter"
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
			name:    "Create database",
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

func TestCreateDeployment(t *testing.T) {

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	type args struct {
		db      *sql.DB
		node    *[]*jettypes.NodeTemplate
		depName string
	}

	testNode01 := vcenter.CreateSyntheticValidNodes(t, 1)

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

						assert.Equal(t, n.GenericSwitch().DhcpUuid(),
							argNodes[0].GenericSwitch().DhcpUuid(), "dhcp id mismatch")

						assert.Equal(t, n.GenericRouter().Uuid(),
							argNodes[0].GenericRouter().Uuid(), "router uuid mismatch")

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
		node    *jettypes.NodeTemplate
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

	testNode01 := vcenter.CreateSyntheticValidNodes(t, 1)

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

func Test_InsertAllocation(t *testing.T) {

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	type args struct {
		db          *sql.DB
		node        *jettypes.NodeTemplate
		projectName string
		ipblock     string
		cidr        string
	}

	testNodes := vcenter.CreateSyntheticValidNodes(t, 1)

	bogus01 := args{nil, (*testNodes)[0], "unit-test", "", ""}
	bogus02 := args{DB, nil, "unit-test", "", ""}
	bogus03 := args{DB, (*testNodes)[0], "", "", ""}
	invalid01 := args{DB, (*testNodes)[0], "unit-test", "", ""}

	// invalid ip valid cidr
	invalid02 := args{DB,
		(*testNodes)[0],
		"unit-test",
		"abcd",
		"172.16.0.0/16"}

	// valid ip invalid cidr
	invalid03 := args{DB,
		(*testNodes)[0],
		"unit-test",
		"172.16.1.0/24", "abcd"}

	// valid ip invalid cidr
	valid01 := args{DB, (*testNodes)[0],
		"unit-test",
		"172.16.1.0/24",
		"172.16.1.0/16"}

	//
	//duplicate := args{DB, testNode01, "unit-test", "", ""}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "nil db",
			args:    bogus01,
			wantErr: true,
		},
		{
			name:    "nil node struct",
			args:    bogus02,
			wantErr: true,
		},
		{
			name:    "empty name",
			args:    bogus03,
			wantErr: true,
		},
		{
			name:    "empty cidr",
			args:    invalid01,
			wantErr: true,
		},
		{
			name:    "empty cidr",
			args:    invalid01,
			wantErr: true,
		},
		{
			name:    "invalid ip block",
			args:    invalid02,
			wantErr: true,
		},
		{
			name:    "invalid cidr block",
			args:    invalid03,
			wantErr: true,
		},
		{
			name:    "valid record block",
			args:    valid01,
			wantErr: false,
		},
	}
	createErr := CreateDeployment(DB, testNodes, "unit-test")
	if createErr != nil {
		log.Fatal("This is baseline test error ", createErr)
	}

	defer DeleteDeployment(DB, "unit-test")

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if _, err := MakeAllocation(tt.args.db,
				tt.args.node, tt.args.projectName,
				tt.args.cidr, tt.args.ipblock); (err != nil) != tt.wantErr {
				t.Errorf("TestAllocation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_GetSubnetAllocation(t *testing.T) {

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	type args struct {
		db          *sql.DB
		node        *jettypes.NodeTemplate
		projectName string
		ipblock     string
		cidr        string
	}

	testNodes := vcenter.CreateSyntheticValidNodes(t, 1)
	// valid ip invalid cidr
	valid01 := args{DB,
		(*testNodes)[0],
		"unit-test",
		"172.16.1.0/24",
		"172.16.1.0/16"}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "valid record block",
			args:    valid01,
			wantErr: false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			if _, err := MakeAllocation(tt.args.db,
				tt.args.node, tt.args.projectName,
				tt.args.cidr, tt.args.ipblock); (err != nil) != tt.wantErr {
				t.Errorf("TestAllocation() error = %v, wantErr %v", err, tt.wantErr)
			}

			ret, _, _ := GetSubnetAllocation(tt.args.db)
			t.Log(len(ret))
			for _, v := range ret {
				t.Log(v.GetAllocation(), " from cidr ", v.GetCidr())
			}
		})
	}
}
