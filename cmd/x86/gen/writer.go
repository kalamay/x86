package gen

import "io"

type Writer interface {
	io.Writer
	io.ByteWriter
	io.StringWriter
}
