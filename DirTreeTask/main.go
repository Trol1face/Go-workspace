package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
)

const (
	midTab string = "├───"
	endTab string = "└───"
)

func getFileSize(dirEntry fs.FileInfo) string {
	dirEntrySize := strconv.FormatInt(dirEntry.Size(), 10)
	if dirEntrySize == "0" {
		return " (empty)"
	}
	return " (" + dirEntrySize + "b)"
}

func isLastFile(fileNumber int, dirEntriesAmount int) bool {
	return fileNumber >= dirEntriesAmount-1
}

func treeWriter(path string, printFiles bool, tabs string) string {
	openedFile, err := os.Open(path)
	if err != nil {
		fmt.Errorf(err.Error())
	}

	dirEntries, err := openedFile.Readdir(0)
	if err != nil {
		fmt.Errorf(err.Error())
	}

	dirEntriesAmount := len(dirEntries)
	writeString := ""
	lastDirPos := -1

	for dirEntryNumber := 0; dirEntryNumber < dirEntriesAmount; dirEntryNumber++ {
		dirEntry := dirEntries[dirEntryNumber]
		if dirEntry.IsDir() {
			lastDirPos = dirEntryNumber
		}
	}

	for dirEntryNumber := 0; dirEntryNumber < dirEntriesAmount; dirEntryNumber++ {

		dirEntry := dirEntries[dirEntryNumber]

		if !printFiles && !dirEntry.IsDir() {
			continue
		}

		dirEntryName := dirEntry.Name()
		dirEntrySize := getFileSize(dirEntry)

		if !dirEntry.IsDir() && !isLastFile(dirEntryNumber, dirEntriesAmount) {

			writeString += tabs + midTab + dirEntryName + dirEntrySize + "\n"

		} else if dirEntry.IsDir() && !isLastFile(dirEntryNumber, dirEntriesAmount) {
			nextTabs := tabs

			if !printFiles && dirEntryNumber == lastDirPos {

				writeString += tabs + endTab + dirEntryName + "\n"
				nextTabs += "\t"

			} else {

				writeString += tabs + midTab + dirEntryName + "\n"
				nextTabs += "│\t"

			}

			nextPath := path + string(os.PathSeparator) + dirEntry.Name()
			writeString += treeWriter(nextPath, printFiles, nextTabs)

		} else if dirEntry.IsDir() && isLastFile(dirEntryNumber, dirEntriesAmount) {

			writeString += tabs + endTab + dirEntryName + "\n"
			nextTabs := tabs + "\t"
			nextPath := path + string(os.PathSeparator) + dirEntry.Name()
			writeString += treeWriter(nextPath, printFiles, nextTabs)

		} else if !dirEntry.IsDir() && isLastFile(dirEntryNumber, dirEntriesAmount) {

			writeString += tabs + endTab + dirEntryName + dirEntrySize + "\n"

		}
	}
	return writeString
}

func dirTree(out io.Writer, path string, printFiles bool) error {

	tabs := ""
	out.Write([]byte(treeWriter(path, printFiles, tabs)))

	return nil
}

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
