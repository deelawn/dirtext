# dirtext
generates a textual representation of the current directory, ignoring hidden files and folders and anything in the .gitignore file

```
go install github.com/deelawn/dirtext/cmd/dirtext

dirtext

Warning: couldn't load .gitignore: open /Users/dylan/repos/dirtext/.gitignore: no such file or directory
dirtext
├── README.md
├── cmd
│   ├── dirtext
│   │   ├── main.go
├── go.mod
```
