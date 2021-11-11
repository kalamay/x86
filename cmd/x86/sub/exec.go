package sub

type ExecCmd struct {
	AsCmd

	Debug bool `short:"d" help:"Run in debugging mode. (requires lldb)"`
}
