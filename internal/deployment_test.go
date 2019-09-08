package internal

import (
	"fmt"
	"github.com/spyroot/jettison/config"
	"log"
	"reflect"
	"testing"
)

var appConfig, _ = config.ReadConfig()

func TestNewDeployment(t *testing.T) {
	type args struct {
		workersTemplate    *config.NodeTemplate
		controllerTemplate *config.NodeTemplate
		ingressTemplate    *config.NodeTemplate
	}
	tests := []struct {
		name    string
		args    args
		want    *Deployment
		wantErr bool
	}{
		{
			name:    "nil constructor",
			args:    args{nil, nil, nil},
			want:    nil,
			wantErr: true,
		},

		//{
		//	name:"nil constructor",
		//	args: args{
		//		appConfig.GetWorkersTemplate(),
		//		appConfig.GetControllersTemplate(),
		//		appConfig.GetIngresTemplate()},
		//	want:nil,
		//	wantErr:false,
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDeployment(tt.args.workersTemplate, tt.args.controllerTemplate, tt.args.ingressTemplate)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDeployment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDeployment() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func deepCheck(d *Deployment, got *[]*config.NodeTemplate) error {

	for _, v := range *got {
		if v.Type == config.ControlType {
			res, err := d.FindController(v.Name)
			if err != nil {
				return fmt.Errorf("DeployTaskList() expected %v, got %v", v.Name, err)
			}

			if !reflect.DeepEqual(res, v) {
				return fmt.Errorf("DeployTaskList() expected both be a same")
			}
		}

		if v.Type == config.WorkerType {
			res, err := d.FindWorker(v.Name)
			if err != nil {
				return fmt.Errorf("DeployTaskList() expected %v, got %v", v.Name, err)
			}
			if !reflect.DeepEqual(res, v) {
				return fmt.Errorf("DeployTaskList() expected both be a same")
			}
		}

		if v.Type == config.IngressType {
			res, err := d.FindIngress(v.Name)
			if err != nil {
				return fmt.Errorf("DeployTaskList() expected %v, got %v", v.Name, err)
			}
			if !reflect.DeepEqual(res, v) {
				return fmt.Errorf("DeployTaskList() expected both be a same")
			}
		}
	}

	return nil
}

func TestDeployment_DeployTaskList(t *testing.T) {
	type args struct {
		workersTemplate    *config.NodeTemplate
		controllerTemplate *config.NodeTemplate
		ingressTemplate    *config.NodeTemplate
	}
	tests := []struct {
		name      string
		args      args
		want      *Deployment
		wantErr   bool
		wantCount int
	}{
		{
			name: "deployment deep check",
			args: args{
				appConfig.GetWorkersTemplate(),
				appConfig.GetControllersTemplate(),
				appConfig.GetIngresTemplate()},
			wantErr:   false,
			wantCount: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewDeployment(tt.args.workersTemplate, tt.args.controllerTemplate, tt.args.ingressTemplate)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeployTaskList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := d.DeployTaskList()
			if got == nil {
				t.Errorf("DeployTaskList() should be not nil")
				return
			}

			if len(*got) != tt.wantCount {
				t.Errorf("DeployTaskList() wrong len expected %d got %d", len(*got), tt.wantCount)
				return
			}

			// do a deep check
			err = deepCheck(d, got)
			if err != nil {
				t.Errorf("%v", err)
				return
			}

			newName1 := []string{"change1", "change2", "change3"}
			newName2 := []string{"wkr1", "wrk2", "wkr3"}

			for k, v := range newName1 {
				d.Controllers[k].Name = v
				_, err = d.FindController(v)
				if err != nil {
					t.Errorf("DeployTaskList() should be not nil expected to find %v", v)
					return
				}
			}

			for k, v := range newName2 {
				d.Workers[k].Name = v
				_, err := d.FindWorker(v)
				if err != nil {
					t.Errorf("DeployTaskList() should be not nil expected to find %v", v)
					return
				}
			}

			for k, v := range newName2 {
				d.Workers[k].Name = v
				_, err := d.FindController(v)
				if err == nil {
					t.Errorf("DeployTaskList() should be not nil expected to find %v", v)
					return
				}
			}

			d.Workers[0].DesiredCount = 0
			d.Workers[1].DesiredCount = 0
			d.Workers[2].DesiredCount = 0

			d.Controllers[0].DesiredCount = 0
			d.Controllers[1].DesiredCount = 0
			d.Controllers[2].DesiredCount = 0

			err = deepCheck(d, got)
			if err != nil {
				for _, g := range *got {
					log.Println(g.Name)
				}
				t.Errorf("%v", err)
				return
			}
		})
	}
}

func TestDeployment_allocateAddress(t *testing.T) {
	type fields struct {
		Workers      []config.NodeTemplate
		Controllers  []config.NodeTemplate
		Ingress      []config.NodeTemplate
		AddressPools map[string]SimpleIpManager
	}
	type args struct {
		poolName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Deployment{
				Workers:      tt.fields.Workers,
				Controllers:  tt.fields.Controllers,
				Ingress:      tt.fields.Ingress,
				AddressPools: tt.fields.AddressPools,
			}
			got, err := d.allocateAddress(tt.args.poolName)
			if (err != nil) != tt.wantErr {
				t.Errorf("allocateAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("allocateAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeployment_buildPools(t *testing.T) {
	type fields struct {
		Workers      []config.NodeTemplate
		Controllers  []config.NodeTemplate
		Ingress      []config.NodeTemplate
		AddressPools map[string]SimpleIpManager
	}
	type args struct {
		wPool            string
		worksSubnet      string
		cPool            string
		controllerSubnet string
		p                string
		ingressSubnet    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Deployment{
				Workers:      tt.fields.Workers,
				Controllers:  tt.fields.Controllers,
				Ingress:      tt.fields.Ingress,
				AddressPools: tt.fields.AddressPools,
			}
			if err := d.buildPools(tt.args.wPool, tt.args.worksSubnet, tt.args.cPool, tt.args.controllerSubnet, tt.args.p, tt.args.ingressSubnet); (err != nil) != tt.wantErr {
				t.Errorf("buildPools() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
