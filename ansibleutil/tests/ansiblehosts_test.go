package tests

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/spyroot/jettison/ansibleutil"
	"github.com/spyroot/jettison/osutil"
	"github.com/stretchr/testify/assert"
)

var syntheticData1 = `---
all:
  children:
    ProjectTest1:
      children:
        TestGroup:
          hosts:
            test-81d5dc89-1215-4ead-aa17-0ace109edbdd-test.com:
              ansible_connection: ssh
              ansible_host: 172.16.81.110
              ansible_port: 22
              ansible_user: vmware`

var syntheticData2 = `
---
all:
  children:
    NewProject1:
      children:
        TestWorkers:
          hosts:
            test-81d5dc89-1215-4ead-aa17-0ace109edbdd-test.com:
              ansible_connection: ssh
              ansible_host: 172.16.81.1
              ansible_port: 22
              ansible_user: vmware2
            test-889cf594-77a8-44b3-8a64-5d88406973a3-test.com:
              ansible_connection: ssh
              ansible_host: 172.16.81.2
              ansible_port: 22
              ansible_user: vmware2
            test-045306a0-e8da-453f-bece-3af46cb3ff13-test.com:
              ansible_connection: ssh
              ansible_host: 172.16.81.0
              ansible_port: 22
              ansible_user: vmware2
            test-ca26369f-38c8-487f-8bda-2967f3a9e831-test.com:
              ansible_connection: ssh
              ansible_host: 172.16.81.1
              ansible_port: 22
              ansible_user: vmware2
            test-fa071ad0-7660-4249-bd4d-1f28073f358f-test.com:
              ansible_connection: ssh
              ansible_host: 172.16.81.0
              ansible_port: 22
              ansible_user: vmware2`

type testsetup func()

func SyntheticDataOne() string {
	return syntheticData1
}

func SyntheticDataTwo() string {
	return syntheticData2
}

var testFile01 = "/tmp/hosts"

func createTestFile() (string, error) {

	inputFile, err := os.OpenFile(testFile01, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return "", fmt.Errorf("failed create a %s file: %s", testFile01, err)
	}

	defer func() {
		if err := inputFile.Close(); err != nil {
			log.Println("failed to close", testFile01, err)
		}
	}()

	_, err = inputFile.WriteString(SyntheticDataTwo())
	if err != nil {
		return "", err
	}

	return testFile01, nil
}

func createTestFileSingleEntry() (string, error) {

	inputFile, err := os.OpenFile(testFile01, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return "", fmt.Errorf("failed create a %s file: %s", testFile01, err)
	}

	defer func() {
		if err := inputFile.Close(); err != nil {
			log.Println("failed to close", testFile01, err)
		}
	}()

	_, err = inputFile.WriteString(SyntheticDataOne())
	if err != nil {
		return "", err
	}

	return testFile01, nil
}

//
//func TestAnsibleInventoryHosts_AddSlaveHost(t *testing.T) {
//	type fields struct {
//		filePath string
//		All      struct {
//			Cluster map[string]*ansibleutil.AnsibleCluster `yaml:"children"`
//		}
//	}
//	type args struct {
//		ansibleHost *ansibleutil.AnsibleHosts
//		project     string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			a := &ansibleutil.AnsibleInventoryHosts{
//				filePath: tt.fields.filePath,
//				All:      tt.fields.All,
//			}
//			if err := a.AddSlaveHost(tt.args.ansibleHost, tt.args.project); (err != nil) != tt.wantErr {
//				t.Errorf("AddSlaveHost() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//

func TestAnsibleInventoryHosts_DeleteProject(t *testing.T) {

	emptyInventory, _ := ansibleutil.CreateFromInventory("/tmp/test123")

	// create and read existing inventory
	_, err := createTestFile()
	if err != nil {
		log.Panic("Can't'create test data")
	}
	noneEmpty, err := ansibleutil.CreateFromInventory("/tmp")
	if err != nil {
		log.Panic("Can't create inventory from test data", err)
	}

	type args struct {
		search   string
		inventor *ansibleutil.AnsibleInventoryHosts
		setup    testsetup
		cleanup  testsetup
		expected *ansibleutil.AnsibleInventoryHost
	}

	var host = ansibleutil.AnsibleInventoryHost{
		AnsibleConnection: "ssh",
		AnsibleHost:       "172.16.81.1",
		AnsiblePort:       22,
		AnsibleUser:       "vmware2",
	}

	tests := []struct {
		name   string
		args   args
		wantOk bool
	}{
		{
			"delete on empty inventory",
			args{
				"test",
				emptyInventory,
				func() {},
				func() {},
				nil},
			false,
		},
		{
			"delete one",
			args{
				"NewProject1",
				noneEmpty,
				func() {},
				func() {},
				&host},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.args.setup()
			defer tt.args.cleanup()

			// we delete entire project
			deleteOk := tt.args.inventor.DeleteProject(tt.args.search)
			if (deleteOk != false) != tt.wantOk {
				t.Errorf("DeleteSaveHost() ok = %v, wantOk %v", deleteOk, tt.wantOk)
				return
			}
			// if we delete
			if deleteOk {
				_, newOk := tt.args.inventor.FindSaveHost(tt.args.search)
				// must false we delete entry
				assert.Equal(t, newOk, false, "we found result after delete")

				deleteAgain := tt.args.inventor.DeleteSaveHost(tt.args.search)
				assert.Equal(t, deleteAgain, false, "we found result after delete")

				err = tt.args.inventor.WriteToFile()
				if err != nil {
					log.Panic("We can't write to a file")
				}
			}
		})
	}
}

func TestAnsibleInventoryHosts_DeleteSaveHost2(t *testing.T) {

	emptyInventory, _ := ansibleutil.CreateFromInventory("/tmp/test123")

	// create and read existing inventory
	_, err := createTestFileSingleEntry()
	if err != nil {
		log.Panic("Can't'create test data")
	}

	noneEmpty, err := ansibleutil.CreateFromInventory("/tmp")
	if err != nil {
		log.Panic("Can't create inventory from test data", err)
	}

	type args struct {
		search   string
		inventor *ansibleutil.AnsibleInventoryHosts
		setup    testsetup
		cleanup  testsetup
		expected *ansibleutil.AnsibleInventoryHost
	}

	var host = ansibleutil.AnsibleInventoryHost{
		AnsibleConnection: "ssh",
		AnsibleHost:       "172.16.81.1",
		AnsiblePort:       22,
		AnsibleUser:       "vmware2",
	}

	tests := []struct {
		name   string
		args   args
		wantOk bool
	}{
		{
			"delete on empty inventory",
			args{
				"test",
				emptyInventory,
				func() {},
				func() {},
				nil},
			false,
		},
		{
			"delete one",
			args{
				"test-81d5dc89-1215-4ead-aa17-0ace109edbdd-test.com",
				noneEmpty,
				func() {},
				func() {},
				&host},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.args.setup()
			defer tt.args.cleanup()

			_, ok := tt.args.inventor.FindSaveHost(tt.args.search)
			if tt.wantOk == true {
				if ok != true {
					t.Errorf("DeleteSaveHost() ok = %v, wantOk %v but no entry in the map to delete", ok, tt.wantOk)
				}
			}

			deleteOk := tt.args.inventor.DeleteSaveHost(tt.args.search)
			if (deleteOk != false) != tt.wantOk {
				t.Errorf("DeleteSaveHost() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}

			// if we delete
			if deleteOk {
				_, newOk := tt.args.inventor.FindSaveHost(tt.args.search)
				// must false we delete entry
				assert.Equal(t, newOk, false, "we found result after delete")

				deleteAgain := tt.args.inventor.DeleteSaveHost(tt.args.search)
				assert.Equal(t, deleteAgain, false, "we found result after delete")

				err = tt.args.inventor.WriteToFile()
				if err != nil {
					log.Panic("We can't write to a file")
				}

				newInventory, err := ansibleutil.CreateFromInventory("/tmp")
				if err != nil {
					log.Panic("We can't read back")
				}

				_, ok := newInventory.FindSaveHost(tt.args.search)
				assert.Equal(t, ok, false, "we found result after write")
			}
		})
	}
}

func TestAnsibleInventoryHosts_DeleteSaveHost(t *testing.T) {

	emptyInventory, _ := ansibleutil.CreateFromInventory("/tmp/test123")

	// create and read existing inventory
	_, err := createTestFile()
	if err != nil {
		log.Panic("Can't'create test data")
	}
	noneEmpty, err := ansibleutil.CreateFromInventory("/tmp")
	if err != nil {
		log.Panic("Can't create inventory from test data", err)
	}

	//	createTestFileSingleEntry
	//defer os.Remove(file)

	type args struct {
		search   string
		inventor *ansibleutil.AnsibleInventoryHosts
		setup    testsetup
		cleanup  testsetup
		expected *ansibleutil.AnsibleInventoryHost
	}

	var host = ansibleutil.AnsibleInventoryHost{
		AnsibleConnection: "ssh",
		AnsibleHost:       "172.16.81.1",
		AnsiblePort:       22,
		AnsibleUser:       "vmware2",
	}

	tests := []struct {
		name   string
		args   args
		wantOk bool
	}{
		{
			"delete on empty inventory",
			args{
				"test",
				emptyInventory,
				func() {},
				func() {},
				nil},
			false,
		},
		{
			"delete one",
			args{
				"test-81d5dc89-1215-4ead-aa17-0ace109edbdd-test.com",
				noneEmpty,
				func() {},
				func() {},
				&host},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.args.setup()
			defer tt.args.cleanup()

			_, ok := tt.args.inventor.FindSaveHost(tt.args.search)
			if tt.wantOk == true {
				if ok != true {
					t.Errorf("DeleteSaveHost() ok = %v, wantOk %v but no entry in the map to delete", ok, tt.wantOk)
				}
			}

			deleteOk := tt.args.inventor.DeleteSaveHost(tt.args.search)
			if (deleteOk != false) != tt.wantOk {
				t.Errorf("DeleteSaveHost() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}

			// if we delete
			if deleteOk {
				_, newOk := tt.args.inventor.FindSaveHost(tt.args.search)
				// must false we delete entry
				assert.Equal(t, newOk, false, "we found result after delete")

				deleteAgain := tt.args.inventor.DeleteSaveHost(tt.args.search)
				assert.Equal(t, deleteAgain, false, "we found result after delete")

				err = tt.args.inventor.WriteToFile()
				if err != nil {
					log.Panic("We can't write to a file")
				}

				newInventory, err := ansibleutil.CreateFromInventory("/tmp")
				if err != nil {
					log.Panic("We can't read back")
				}

				_, ok := newInventory.FindSaveHost(tt.args.search)
				assert.Equal(t, ok, false, "we found result after delete")
			}
		})
	}
}

func TestAnsibleInventoryHosts_FindSaveHost(t *testing.T) {

	emptyInventory, _ := ansibleutil.CreateFromInventory("/tmp/test123")

	// create and read existing inventory
	_, err := createTestFile()
	if err != nil {
		log.Panic("Can't'create test data")
	}
	noneEmpty, err := ansibleutil.CreateFromInventory("/tmp")
	if err != nil {
		log.Panic("Can't create inventory from test data", err)
	}

	type args struct {
		search   string
		inventor *ansibleutil.AnsibleInventoryHosts
		setup    testsetup
		cleanup  testsetup
		expected *ansibleutil.AnsibleInventoryHost
	}

	var host = ansibleutil.AnsibleInventoryHost{
		AnsibleConnection: "ssh",
		AnsibleHost:       "172.16.81.1",
		AnsiblePort:       22,
		AnsibleUser:       "vmware2",
	}

	tests := []struct {
		name   string
		args   args
		wantOk bool
	}{
		{
			"empty inventory",
			args{
				"test",
				emptyInventory,
				func() {},
				func() {},
				nil},
			false,
		},
		{
			"positive result",
			args{
				"test-81d5dc89-1215-4ead-aa17-0ace109edbdd-test.com",
				noneEmpty,
				func() {},
				func() {},
				&host},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.args.setup()
			defer tt.args.cleanup()

			got, ok := tt.args.inventor.FindSaveHost(tt.args.search)
			if (ok != false) != tt.wantOk {
				t.Errorf("CreateFromInventory() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}

			if ok {
				assert.NotNil(t, got, "return value must not nil")
				assert.Equal(t, got, tt.args.expected)
			}
		})
	}
}

//func TestAnsibleInventoryHosts_WriteToFile(t *testing.T) {
//	type fields struct {
//		filePath string
//		All      struct {
//			Cluster map[string]*ansibleutil.AnsibleCluster `yaml:"children"`
//		}
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			a := &ansibleutil.AnsibleInventoryHosts{
//				filePath: tt.fields.filePath,
//				All:      tt.fields.All,
//			}
//			if err := a.WriteToFile(); (err != nil) != tt.wantErr {
//				t.Errorf("WriteToFile() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}

func TestCreateFromInventory(t *testing.T) {
	type args struct {
		filePath string
		setup    testsetup
		cleanup  testsetup
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"path to a valid dir",
			args{"/tmp/test",
				func() {
					err := os.MkdirAll("/tmp/test", 0700)
					if err != nil {
						log.Panic("Can crete test")
					}
				}, func() {
					err := os.RemoveAll("/tmp/test/")
					if err != nil {
						log.Panic("failed to remove dir")
					}
				}},
			false,
		},
		{
			"path to a not valid dir",
			args{"/tmp/test2",
				func() {}, func() {},
			},
			true,
		},
		{
			"path to a file",
			args{"/tmp/test",
				func() {
					f, err := os.Create("/tmp/test")
					if err != nil {
						log.Panic("Can crete test file")
					}
					_ = f.Close()
					if osutil.CheckIfExist("/tmp/test") == false {
						log.Panic("file not there")
					}

				}, func() {
					err := os.Remove("/tmp/test")
					if err != nil {
						log.Panic("failed to remove file")
					}
				}},
			true,
		},
		{
			"path to a file",
			args{"/tmp/test",
				func() {
					f, err := os.Create("/tmp/test")
					if err != nil {
						log.Panic("Can crete test file")
					}
					_ = f.Close()
					if osutil.CheckIfExist("/tmp/test") == false {
						log.Panic("file not there")
					}

				}, func() {
					err := os.Remove("/tmp/test")
					if err != nil {
						log.Panic("failed to remove file")
					}
				}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.args.setup()
			defer tt.args.cleanup()

			got, err := ansibleutil.CreateFromInventory(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateFromInventory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				assert.Nil(t, got, "Error case return must be nil")
				return
			}

			assert.NotNil(t, got, "return value must not nil")
			//we should be able to add
			if got == nil {
				ansibleHost := ansibleutil.AnsibleHosts{
					Name:     "test01.vmware.lab",
					Hostname: "172.16.81.131",
					Port:     22,
					User:     "vmware",
					Group:    "TestIngress",
				}
				err = got.AddSlaveHost(&ansibleHost, "NewProject")
				if err != nil {
					t.Log("Got error")
				}
			}
		})
	}
}
