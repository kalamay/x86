// +build !amd64

package sub

import (
	"fmt"
	"os"

	"github.com/kalamay/x86/instruction"
)

func (cli *ExecCmd) Run(data *instruction.Set) error {
	fmt.Fprintf(os.Stderr, "exec not supported on platform")
	os.Exit(1)
}
