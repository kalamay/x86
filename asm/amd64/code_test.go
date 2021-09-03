package amd64

import "testing"

func TestCode(t *testing.T) {
	c := Code(0)
	assertCodeValues(t, c)

	c.Set(0, 0x12)
	assertCodeValues(t, c, 0x12)

	c.Set(2, 0x34)
	assertCodeValues(t, c, 0x12, 0x00, 0x34)

	c.Insert(0x56)
	assertCodeValues(t, c, 0x56, 0x12, 0x00, 0x34)

	c.Clear(3)
	assertCodeValues(t, c, 0x56, 0x12)

	c.Insert(1)
	assertCodeValues(t, c, 1, 0x56, 0x12)

	c.Insert(2)
	assertCodeValues(t, c, 2, 1, 0x56, 0x12)

	c.Insert(3)
	assertCodeValues(t, c, 3, 2, 1, 0x56, 0x12)

	c.Insert(4)
	assertCodeValues(t, c, 4, 3, 2, 1, 0x56, 0x12)

	c.Insert(5)
	assertCodeValues(t, c, 5, 4, 3, 2, 1, 0x56, 0x12)

	c.Insert(6)
	assertCodeValues(t, c, 6, 5, 4, 3, 2, 1, 0x56)

	c.Insert(7)
	assertCodeValues(t, c, 7, 6, 5, 4, 3, 2, 1)
}

func assertCodeValues(t *testing.T, c Code, v ...byte) {
	if len(v) != c.Len() {
		t.Errorf("length failed: expect=%d, actual=%d", len(v), c.Len())
		return
	}

	for i, expect := range v {
		actual := c.At(i)
		if expect != actual {
			t.Errorf("value %d failed: expect=0x%02x, actual=0x%02x", i, expect, actual)
		}
	}
}
