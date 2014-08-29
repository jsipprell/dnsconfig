// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

// Read system DNS config from /etc/resolv.conf

package dnsconfig

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type DnsConfig struct {
	servers  []string // servers to use
	search   []string // suffixes to append to local name
	ndots    int      // number of dots in name to trigger absolute lookup
	timeout  int      // seconds before giving up on packet
	attempts int      // lost packets before giving up on server
	rotate   bool     // round robin among servers
}

// See resolv.conf(5) on a Linux machine.
// TODO(rsc): Supposed to call uname() and chop the beginning
// of the host name to get the default search domain.
func DnsReadConfig(filename string) (*DnsConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	conf := new(DnsConfig)
	// conf.servers = make([]string, 0, 3) // small, but the standard limit
	// conf.search = make([]string, 0)
	// conf.ndots = 1
	// conf.timeout = 5
	// conf.attempts = 2
	// conf.rotate = false

	for scanner.Scan() {
		line := scanner.Text()
		scannerLine := bufio.NewScanner(strings.NewReader(line))
		scannerLine.Split(bufio.ScanWords)
		var lineArr []string
		for scannerLine.Scan() {
			lineArr = append(lineArr, scannerLine.Text())
		}

		//empty line
		if len(lineArr) == 0 {
			continue
		}
		switch lineArr[0] {
		case "nameserver": // add one name server
			if len(lineArr) > 1 {
				conf.servers = append(conf.servers, lineArr[1])
			}

		case "domain": // set search path to just this domain
			if len(lineArr) > 1 {
				conf.search = make([]string, 1)
				conf.search[0] = lineArr[1]
			} else {
				conf.search = make([]string, 0)
			}

		case "search": // set search path to given servers
			conf.search = make([]string, len(lineArr)-1)
			for i := 0; i < len(conf.search); i++ {
				conf.search[i] = lineArr[i+1]
			}

		case "options": // magic options
			for i := 1; i < len(lineArr); i++ {
				s := lineArr[i]
				switch {
				case strings.HasPrefix(s, "ndots:"):
					v := strings.TrimPrefix(s, "ndots:")
					conf.ndots, _ = strconv.Atoi(v)
				case strings.HasPrefix(s, "timeout:"):
					v := strings.TrimPrefix(s, "timeout:")
					conf.timeout, _ = strconv.Atoi(v)
				case strings.HasPrefix(s, "attempts:"):
					v := strings.TrimPrefix(s, "attempts:")
					conf.attempts, _ = strconv.Atoi(v)
				case s == "rotate":
					conf.rotate = true
				}
			}
		}
	}
	return conf, nil
}

func DnsWriteConfig(conf *DnsConfig, filename string) (err error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	w := bufio.NewWriter(file)

	for _, server := range conf.servers {
		line := "nameserver " + server
		fmt.Fprintln(w, line)
	}
	for _, s := range conf.search {
		line := "search " + s
		fmt.Fprintln(w, line)
	}
	if conf.ndots != 0 || conf.timeout != 0 || conf.attempts != 0 || conf.rotate != false {
		line := "options"
		if conf.ndots != 0 {
			line += " ndots:" + strconv.Itoa(conf.ndots)
		}
		if conf.timeout != 0 {
			line += " timeout:" + strconv.Itoa(conf.timeout)
		}
		if conf.attempts != 0 {
			line += " attempts:" + strconv.Itoa(conf.attempts)
		}
		if conf.rotate == true {
			line += " rotate"
		}
		fmt.Fprintln(w, line)
	}
	w.Flush()

	return
}
