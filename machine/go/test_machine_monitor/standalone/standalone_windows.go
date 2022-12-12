package standalone

import (
	"context"

	"github.com/shirou/gopsutil/host"
	"github.com/yusufpapurcu/wmi"
	"go.skia.org/infra/go/skerr"
	"go.skia.org/infra/machine/go/test_machine_monitor/standalone/crossplatform"
	"go.skia.org/infra/machine/go/test_machine_monitor/standalone/windows"
)

func OSVersions(ctx context.Context) ([]string, error) {
	platform, _, version, err := host.PlatformInformationWithContext(ctx)
	// Return values are like these:
	// "Microsoft Windows Server 2019 Datacenter", "Server", "10.0.17763 Build 17763"
	// "Microsoft Windows 10 Pro", "Standalone Workstation", "10.0.19043 Build 19043"

	if err != nil {
		return nil, skerr.Wrapf(err, "failed to get Windows version")
	}
	return windows.OSVersions(platform, version)
}

func CPUs(ctx context.Context) ([]string, error) {
	return crossplatform.CPUs("", "")
}

// GPUs returns Swarming-style dimensions representing all GPUs on the host. Each GPU may yield up
// to 3 returned elements: "vendorID", "vendorID:deviceID", and "vendorID:deviceID-driverVersion".
// If no GPUs are found or if the host is running within VMWare, returns ["none"].
func GPUs(ctx context.Context) ([]string, error) {
	var results []windows.GPUQueryResult
	err := wmi.Query("SELECT DriverVersion, PNPDeviceID FROM Win32_VideoController", &results)
	if err != nil {
		return nil, skerr.Wrapf(err, "failed to run WMI query to get GPU info")
	}
	return windows.GPUs(results), nil
}
