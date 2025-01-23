# DSA Randomizer

This is a small side project to get familiar with some new concepts in Go. It also serves as a tracker of all the code challenge problems I work on. The idea of the CLI is allowing users to enter in problems they solve and every day they come back randomly get assigned one of those problems again. Essentially always forcing you to stay on top of many data structure and algorithms concepts.

General flow is to run

```bash
./dsa-randomizer problem start
```

This starts a new problem and the user has a set amount of time to solve that problem again before they break there streak.

After they have completed the problem run

```bash
./dsa-randomizer problem done
```

This increments there streak if done in the correct amount of time.

## Setup

To get started first create a file called `randomizer.db`

```bash
touch randomizer.db
```

Build the tool

```bash
go build .
```

Then run to run all the initial db migrations to allow the program to work correctly

```bash
./dsa-randomizer db setup
```

You can adjust your timer setting by using this command, passing in the number of hours you want to do a problem. Default is 1 hour

```bash
./dsa-randomizer user timer 5
```

Another useful tip is to add the dsa-randomizer to your path so you can run the program anywhere

```bash
export PATH=$PATH:$HOME/Development/dsa-randomizer/
```

I also made an alias for it so I can run commands like `dr user streak`

```bash
alias dr="dsa-randomizer"
```

## Nice to haves

Running this command will give you your current streak

```bash
./dsa-randomizer user streak
```

You can add problems by using this command, the n flag will be for the name of the problem and the -l flag is for the link directly to the problem

```bash
./dsa-randomizer problem add <name> <link>
```
