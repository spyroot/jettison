/*
Copyright (c) 2015 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Basic helpers for unit testing.

Author Mustafa Bayramov
mbaraymov@vmware.com
*/
package vcenter

import (
	"context"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"log"
	"os"
)

type Helper struct {
	c   *vim25.Client
	f   *find.Finder
	fns []func()
}

type MinVimConfig struct{}

func (v *MinVimConfig) Endpoint() string {
	return "172.16.254.203"
}

func (v *MinVimConfig) VimUsername() string {
	return "Administrator@vmwarelab.edu"
}
func (v *MinVimConfig) VimPassword() string {
	return "***REMOVED***"
}

func (v *MinVimConfig) Datacenter() string {
	return "Datacenter"
}

type TestingEnv struct {
	ctx           context.Context
	TestImageName string
	TestImageId   string
}

func connect() (*govmomi.Client, context.Context) {

	ctx := context.Background()
	vsphereClient, err := Connect(ctx,
		os.Getenv("VCENTER_HOSTNAME"),
		os.Getenv("VCENTER_USERNAME"),
		os.Getenv("VCENTER_PASSWORD"))

	if err != nil {
		log.Fatal("Failed VimSetupHelper()", err)
	}

	return vsphereClient, ctx
}

/**
  Unit test helper init environment, open up all required connection
  Create single test image in vcenter for a unit tests.
*/
func VimSetupHelper() (*TestingEnv, *govmomi.Client) {

	client, ctx := connect()

	// create test vm
	uuid, err := CreateVmifneeded(ctx, client.Client)
	if err != nil {
		log.Fatal("Failed deploy VM for test")
	}

	return &TestingEnv{ctx, TestImageName, uuid}, client
}

/*
   Testing Helper
*/
func NewHelper(client *vim25.Client) *Helper {

	h := &Helper{
		c:   client,
		fns: make([]func(), 0),
	}

	h.f = find.NewFinder(h.c, true)

	return h
}

func (h *Helper) Defer(fn func()) {
	h.fns = append(h.fns, fn)
}

func (h *Helper) Teardown() {
	for _, fn := range h.fns {
		fn()
	}
}

func (h *Helper) Datacenter() *object.Datacenter {
	dc, err := h.f.DefaultDatacenter(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	h.f.SetDatacenter(dc)

	log.Println("Default datacenter", dc.Name())

	return dc
}

func (h *Helper) DatacenterFolders() *object.DatacenterFolders {
	df, err := h.Datacenter().Folders(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	return df
}

func (h *Helper) ComputeResource(clusterName string) *object.ComputeResource {
	cr, err := h.f.ComputeResource(context.Background(), clusterName)

	//cr, err := h.f.DefaultComputeResource(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Default datacenter", cr.Name())

	return cr
}

func (h *Helper) LocalDatastores(ctx context.Context, cr *object.ComputeResource) ([]*object.Datastore, error) {

	// List datastores for compute resource
	dss, err := cr.Datastores(ctx)
	if err != nil {
		return nil, err
	}

	// Filter local datastores
	var ldss []*object.Datastore
	for _, ds := range dss {
		var mds mo.Datastore
		err = property.DefaultCollector(h.c).RetrieveOne(ctx, ds.Reference(), nil, &mds)
		if err != nil {
			return nil, err
		}

		switch i := mds.Info.(type) {
		case *types.VmfsDatastoreInfo:
			if i.Vmfs.Local != nil && *i.Vmfs.Local == true {
				break
			}
		default:
			continue
		}

		ds.InventoryPath = mds.Name
		ldss = append(ldss, ds)
	}

	return ldss, nil
}
