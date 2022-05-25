# CLI presenter
A small tool to produce demos on the command line. Works best in combination with `asciinema`, which records the terminal output.

IMPORTANT: This is mostly for my own use, so don't expect any support. Feel free to use it at your own risk. Pull Requests are welcome.

## Presentation file format
Description, will be output to terminal
```
! this is a description
```

Pause for 1 second
```
sleep 1
```

A command to be executed
```
$ execute me
```

Comments, will not be executed:
```
# this is a comment
```

### Example
```ini
# save this in presentation.txt

! Welcome to this example presentation. 
sleep 1
! This is how you can list the current directory's files:
$ ls -la
```

## Build
```sh
go build -o cli-presenter main.go
```

## Run presentation:
```sh
./cli-presenter presentation.txt
```

## Record presentation
```sh
asciinema rec -c "./cli-presenter presentation.txt"
```
