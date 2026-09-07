package buildinfo

import (
	"strings"
	"testing"
)

func TestReleaseMetadataIsPinned(t *testing.T) {
	if XrayUpstreamTag != "v26.7.11" {
		t.Fatalf("XrayUpstreamTag = %q", XrayUpstreamTag)
	}
	if XrayUpstreamCommit != "50231eaff98ccc31b5cbd247a721c16e97fe5ec1" {
		t.Fatalf("XrayUpstreamCommit = %q", XrayUpstreamCommit)
	}
	if XrayForkVersion != "v26.7.11-yz.5" {
		t.Fatalf("XrayForkVersion = %q", XrayForkVersion)
	}
	if XrayForkCommit != "dcb690846b525851f0ee8dc47388e110d4600042" {
		t.Fatalf("XrayForkCommit = %q", XrayForkCommit)
	}
	if SingBoxRequestedVersion != "v1.14.0" {
		t.Fatalf("SingBoxRequestedVersion = %q", SingBoxRequestedVersion)
	}
	if SingBoxResolvedVersion != "v1.14.0-yz.2" {
		t.Fatalf("SingBoxResolvedVersion = %q", SingBoxResolvedVersion)
	}
}

func TestReportContainsForkAndDependencyIdentity(t *testing.T) {
	report := Report("xboard-node", "v1.13-yz.1", "2026-07-25T00:00:00Z", "abc1234")
	for _, want := range []string{
		"xboard-node v1.13-yz.1",
		XrayForkVersion,
		XrayUpstreamCommit,
		XrayForkCommit,
		SingBoxRequestedVersion,
		SingBoxResolvedVersion,
		SingBoxUpstreamCommit,
		SingBoxForkCommit,
		ShadowsocksUpstreamVersion,
		ShadowsocksUpstreamCommit,
		AnyTLSUpstreamVersion,
		AnyTLSUpstreamCommit,
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("report %q does not contain %q", report, want)
		}
	}
}
