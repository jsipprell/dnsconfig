package main

import (
	"fmt"
	"github.com/jsipprell/dnsconfig"
	"os"
	"strconv"
	"strings"
)

type setFunc func(*dnsconfig.DnsConfig, string) error

func addns(conf *dnsconfig.DnsConfig, arg string) error {
	conf.Servers = append(conf.Servers, strings.TrimSpace(arg))
	return nil
}

func addsearch(conf *dnsconfig.DnsConfig, arg string) error {
	conf.Search = append(conf.Search, strings.TrimSpace(arg))
	return nil
}

func ndots(conf *dnsconfig.DnsConfig, arg string) error {
	i, err := strconv.Atoi(arg)
	if err != nil {
		return err
	}
	if i < 0 {
		return fmt.Errorf("ndots out of range: %d", i)
	}

	conf.Ndots = i
	return nil
}

func main() {
	config, err := dnsconfig.DnsReadConfig(dnsconfig.ResolvPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var (
		handler setFunc
		update  int
	)

	for _, arg := range os.Args[1:] {
		if handler != nil {
			if err := handler(config, arg); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
			handler = nil
			update++
			continue
		}
		switch strings.ToLower(arg) {
		case "purge":
			config.Servers = make([]string, 0, 2)
			config.Search = make([]string, 0, 2)
		case "purge_ns":
			config.Servers = make([]string, 0, 2)
		case "purge_search":
			config.Search = make([]string, 0, 2)
		case "ndots":
			handler = ndots
		case "ns":
			handler = addns
		case "search":
			handler = addsearch
		default:
			fmt.Fprintf(os.Stderr, "unsupported action: %q", arg)
			os.Exit(10)
		}
	}

	if update > 0 {
		if err := dnsconfig.DnsReplaceConfig(config, dnsconfig.ResolvPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(100)
		}
	}

}
