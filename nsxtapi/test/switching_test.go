package test

import (
	"github.com/spyroot/jettison/nsxtapi"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/common"
	"github.com/vmware/go-vmware-nsxt/manager"
	"reflect"
	"testing"
)

func TestCreateLogicalPort(t *testing.T) {

	c, teardown := setupTest()
	defer teardown()

	tag1 := nsxtapi.MakeSwitchTags("testTenant01", "controller")
	switch01, err :=
		nsxtapi.CreateLogicalSwitch(&c, "overlay-trasport-zone", "test-delete", tag1)
	if err != nil {
		t.Fatal("Failed create switch for test")
	}
	defer nsxtapi.DeleteLogicalSwitch(&c, switch01.Id)

	type args struct {
		nsxClient *nsxt.APIClient
		portName  string
		switchId  string
		tags      []common.Tag
	}
	tests := []struct {
		name    string
		args    args
		want    *manager.LogicalPort
		wantErr bool
	}{
		{
			name: "everything is empty",
			args: args{
				&c,
				"",
				"",
				nil,
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "valid switch id the rest is empty",
			args: args{
				&c,
				"",
				switch01.Id,
				nil,
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "wrong switch id",
			args: args{
				&c,
				"test",
				"test",
				nil,
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "valid case",
			args: args{
				&c,
				"test",
				switch01.Id,
				nil,
			},
			wantErr: false,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nsxtapi.CreateLogicalPort(tt.args.nsxClient, tt.args.portName, tt.args.switchId, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Log(got.Id)
				t.Errorf("CreateLogicalPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && got != nil {
				t.Errorf("err is not nill return must be a nil")
				return
			}

			if err == nil && got == nil {
				t.Errorf("err is a nil return must not be a nil")
				return
			}

			if err != nil && got == nil && tt.wantErr == true {
				return
			}

			assert.Nil(t, err, "error must be a nil")
			assert.NotNil(t, got, "got result must not be nil")
			assert.Equal(t, got.LogicalSwitchId, tt.args.switchId)
			assert.ElementsMatch(t, got.Tags, tt.args.tags)

			nsxtapi.DeleteLogicalPort(tt.args.nsxClient, got.Id)
		})
	}
}

//
//// tests find logical switch
//func TestFindLogicalSwitch(t *testing.T) {
//
//	c, teardown := setupTest()
//	defer teardown()
//
//	type args struct {
//		nsxClient       *nsxt.APIClient
//		logicalSwitchId string
//	}
//
//	tests := []struct {
//		name     string
//		args     args
//		want     string
//		wantErr  bool
//	}{
//		// test passing null value for connector
//		{
//			"test1",
//			args{
//				nil,
//				""},
//			"",
//			true,
//		},
//		// test should pass with error since logicalSwitch empty
//		{
//			"test2",
//			args{
//				&c,
//				""},
//			"",
//			true,
//		},
//		// test should pass with error since logicalSwitch empty
//		{
//			"lookup by switch id",
//			args{
//				&c,
//				"91c9a86e-20f2-410f-9561-55e5161c1842",
//			},
//			env.TestVim.JetConfig.GetLogicalSwitch(),
//			false,
//		},
//		// test should pass with error since logicalSwitch empty
//		{
//			"lookup by name",
//			args{
//				nsxClient,
//				env.TestVim.JetConfig.GetLogicalSwitch(),
//			},
//			env.TestVim.JetConfig.GetLogicalSwitch(),
//			false,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := nsxtapi.FindLogicalSwitch(tt.args.nsxClient, tt.args.logicalSwitchId, nil)
//			log.Println(err != nil, tt.wantErr)
//			if err != nil {
//				if tt.wantErr {
//					// if we want error we pass
//					return
//				} else {
//					// if we dont want error but we got one, we failed.
//					t.Errorf("TestFindLogicalSwitch() error = %v, wantErr %v", err, tt.wantErr)
//				}
//				return
//			}
//
//			if got.DisplayName != tt.want {
//				t.Errorf("TestFindLogicalSwitch() got = %v, want %v", got, tt.want)
//			} else {
//				t.Log(got.Id, tt.want)
//			}
//		})
//	}
//}

func TestCreateLogicalSwitch(t *testing.T) {
	type args struct {
		nsxClient  *nsxt.APIClient
		zoneName   string
		switchName string
		tags       []common.Tag
	}
	tests := []struct {
		name    string
		args    args
		want    *manager.LogicalSwitch
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nsxtapi.CreateLogicalSwitch(tt.args.nsxClient, tt.args.zoneName, tt.args.switchName, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLogicalSwitch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateLogicalSwitch() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteLogicalSwitch(t *testing.T) {

	c, teardown := setupTest()
	defer teardown()

	tag1 := nsxtapi.MakeSwitchTags("testTenant01", "controller")
	switch01, err :=
		nsxtapi.CreateLogicalSwitch(&c, "overlay-trasport-zone", "test-delete", tag1)
	if err != nil {
		t.Fatal("Failed create switch for test")
	}

	type args struct {
		nsxClient  *nsxt.APIClient
		switchUuid string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    bool
	}{
		{
			name: "wrong id",
			args: args{
				nsxClient:  &c,
				switchUuid: "wrong",
			},
			wantErr: true,
			want:    false,
		},
		{
			name: "valid delete",
			args: args{
				nsxClient:  &c,
				switchUuid: switch01.Id,
			},
			wantErr: false,
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nsxtapi.DeleteLogicalSwitch(tt.args.nsxClient, tt.args.switchUuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteLogicalSwitch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && got != false {
				t.Errorf("FindLogicalSwitchByTag() return must be false")
				return
			}
			if err == nil && got == false {
				t.Errorf("FindLogicalSwitchByTag() return must be true")
				return
			}
		})
	}
}

func TestFindLogicalPort(t *testing.T) {

	c, teardown := setupTest()
	defer teardown()

	tag1 := nsxtapi.MakeSwitchTags("testTenant03", "findports")
	switch01, err := nsxtapi.CreateLogicalSwitch(&c, "overlay-trasport-zone", "test-find-port", tag1)
	assert.Nil(t, err)
	assert.NotNil(t, switch01)

	var portTag01 = []common.Tag{
		{
			Scope: "jettison-tenant",
			Tag:   "test",
		},
	}
	var portTag02 = []common.Tag{
		{
			Scope: "jettison-tenant",
			Tag:   "test",
		},
	}
	var portTag03 = []common.Tag{
		{
			Scope: "jettison-tenant",
			Tag:   "test",
		},
		{
			Scope: "client-port",
			Tag:   "test123",
		},
	}

	port01, err := nsxtapi.CreateLogicalPort(&c, "port01", switch01.Id, portTag01)
	assert.Nil(t, err)
	assert.NotNil(t, port01)
	defer nsxtapi.DeleteLogicalPort(&c, port01.Id)

	port02, err := nsxtapi.CreateLogicalPort(&c, "port02", switch01.Id, portTag02)
	assert.Nil(t, err)
	assert.NotNil(t, port02)
	defer nsxtapi.DeleteLogicalPort(&c, port02.Id)

	port03, err := nsxtapi.CreateLogicalPort(&c, "port03", switch01.Id, portTag03)
	assert.Nil(t, err)
	assert.NotNil(t, port02)
	nsxtapi.DeleteLogicalPort(&c, port03.Id)

	nsxtapi.DeleteLogicalSwitch(&c, switch01.Id)

	type args struct {
		nsxClient *nsxt.APIClient
		switchId  string
		tags      []common.Tag
	}

	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
		wantLen int
	}{
		{
			name: "empty port",
			args: args{
				nsxClient: &c,
			},
			wantErr: true,
		},
		{
			name: "empty tags",
			args: args{
				nsxClient: &c,
				switchId:  switch01.Id,
			},
			wantErr: true,
		},
		{
			name: "empty tags",
			args: args{
				nsxClient: &c,
				switchId:  switch01.Id,
				tags:      portTag01,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nsxtapi.FindLogicalPort(tt.args.nsxClient, tt.args.switchId, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindLogicalPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if _, ok := err.(*nsxtapi.ObjectNotFound); ok {
				assert.Nil(t, got)
				return
			}
			if err != nil {
				assert.Nil(t, got)
			}

			if err == nil {
				assert.Equal(t, len(got), tt.wantLen)
			}
		})
	}
}

func TestFindLogicalSwitchByTag(t *testing.T) {

	c, teardown := setupTest()
	defer teardown()

	tag1 := nsxtapi.MakeSwitchTags("testTenant01", "controller")
	tag2 := nsxtapi.MakeSwitchTags("testTenant01", "worker")
	tag3 := nsxtapi.MakeSwitchTags("testTenant02", "controller")
	tag4 := nsxtapi.MakeSwitchTags("testTenant02", "worker")
	tag5 := nsxtapi.MakeSwitchTags("testTenant03", "ingress")

	var emptyTag []common.Tag

	switch01, err :=
		nsxtapi.CreateLogicalSwitch(&c, "overlay-trasport-zone", "test01", tag1)
	if err != nil {
		t.Fatal("Failed create switch for test")
	}
	defer nsxtapi.DeleteLogicalSwitch(&c, switch01.Id)

	switch02, err := nsxtapi.CreateLogicalSwitch(&c, "overlay-trasport-zone", "test01", tag2)
	if err != nil {
		t.Fatal("Failed create switch for test")
	}
	defer nsxtapi.DeleteLogicalSwitch(&c, switch02.Id)

	switch03, err :=
		nsxtapi.CreateLogicalSwitch(&c, "overlay-trasport-zone", "test01", tag3)
	if err != nil {
		t.Fatal("Failed create switch for test")
	}
	defer nsxtapi.DeleteLogicalSwitch(&c, switch03.Id)

	switch04, err := nsxtapi.CreateLogicalSwitch(&c, "overlay-trasport-zone", "test01", tag4)
	if err != nil {
		t.Fatal("Failed create switch for test")
	}
	defer nsxtapi.DeleteLogicalSwitch(&c, switch04.Id)

	switch05, err := nsxtapi.CreateLogicalSwitch(&c, "overlay-trasport-zone", "test01", tag5)
	if err != nil {
		t.Fatal("Failed create switch for test")
	}
	defer nsxtapi.DeleteLogicalSwitch(&c, switch05.Id)

	type args struct {
		nsxClient *nsxt.APIClient
		tags      []common.Tag
	}
	tests := []struct {
		name    string
		args    args
		want    string
		tags    []common.Tag
		wantErr bool
	}{
		{
			name: "find switch for tag1",
			args: args{
				nsxClient: &c,
				tags:      tag1,
			},
			want:    switch01.Id,
			wantErr: false,
			tags:    tag1,
		},
		{
			name: "find switch for tag2",
			args: args{
				nsxClient: &c,
				tags:      tag2,
			},
			want:    switch02.Id,
			wantErr: false,
			tags:    tag2,
		},
		{
			name: "nil tag",
			args: args{
				nsxClient: &c,
				tags:      emptyTag,
			},
			want:    switch01.Id,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nsxtapi.FindLogicalSwitchByTag(tt.args.nsxClient, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindLogicalSwitchByTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// if error return must be a nil
			if err != nil && got != nil {
				t.Errorf("FindLogicalSwitchByTag() return must be nil")
				return
			}

			if got == nil {
				t.Errorf("FindLogicalSwitchByTag() error is not nil got must not nil")
				return
			}

			if !reflect.DeepEqual(got.Id, tt.want) {
				t.Errorf("FindLogicalSwitchByTag() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got.Tags, tt.tags) {
				t.Errorf("FindLogicalSwitchByTag() got = %v, want %v", got, tt.want)
			}
		})
	}
}
