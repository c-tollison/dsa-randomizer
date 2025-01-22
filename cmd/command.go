package cmd

type Command interface {
	Command() string
	Help() string
	Run([]string) error
}

var Commands = []Command{
	&DbCommand{},
	&UserCommand{},
	&ProblemCommand{},
}
