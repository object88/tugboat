package probes

import (
	"testing"
)

func Test_Probe_New(t *testing.T) {
	p := New()
	if !p.IsLive() {
		t.Errorf("unexpected false for IsLive")
	} else if !p.IsReady() {
		t.Errorf("unexpected true for IsReady")
	}
	if p.states != 0xffffffffffffffff {
		t.Errorf("unexpected initial state: %x", p.states)
	}
}

func Test_Probe_SetCapacity(t *testing.T) {
	p := New()
	err := p.SetCapacity(2)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	} else if p == nil {
		t.Errorf("unexpected nil return from `New`")
	}

	if !p.IsLive() {
		t.Errorf("unexpected false for IsLive")
	} else if p.IsReady() {
		t.Errorf("unexpected true for IsReady")
	}

	if p.cap != 2 {
		t.Errorf("unexpected cap set")
	}
}

func Test_Probe_BadCapacity(t *testing.T) {
	tcs := []struct {
		name     string
		capacity int
	}{
		{
			name:     "negative",
			capacity: -1,
		},
		{
			name:     "zero",
			capacity: 0,
		},
		{
			name:     "over",
			capacity: 33,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			p := New()
			err := p.SetCapacity(tc.capacity)
			if err == nil {
				t.Errorf("did not get expected error")
			}
		})
	}
}

func Test_Set_Ready_OneOfOne(t *testing.T) {
	p := New()
	p.SetCapacity(1)
	p.Reporter(0).Ready()
	if !p.IsReady() {
		t.Errorf("probe is not ready")
	}
}

func Test_Set_Ready_OneOfTwo(t *testing.T) {
	p := New()
	p.SetCapacity(2)
	p.Reporter(0).Ready()
	if p.IsReady() {
		t.Errorf("probe is ready")
	}
}

func Test_Set_Ready_TwoOfTwo(t *testing.T) {
	p := New()
	p.SetCapacity(2)
	p.Reporter(0).Ready()
	p.Reporter(1).Ready()
	if !p.IsReady() {
		t.Errorf("probe is not ready")
	}
}

func Test_Probe_Kill(t *testing.T) {
	p := New()
	p.SetCapacity(1)
	r := p.Reporter(0)
	r.Kill()
	if p.IsLive() {
		t.Errorf("probe is still live: %s", p)
	}
}
