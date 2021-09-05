package enpsql

import (
	"fmt"
	"strings"
	"time"
)

func newLogBlock() *logBlock { return &logBlock{st: &strings.Builder{}} }

type logBlock struct {
	st *strings.Builder
}

func (l *logBlock) String() string { return l.st.String() }

func (l *logBlock) date() {
	l.st.WriteRune('`')
	l.st.WriteString(time.Now().Format("15:04:05.000"))
	l.st.WriteRune('`')
	l.st.WriteRune(' ')
}

func (l *logBlock) Printf(f string, i ...interface{}) {
	f = strings.ReplaceAll(f, "\n", "\n`    |` ")
	l.date()
	fmt.Fprintf(l.st, f, i...)
	l.st.WriteRune('\n')
}
