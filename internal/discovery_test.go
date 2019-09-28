package internal

import (
	"github.com/spyroot/jettison/nsxtapi"
	"github.com/spyroot/jettison/plugins"
	"github.com/spyroot/jettison/vcenter"
	"log"
	"testing"
)

func setupTest(t *testing.T) (*vcenter.TestingEnv, func(t *testing.T)) {

	vimHelper := VimSetupHelper()

	return vimHelper, func(t *testing.T) {
		if vimHelper.TestVim.Db != nil {
			log.Println("Exit status", vimHelper.TestVim.Db.Stats())
		}
		//		vimHelper.TestVim.Db.Close()
		t.Log("teardown test")
	}
}

func Test_cleanPorts(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func Test_discoveryCluster(t *testing.T) {

	env, teardown := setupTest(t)
	defer teardown(t)
	jetConf := env.TestVim.GetJetConfig()

	type args struct {
		vim *Vim
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "baseline",
			args:    args{env.TestVim},
			wantErr: false,
		},
		{
			name:    "add-rediscover",
			args:    args{env.TestVim},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := main.discoveryCluster(tt.args.vim); (err != nil) != tt.wantErr {
				t.Errorf("discoveryCluster() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(jetConf.GetNsx().GetTier0()) == 0 {
				t.Errorf("discoveryCluster() expect number of t0 > 0")
			}
			if len(jetConf.GetNsx().GetTier1()) == 0 {
				t.Errorf("discoveryCluster() expect number of t1 > 0 ")
			}
		})
	}
}

/*
   discover all tier0/-tier1, add new one / re-discover
*/
func Test_AddReDiscoveryCluster(t *testing.T) {

	env, teardown := setupTest(t)
	defer teardown(t)
	jetConf := env.TestVim.GetJetConfig()

	type args struct {
		vim *Vim
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "add-rediscover",
			args:    args{env.TestVim},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// discover
			if err := main.discoveryCluster(tt.args.vim); (err != nil) != tt.wantErr {
				t.Errorf("discoveryCluster() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(jetConf.GetNsx().GetEdgeClusterId()) == 0 {
				t.Errorf("discoveryCluster() expected cluster id set")
			}

			tierzeroCount := len(jetConf.GetNsx().GetTier0())
			tieroneCuount := len(jetConf.GetNsx().GetTier1())

			req1 := nsxtapi.RouterCreateReq{
				Name:       "jettison-test1",
				RouterType: "TIER0",
				ClusterID:  jetConf.GetNsx().GetEdgeClusterId(),
				Tags:       nil,
			}

			tier0, err := nsxtapi.CreateLogicalRouter(tt.args.vim.GetNsx(), req1)
			if (err != nil) != tt.wantErr {
				t.Errorf("discoveryCluster() error = %v, wantErr %v", err, tt.wantErr)
			}

			req2 := nsxtapi.RouterCreateReq{
				Name:       "jettison-test2",
				RouterType: "TIER1",
				ClusterID:  jetConf.GetNsx().GetEdgeClusterId(),
				Tags:       nil,
			}
			tier1, err := nsxtapi.CreateLogicalRouter(tt.args.vim.GetNsx(), req2)
			if (err != nil) != tt.wantErr {
				t.Errorf("discoveryCluster() error = %v, wantErr %v", err, tt.wantErr)
			}
			// re-discover
			if err := main.discoveryCluster(tt.args.vim); (err != nil) != tt.wantErr {
				t.Errorf("discoveryCluster() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (tierzeroCount + 1) != len(jetConf.GetNsx().GetTier0()) {
				t.Errorf("discoveryCluster() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (tieroneCuount + 1) != len(jetConf.GetNsx().GetTier1()) {
				t.Errorf("discoveryCluster() old count = %d, new count %d", tieroneCuount, len(jetConf.GetNsx().GetTier1()))
			}
			err = nsxtapi.DeleteLogicalRouter(tt.args.vim.GetNsx(), tier0.Id)
			if (err != nil) != tt.wantErr {
				t.Errorf("discoveryCluster() error = %v, wantErr %v", err, tt.wantErr)
			}
			err = nsxtapi.DeleteLogicalRouter(tt.args.vim.GetNsx(), tier1.Id)
			if (err != nil) != tt.wantErr {
				t.Errorf("discoveryCluster() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_discoveryDhcp(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func Test_discoveryEdges(t *testing.T) {
	tests := []struct {
		name string
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			//			discoveryCluster()
		})
	}
}

func Test_discoverySwitching(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func Test_discoveryTransportZon(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
