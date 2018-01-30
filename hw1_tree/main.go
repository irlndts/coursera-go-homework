package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
)

type ByName []os.FileInfo

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name() < a[j].Name() }

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	return recursiveTree(1, out, dir, printFiles, []bool{})
}

func recursiveTree(level int, out io.Writer, dir *os.File, printFiles bool, t []bool) error {
	files, err := dir.Readdir(0)
	if err != nil {
		return err
	}

	if !printFiles {
		// delete files
		var dirs []os.FileInfo
		for _, f := range files {
			if f.IsDir() {
				dirs = append(dirs, f)
			}
		}
		files = dirs
	}

	sort.Sort(ByName(files))
	for i, f := range files {
		var newt []bool
		var str string
		if i != len(files)-1 {
			str = tree(t) + "├───" + f.Name()
			newt = append(t, true)
		} else {
			str = tree(t) + "└───" + f.Name()
			newt = append(t, false)
		}

		if !f.IsDir() {
			size := int(f.Size())
			if size != 0 {
				str += " (" + strconv.Itoa(size) + "b)"
			} else {
				str += " (empty)"
			}
		}
		fmt.Fprintf(out, "%s\n", str)

		if f.IsDir() {
			deeperDir, err := os.Open(dir.Name() + "/" + f.Name())
			if err != nil {
				return err
			}
			if err := recursiveTree(level+1, out, deeperDir, printFiles, newt); err != nil {
				return err
			}
		}
	}
	return nil
}

func tree(t []bool) string {
	var str string
	for _, b := range t {
		if b {
			str += "│\t"
		} else {
			str += "\t"
		}
	}
	return str
}
