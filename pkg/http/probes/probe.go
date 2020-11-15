package probes

import (
	"fmt"
	"strings"
)

const (
	Max int = 32

	lowerBits uint64 = 0x00000000ffffffff
	upperBits uint64 = 0xffffffff00000000
)

type state int

const (
	down state = iota
	up
)

// Reporter is used by process subsystems to indicate when it is ready or not
// and when to stop the process
type Reporter interface {
	// Kill will lower the `liveness` probe
	Kill()

	// Ready will raise the `readiness` probe
	Ready()

	// NotReady will lower the `readiness` probe
	NotReady()
}

// Probe manages up to 32 `liveness` and `readiness` states
type Probe struct {
	cap    int
	states uint64
}

// New returns a new Probe instance if 1 <= cap <= 32
func New() *Probe {
	return &Probe{
		cap:    0,
		states: upperBits | lowerBits,
	}
}

// SetCapacity set the capacity of the probe, if 1 <= cap <= 32
func (p *Probe) SetCapacity(cap int) error {
	if cap < 1 || cap > Max {
		return fmt.Errorf("invalid capacity %d", cap)
	}

	p.cap = cap
	p.states <<= cap
	return nil
}

// Reporter returns a Reporter interface
func (p *Probe) Reporter(index int) Reporter {
	if index < 1 || index >= p.cap {
		return nil
	}
	return &reporter{
		index: uint(index),
		p:     p,
	}
}

// IsLive reports whether all live bits are Up
func (p *Probe) IsLive() bool {
	if p == nil {
		return false
	}
	return p.states&upperBits == upperBits
}

// IsReady reports whether all ready bits are Up
func (p *Probe) IsReady() bool {
	if p == nil {
		return false
	}
	return p.states&lowerBits == lowerBits
}

// String satsifies the Stringer interface
func (p *Probe) String() string {
	if p == nil {
		return "nil"
	} else if p.cap == 0 {
		return "no states"
	}

	var sb strings.Builder

	f := func(s uint64) {
		first := true
		for i := 0; i < p.cap; i++ {
			if !first {
				sb.WriteString(",")
			}
			first = false
			if s&1 == 1 {
				sb.WriteString("up")
			} else {
				sb.WriteString("down")
			}
			s >>= 1
		}
	}

	sb.WriteString("{ \"live\": [")
	f(p.states >> Max)
	sb.WriteString("], \"ready\": [")
	f(p.states)
	sb.WriteString("] }")
	return sb.String()
}

// GoString satisfies the `fmt.GoStringer` interface
func (p *Probe) GoString() string {
	return fmt.Sprintf("{ \"cap\": %d, \"ready\": %b }", p.cap, p.states)
}

func (p *Probe) set(index uint, value state) {
	switch value {
	case down:
		fmt.Printf("before: state: %d\n", p.states)
		p.states &^= (1 << index)
		fmt.Printf("after:  state: %d\n", p.states)
	case up:
		p.states |= (1 << index)
	}
}

type reporter struct {
	index uint
	p     *Probe
}

func (r *reporter) Kill() {
	r.p.set(r.index+uint(Max), down)
}

func (r *reporter) Ready() {
	r.p.set(r.index, up)
}

func (r *reporter) NotReady() {
	r.p.set(r.index, down)
}
