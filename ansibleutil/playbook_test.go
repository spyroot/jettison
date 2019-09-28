package ansibleutil

import (
	"bytes"
	"github.com/spyroot/jettison/jettypes"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"testing"
)

func TestMakeNewPlaybook(t *testing.T) {
	type args struct {
		reader io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "valid template",
			args:    args{GenerateTemplate()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeNewPlaybook(tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeNewPlaybook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				assert.Nil(t, got)
			}

			if err == nil {
				assert.NotNil(t, got)
				log.Println(got)

				section1 := got.GetSection(jettypes.ControlType.String())
				assert.NotNil(t, section1)
				assert.Equal(t, section1.Hosts, jettypes.ControlType.String())

				section2 := got.GetSection(jettypes.WorkerType.String())
				assert.Equal(t, section2.Hosts, jettypes.WorkerType.String())

				section3 := got.GetSection(jettypes.IngressType.String())
				assert.Equal(t, section3.Hosts, jettypes.IngressType.String())

			}
		})
	}
}

func TestPlaybook_Transform(t *testing.T) {

	template := GenerateTemplate()
	playbook, err := MakeNewPlaybook(template)
	assert.NotNil(t, playbook)
	assert.Nil(t, err)

	type args struct {
		p   *Playbook
		old string
		new string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid transform",
			args: args{playbook,
				jettypes.ControlType.String(),
				"Test" + jettypes.ControlType.String(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.p.Transform(tt.args.old, tt.args.new)
			var buff bytes.Buffer
			err = playbook.Write(&buff)
			assert.Nil(t, err)

			t.Log(buff.String())

		})
	}
}

//
func TestPlaybook_AddVars(t *testing.T) {

	template := GenerateTemplate()
	playbook, err := MakeNewPlaybook(template)
	assert.NotNil(t, playbook)
	assert.Nil(t, err)

	type args struct {
		playbooks *Playbook
		section   string
		vars      map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "add emptry vars",
			args: args{playbook,
				jettypes.ControlType.String(),
				map[string]string{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.playbooks.AddVars(tt.args.section, tt.args.vars)
		})
	}
}

func TestGetSection(t *testing.T) {
	type args struct {
		p       *Playbook
		section string
	}

	template := GenerateTemplate()
	playbook, err := MakeNewPlaybook(template)
	assert.NotNil(t, playbook)
	assert.Nil(t, err)

	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *PlaybookSection
	}{
		{
			name:    "valid template",
			args:    args{playbook, ""},
			wantErr: false,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.p.GetSection(tt.args.section)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeNewPlaybook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				assert.Nil(t, got)
			}

			if err == nil {
				assert.NotNil(t, got)
			}
		})
	}
}
