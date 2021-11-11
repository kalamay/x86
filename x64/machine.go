package x64

import (
	"errors"
	"fmt"

	"github.com/kalamay/x86/instruction"
	"github.com/kalamay/x86/operand"
)

var (
	ErrInstLength     = errors.New("instruction length exceeded 15 bytes")
	ErrFailedEncode   = errors.New("unable to encode instruction")
	ErrSymbolDefinied = errors.New("symbol already defined")
)

type Machine struct {
	labels  map[string]int
	pending []pending
	encoded []instruction.Format
	written int
}

type pending struct {
	id   int
	call *EmitCall
}

func NewMachine() *Machine {
	return &Machine{
		labels: map[string]int{},
	}
}

func (m *Machine) Open() {
	for k := range m.labels {
		delete(m.labels, k)
	}
	m.pending = m.pending[:0]
	m.encoded = m.encoded[:0]
	m.written = 0
}

func (m *Machine) Emit(e *Emit, call *EmitCall) {
	id := len(m.encoded)
	m.encoded = append(m.encoded, instruction.Format{})

	if m.ready(id, call) {
		if err := m.encode(id, call); err != nil {
			e.AddError(err, call)
		} else if len(m.pending) == 0 {
			m.write(e, id)
		}
	} else {
		m.pending = append(m.pending, pending{id: id, call: call})
	}
}

func (m *Machine) Label(e *Emit, label *EmitLabel) {
	name := label.Name()
	if _, ok := m.labels[name]; ok {
		e.AddError(ErrSymbolDefinied, label)
		return
	}
	m.labels[name] = len(m.encoded)

	enc := 0
	for _, pend := range m.pending {
		if m.ready(pend.id, pend.call) {
			if err := m.encode(pend.id, pend.call); err != nil {
				e.AddError(err, pend.call)
			}
			enc++
		} else {
			break
		}
	}

	if enc > 0 {
		copy(m.pending, m.pending[enc:])
		m.pending = m.pending[:len(m.pending)-enc]

		for i := m.written; i < len(m.encoded); i++ {
			if m.encoded[i].Len == 0 {
				break
			}
			m.write(e, i)
		}
	}
}

func (m *Machine) Close(e *Emit) {
	if len(m.pending) > 0 {
		label := ""
		for _, arg := range m.pending[0].call.Args {
			if arg, ok := arg.(operand.Label); ok {
				if _, ok := m.labels[string(arg)]; !ok {
					label = string(arg)
				}
			}
		}
		if label == "" {
			e.AddError(ErrFailedEncode, m.pending[0].call)
		} else {
			e.AddError(fmt.Errorf("symbol %q is not defined", label), m.pending[0].call)
		}
	}
}

func (m *Machine) ready(id int, call *EmitCall) bool {
	for i := range call.Args {
		if arg, ok := call.Args[i].(operand.Label); ok {
			to, ok := m.labels[string(arg)]
			if !ok {
				return false
			}
			if r, ok := instruction.ResolveRel(id, to, m.encoded); ok {
				call.Args[i] = r
			} else {
				return false
			}
		}
	}
	return true
}

func (m *Machine) encode(id int, call *EmitCall) error {
	if m.encoded[id].Len > 0 {
		return nil
	}

	enc, err := Select(call.Instruction, call.Args)
	if err != nil {
		return err
	}

	enc.Prefix.Encode(&m.encoded[id], call.Args)
	switch {
	case enc.EVEX.Encode(&m.encoded[id], call.Args):
	case enc.VEX.Encode(&m.encoded[id], call.Args):
	case enc.REX.Encode(&m.encoded[id], call.Args):
	}
	enc.Opcode.Encode(&m.encoded[id], call.Args)
	enc.ModRM.Encode(&m.encoded[id], call.Args)
	enc.RegisterByte.Encode(&m.encoded[id], call.Args)
	enc.Immediate.Encode(&m.encoded[id], call.Args)
	enc.CodeOffset.Encode(&m.encoded[id], call.Args)
	enc.DataOffset.Encode(&m.encoded[id], call.Args)

	return nil
}

func (m *Machine) write(e *Emit, id int) {
	e.Write(m.encoded[id].Bytes())
	m.written++
}
