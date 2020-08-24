package printer

import (
	"encoding/json"
	"io"

	"github.com/zcong1993/notifiers/types"
)

type Printer struct {
	output io.Writer
}

func NewPrinter(output io.Writer) *Printer {
	return &Printer{output: output}
}

func (p *Printer) Notify(msg *types.Message) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = p.output.Write(b)
	return err
}

func (p *Printer) GetName() string {
	return "printer"
}
