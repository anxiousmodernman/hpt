package main

import (
	"bytes"
	"testing"
)

func TestGetOSRelease(t *testing.T) {
	buf := bytes.NewBufferString(osReleaseFile)

	osr, err := ScanOSRelease(buf)
	if err != nil {
		t.Fail()
	}

	if osr.ID != "\"centos\"" {
		t.Errorf("expected centos got: %s", osr.ID)
	}
}

var osReleaseFile = `
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
`
