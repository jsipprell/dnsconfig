package dnsconfig

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
)

type flusher interface {
	Flush() error
}

// Atomically replace a resolv.conf
func DnsReplaceConfig(conf *DnsConfig, filename string) (err error) {
	var (
		fi       os.FileInfo
		f        *os.File
		renameTo string
	)
	if fi, err = os.Stat(filename); !os.IsNotExist(err) {
		defer func() {
			if f != nil {
				filename = f.Name()
				f.Chmod(fi.Mode())
				f.Close()
				if renameTo != "" && filename != renameTo && err == nil {
					err = os.Rename(filename, renameTo)
				}
			}
		}()

		if f, err = ioutil.TempFile(filepath.Dir(filename), filepath.Base(filename)); err != nil {
			return
		}
		renameTo = filename
		filename = f.Name()
	}

	if f != nil {
		err = writeConfigFile(conf, bufio.NewWriter(f))
	} else {
		err = DnsWriteConfig(conf, filename)
	}

	return
}
