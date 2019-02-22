package main

import (
	"bufio"
	"os"
	"strings"
	//"github.com/pkg/term"
)

/*
func getch() []byte {
	t, _ := term.Open("/dev/tty")
	term.RawMode(t)
	bytes := make([]byte, 3)
	numRead, err := t.Read(bytes)
	t.Restore()
	t.Close()
	if err != nil {
		return nil
	}
	return bytes[0:numRead]
}
*/
func getch() []byte {
	reader := bufio.NewReader(os.Stdin)
	keyword, _ := reader.ReadString('\n')
	keyword = strings.Replace(keyword, "\n", "", -1)
	return []byte(keyword)
}
