package main

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type Distro string

const (
	Centos = Distro("centos")
	Ubuntu = Distro("ubuntu")
)

// OSRelease is a
type OSRelease struct {
	ID      string
	Version string
}

// GetOSRelease can be called to return a structure that tells us which Linux
// distribution we're running on.
func GetOSRelease() (OSRelease, error) {

	f, err := os.Open("/etc/os-release")
	if err != nil {
		return OSRelease{}, errors.Wrap(err, "could not read file")
	}
	return ScanOSRelease(f)
}

// ScanOSRelease accepts a reader built from /etc/os-release and parses important
// values from it into an OSRelease.
func ScanOSRelease(r io.Reader) (OSRelease, error) {

	scnr := bufio.NewScanner(r)

	var id string
	var version string
	for scnr.Scan() {
		line := scnr.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		splitted := strings.Split(line, "=")
		if len(splitted) == 2 {
			if splitted[0] == "ID" {
				id = splitted[1]
			}
			if splitted[0] == "VERSION_ID" {
				version = splitted[1]
			}
		}
	}
	if id == "" {
		return OSRelease{}, errors.New("could not determine os ID")
	}
	return OSRelease{id, version}, nil
}

// GetOSRelease parses the /etc/os-release file, which looks like this on Centos7
/*
ID="centos"
ID_LIKE="rhel fedora"
VERSION_ID="7"
PRETTY_NAME="CentOS Linux 7 (Core)"
ANSI_COLOR="0;31"
CPE_NAME="cpe:/o:centos:centos:7"
HOME_URL="https://www.centos.org/"
BUG_REPORT_URL="https://bugs.centos.org/"

CENTOS_MANTISBT_PROJECT="CentOS-7"
CENTOS_MANTISBT_PROJECT_VERSION="7"
REDHAT_SUPPORT_PRODUCT="centos"
REDHAT_SUPPORT_PRODUCT_VERSION="7"
*/
