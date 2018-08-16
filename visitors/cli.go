package visitors

import (
	"fmt"

	"github.com/linuxboot/fiano/uefi"
)

var visitorRegistry = map[string]visitorEntry{}

type visitorEntry struct {
	numArgs       int
	createVisitor func([]string) (uefi.Visitor, error)
}

// RegisterCLI registers a function `createVisitor` to be called when parsing
// the arguments with `ParseCLI`. For a Visitor to be accessible from the
// command line, it should have an init function which registers a
// `createVisitor` function here.
func RegisterCLI(name string, numArgs int, createVisitor func([]string) (uefi.Visitor, error)) {
	if _, ok := visitorRegistry[name]; ok {
		panic(fmt.Sprintf("two visitors registered the same name: '%s'", name))
	}
	visitorRegistry[name] = visitorEntry{
		numArgs:       numArgs,
		createVisitor: createVisitor,
	}
}

// ParseCLI constructs a list of visitors from the given CLI argument list.
// TODO: display some type of help message
func ParseCLI(args []string) ([]uefi.Visitor, error) {
	visitors := []uefi.Visitor{}
	for len(args) > 0 {
		cmd := args[0]
		args = args[1:]
		o, ok := visitorRegistry[cmd]
		if !ok {
			return []uefi.Visitor{}, fmt.Errorf("could not find visitor '%s'", cmd)
		}
		if o.numArgs > len(args) {
			return []uefi.Visitor{}, fmt.Errorf("too few arguments for visitor '%s', got %d, expected %d",
				cmd, len(args), o.numArgs)
		}
		visitor, err := o.createVisitor(args[:o.numArgs])
		if err != nil {
			return []uefi.Visitor{}, err
		}
		visitors = append(visitors, visitor)
		args = args[o.numArgs:]
	}
	return visitors, nil
}

// ExecuteCLI applies each Visitor over the firmware in sequence.
func ExecuteCLI(n uefi.Firmware, v []uefi.Visitor) error {
	for i := range v {
		if err := n.Apply(v[i]); err != nil {
			return err
		}
	}
	return nil
}