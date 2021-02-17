package common

import (
	"flag"
	"strings"
)

// StringsFlag collects multiple usages of the same flag into an array of strings.
// Duplicate values will be ignored.
type StringsFlag struct {
	list      *[]string
	isDefault bool
	flag      *flag.Flag
}

// StringArrFlag creates and registers a new StringFlag with the given FlagSet.
// If no FlagSet is passed, flag.CommandLine will be used as target FlagSet.
func StringArrFlag(fs *flag.FlagSet, name, def, usage string) *StringsFlag {
	var arr *[]string
	if def != "" {
		arr = &[]string{def}
	} else {
		arr = &[]string{}
	}

	return StringArrVarFlag(fs, arr, name, usage)
}

// StringArrVarFlag creates and registers a new StringFlag with the give FlagSet
func StringArrVarFlag(fs *flag.FlagSet, arr *[]string, name, usage string) *StringsFlag {
	if fs == nil {
		fs = flag.CommandLine
	}
	f := NewStringsFlag(arr)
	f.Register(fs, name, usage)
	return f
}

// NewStringsFlag creates a new, but unregistered StringsFlag instance.
func NewStringsFlag(arr *[]string) *StringsFlag {
	if arr == nil {
		panic("No target array")
	}

	return &StringsFlag{list: arr, isDefault: true}
}

// Register registers the StringsFlag instance with a FlagSet.
// A valid FlagSet must be used.
// Register panics if the flag is already registered.
func (f *StringsFlag) Register(fs *flag.FlagSet, name, usage string) {
	if f.flag != nil {
		panic("StringsFlag is already registered")
	}

	fs.Var(f, name, usage)
	f.flag = fs.Lookup(name)
	if f.flag == nil {
		panic("Failed to lookup registered flag")
	}

	if len(*f.list) > 0 {
		f.flag.DefValue = (*f.list)[0]
	}
}

// String joins all ite's values set into a comma-separated string.
func (f *StringsFlag) String() string {
	if f == nil || f.list == nil {
		return ""
	}

	l := *f.list
	return strings.Join(l, ", ")
}

// Set is used to pass usage of the flag to StringsFlag.
// Set adds the new value to the backing array.
// The array will be emptied on Set, if the backing array still contains the default value
func (f *StringsFlag) Set(v string) error {
	if f.isDefault {
		*f.list = []string{v}
	} else {
		for _, old := range *f.list {
			if old == v {
				return nil
			}
		}
		*f.list = append(*f.list, v)
	}
	f.isDefault = false
	return nil
}
