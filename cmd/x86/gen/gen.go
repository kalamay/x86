package gen

import (
	"encoding/xml"
	"errors"
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/kalamay/x86/instruction"
)

var ErrStructExpected = errors.New("Struct expected")

type Context struct {
	Package  string
	Var      string
	Receiver string
	IDType   string
	Imports  map[string]string
}

func ImportsOf(t reflect.Type, imports map[string]string, recur bool) map[string]string {
	if imports == nil {
		imports = make(map[string]string)
	}

	p := t.PkgPath()
	if len(p) > 0 {
		if _, ok := imports[p]; !ok {
			imports[p] = path.Base(p)
		}
	}

	if recur {
		switch t.Kind() {
		case reflect.Struct:
			for i := 0; i < t.NumField(); i++ {
				if f := t.Field(i); encodeFieldType(f) {
					ImportsOf(f.Type, imports, true)
				}
			}
		case reflect.Array, reflect.Ptr, reflect.Slice:
			ImportsOf(t.Elem(), imports, true)
		}
	}

	return imports
}

func (ctx *Context) Source(w Writer, is *instruction.Set) error {
	ctx.Imports = ImportsOf(reflect.TypeOf(is), ctx.Imports, true)

	fmt.Fprintf(w, "package %s\n\n", ctx.Package)

	if len(ctx.Imports) > 0 {
		w.WriteString("import (\n")
		for pkg, name := range ctx.Imports {
			if path.Base(pkg) != name {
				fmt.Fprintf(w, "\t%s %q\n", name, pkg)
			} else {
				fmt.Fprintf(w, "\t%q\n", pkg)
			}
		}
		w.WriteString(")\n\n")
	}

	w.WriteString("const (\n")
	for i := 0; i < len(is.Instructions); i++ {
		in := &is.Instructions[i]
		if i == 0 {
			fmt.Fprintf(w, "\t%s %s = iota // %s.\n", in.Name, ctx.IDType, in.Summary)
		} else {
			fmt.Fprintf(w, "\t%s // %s.\n", in.Name, in.Summary)
		}
	}
	w.WriteString(")\n\n")

	for i := 0; i < len(is.Instructions); i++ {
		in, max := &is.Instructions[i], [6]int{}
		for _, f := range in.Forms {
			for i := uint8(0); i < f.Operands.Len; i++ {
				if n := len(f.Operands.Val[i].String()); n >= max[i] {
					max[i] = n + 1
				}
			}
		}

		full := 0
		for _, v := range max[:] {
			full += v
		}

		fmt.Fprintf(w, "// %s: %s.\n//\n// Forms:", in.Name, in.Summary)
		for _, f := range in.Forms {
			cur := 0
			fmt.Fprintf(w, "\n//\t%s ", in.Name)
			for i := uint8(0); i < f.Operands.Len; i++ {
				n, _ := fmt.Fprintf(w, "%-*s", max[i], f.Operands.Val[i])
				cur += n
			}
			if f.ISA != 0 {
				fmt.Fprintf(w, " %*s[%s]", full-cur, "", strings.Join(f.ISA.Names(), ", "))
			}
		}
		fmt.Fprintf(w, "\nfunc (e %s) %s(a ...operand.Arg) { e.Emit(%s, a) }\n\n",
			ctx.Receiver,
			in.Name,
			in.Name,
		)
	}

	fmt.Fprintf(w, "\nvar %s = ", ctx.Var)
	return ctx.genValue(w, reflect.ValueOf(*is))
}

func (ctx *Context) genValue(w Writer, r reflect.Value) error {
	switch r.Kind() {
	case reflect.Bool, reflect.Float32, reflect.Float64:
		fmt.Fprintf(w, "%v", r.Interface())
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fmt.Fprintf(w, "%d", r.Interface())
		return nil
	case reflect.String:
		fmt.Fprintf(w, "%q", r.Interface())
		return nil
	case reflect.Struct:
		return ctx.genStruct(w, r)
	case reflect.Ptr:
		if r.IsNil() {
			w.WriteString("nil")
			return nil
		} else {
			w.WriteByte('&')
			return ctx.genStruct(w, r.Elem())
		}
	case reflect.Slice, reflect.Array:
		return ctx.genItems(w, r)
	}

	return fmt.Errorf("source.Encode: unexpeted kind %s (%s)", r.Kind(), r.Type().Name())
}

func (ctx *Context) genItems(w Writer, r reflect.Value) error {
	array := r.Kind() == reflect.Array

	if array {
		fmt.Fprintf(w, "[%d]", r.Len())
	} else {
		w.WriteString("[]")
	}

	nl := splitLines(r.Type().Elem(), false)

	if name, ok := ctx.Imports[r.Type().Elem().PkgPath()]; ok {
		w.WriteString(name)
		w.WriteByte('.')
	}
	w.WriteString(r.Type().Elem().Name())
	w.WriteString(" {")
	if nl {
		w.WriteByte('\n')
	}

	n := r.Len()
	if array {
		for ; n > 0; n-- {
			if !r.Index(n - 1).IsZero() {
				break
			}
		}
	}

	for i := 0; i < n; i++ {
		if err := ctx.genValue(w, r.Index(i)); err != nil {
			return err
		}
		w.WriteByte(',')
		if nl {
			w.WriteByte('\n')
		}
	}

	w.WriteString("}")
	return nil
}

func (ctx *Context) genStruct(w Writer, e reflect.Value) error {
	if e.Kind() != reflect.Struct {
		return ErrStructExpected
	}

	t := e.Type()
	eof := byte(' ')

	for i := 0; i < t.NumField(); i++ {
		if f, v := t.Field(i), e.Field(i); encodeField(f, v) {
			if splitLines(f.Type, true) {
				eof = '\n'
				break
			}
		}
	}

	if name, ok := ctx.Imports[t.PkgPath()]; ok {
		w.WriteString(name)
		w.WriteByte('.')
	}
	w.WriteString(t.Name())
	w.WriteString(" {")
	w.WriteByte(eof)

	for i := 0; i < t.NumField(); i++ {
		if f, v := t.Field(i), e.Field(i); encodeField(f, v) {
			w.WriteString(f.Name)
			w.WriteString(": ")
			if err := ctx.genValue(w, v); err != nil {
				return err
			}
			w.WriteByte(',')
			w.WriteByte(eof)
		}
	}
	w.WriteString("}")

	return nil
}

func splitLines(t reflect.Type, recur bool) bool {
	switch t.Kind() {
	case reflect.Struct, reflect.Ptr:
		return true
	case reflect.Slice, reflect.Array:
		return !recur || splitLines(t.Elem(), false)
	default:
		return false
	}
}

func encodeField(f reflect.StructField, v reflect.Value) bool {
	return !v.IsZero() && encodeFieldType(f)
}

func encodeFieldType(f reflect.StructField) bool {
	return f.IsExported() && !f.Type.AssignableTo(reflect.TypeOf(xml.Name{}))
}
