package system

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

type HostsSection struct {
	name string
	path string

	hosts map[string][]string
}

func NewHostsSection(name string) (*HostsSection, error) {
	path := "/etc/hosts"

	if runtime.GOOS == "windows" {
		path = os.ExpandEnv("${SystemRoot}/System32/drivers/etc/hosts")
	}

	if val, ok := os.LookupEnv("HOSTS_PATH"); ok {
		path = val
	}

	return &HostsSection{
		name: name,
		path: path,

		hosts: make(map[string][]string),
	}, nil
}

func (s *HostsSection) Add(address string, hosts ...string) {
	s.hosts[address] = hosts
}

func (s *HostsSection) Remove(address string) {
	delete(s.hosts, address)
}

func (s *HostsSection) Clear() {
	clear(s.hosts)
}

func (s *HostsSection) Flush() error {
	ln := "\n"

	if runtime.GOOS == "windows" {
		ln = "\r\n"
	}

	file, err := os.OpenFile(s.path, os.O_RDWR, 0644)

	if err != nil {
		return err
	}

	data, err := io.ReadAll(file)

	if err != nil {
		return err
	}

	text := string(data)

	headerStart := fmt.Sprintf("# Start Section %s%s", s.name, ln)
	headerEnd := fmt.Sprintf("# End Section %s%s", s.name, ln)

	sectionStart := strings.Index(text, headerStart)
	sectionEnd := strings.LastIndex(text, headerEnd)

	if sectionStart > 0 && sectionEnd > 0 {
		text = text[:sectionStart] + text[sectionEnd+len(headerEnd):]
	}

	if len(s.hosts) > 0 {
		text += headerStart

		for address, hosts := range s.hosts {
			text += fmt.Sprintf("%s %s%s", address, strings.Join(hosts, " "), ln)
		}

		text += fmt.Sprintf("%s%s", headerEnd, ln)
	}

	text = strings.TrimRight(text, ln) + ln

	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	if err := file.Truncate(0); err != nil {
		return err
	}

	if _, err := file.WriteString(text); err != nil {
		return err
	}

	return nil
}
