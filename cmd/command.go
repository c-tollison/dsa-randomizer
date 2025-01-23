package cmd

type Command interface {
	Command() string
	Help() string
	Run([]string) error
}
