package wd

import (
	"bytes"
	"io"
	"net/url"
	"strings"
)

type logFilter struct {
	w io.Writer
}

func extractPath(p []byte) (o []byte) {
	const source = ", source: "
	idx := bytes.LastIndex(p, []byte(source))
	if idx == -1 {
		return nil
	}
	sub := p[idx+len(source):]
	parts := bytes.Split(sub, []byte{' '})
	if len(parts) < 2 {
		return sub
	}
	u, err := url.Parse(string(parts[0]))
	if err != nil {
		return parts[0]
	}
	o = []byte(strings.TrimPrefix(u.Path, "/"))
	o = append(o, ' ')
	o = append(o, parts[1]...)
	return o
}

func extractMessage(p []byte) (o []byte) {
	startIndex := 0
	endIndex := 0
	escaped := false
	for i, b := range p {
		if b == '\\' {
			escaped = true
			continue
		}
		escaped = false
		if !escaped && b == '"' {
			if startIndex == 0 {
				startIndex = i
			} else {
				endIndex = i
				break
			}
		}
	}
	return append(o, p[startIndex+1:endIndex]...)
}

func (f *logFilter) Write(p []byte) (n int, err error) {
	if !bytes.Contains(p, []byte("CONSOLE(")) {
		return len(p), nil
	}
	o := extractMessage(p)
	o = append(o, ':')
	o = append(o, ' ')
	o = append(o, extractPath(p)...)
	return f.w.Write(o)
}
