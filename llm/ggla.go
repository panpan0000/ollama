package llm

import (
	"encoding/binary"
	"errors"
	"io"
	"slices"
)

type containerGGLA struct {
	version uint32
}

func (c *containerGGLA) Name() string {
	return "ggla"
}

func (c *containerGGLA) Decode(rs io.ReadSeeker) (model, error) {
	if err := binary.Read(rs, binary.LittleEndian, &c.version); err != nil {
		return nil, err
	}

	switch c.version {
	case 1:
	default:
		return nil, errors.New("invalid version")
	}

	model := newGGLA(c)
	err := model.decode(rs)
	return model, err
}

type ggla struct {
	*containerGGLA

	kv      KV
	tensors []Tensor
}

func newGGLA(container *containerGGLA) *ggla {
	return &ggla{
		containerGGLA: container,
		kv:            make(KV),
	}
}

func (m *ggla) decode(rs io.ReadSeeker) error {
	var r uint32
	if err := binary.Read(rs, binary.LittleEndian, &r); err != nil {
		return err
	}
	m.kv["r"] = r

	var alpha uint32
	if err := binary.Read(rs, binary.LittleEndian, &alpha); err != nil {
		return err
	}
	m.kv["alpha"] = alpha

	for {
		var dims uint32
		if err := binary.Read(rs, binary.LittleEndian, &dims); err != nil {
			return err
		}

		var namesize uint32
		if err := binary.Read(rs, binary.LittleEndian, &namesize); err != nil {
			return err
		}

		var t Tensor
		if err := binary.Read(rs, binary.LittleEndian, &t.Kind); err != nil {
			return err
		}

		t.Shape = make([]uint64, dims)
		for i := 0; uint32(i) < dims; i++ {
			var shape32 uint32
			if err := binary.Read(rs, binary.LittleEndian, &shape32); err != nil {
				return err
			}

			t.Shape[i] = uint64(shape32)
		}

		// ggla tensor shape is reversed
		// ref: https://github.com/ggerganov/llama.cpp/blob/29ae62d2ae163e2b68aa0ad3bf2ab4636de0c957/convert-lora-to-ggml.py#L44
		slices.Reverse(t.Shape)

		name := make([]byte, namesize)
		if err := binary.Read(rs, binary.LittleEndian, &name); err != nil {
			return err
		}

		t.Name = string(name)

		offset, err := rs.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}

		if _, err := rs.Seek((offset+31)&-32, io.SeekStart); err != nil {
			return err
		}

		offset, err = rs.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}

		t.Offset = uint64(offset)

		if _, err := rs.Seek(int64(t.size()), io.SeekCurrent); err != nil {
			return err
		}

		m.tensors = append(m.tensors, t)
	}
}

func (m *ggla) KV() KV {
	return m.kv
}

func (m *ggla) Tensor() []Tensor {
	return m.tensors
}

func (*ggla) ModelFamily() string {
	return "ggla"
}

func (*ggla) ModelType() string {
	panic("not implemented")
}

func (*ggla) FileType() string {
	panic("not implemented")
}

func (*ggla) NumLayers() uint32 {
	panic("not implemented")
}

func (*ggla) NumGQA() uint32 {
	panic("not implemented")
}

func (*ggla) NumEmbed() uint32 {
	panic("not implemented")
}

func (*ggla) NumHead() uint32 {
	panic("not implemented")
}

func (*ggla) NumHeadKv() uint32 {
	panic("not implemented")
}

func (*ggla) NumCtx() uint32 {
	panic("not implemented")
}