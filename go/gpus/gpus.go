// Package gpus provides a table of GPU vendor information, ported from Swarming:
// https://github.com/luci/luci-py/blob/887c873be30051f382da8b6aa8076a7467c80388/appengine/swarming/swarming_bot/api/platforms/gpu.py
// https://github.com/luci/luci-py/blob/887c873be30051f382da8b6aa8076a7467c80388/appengine/swarming/swarming_bot/api/platforms/win.py#L405
package gpus

import (
	"regexp"
	"strings"

	"go.skia.org/infra/go/util_generics"
)

type VendorID string
type vendorNameAndDevices struct {
	// Name is the canonical vendor name.
	Name string
	// Devices maps device IDs to names.
	Devices map[string]string
}

const (
	AMD    = "1002"
	Nvidia = "10de"
	Intel  = "8086"
	VMWare = "15ad"
)

// Static lookup tables:
var vendorMap = map[VendorID]vendorNameAndDevices{
	AMD: {
		Name: "AMD",
		Devices: map[string]string{
			// http://developer.amd.com/resources/ati-catalyst-pc-vendor-id-1002-li/
			"6613": "Radeon R7 240", // The table is incorrect
			"6646": "Radeon R9 M280X",
			"6779": "Radeon HD 6450/7450/8450",
			"679e": "Radeon HD 7800",
			"67ef": "Radeon RX 560",
			"6821": "Radeon R8 M370X", // "HD 8800M" or "R7 M380" based on rev_id
			"683d": "Radeon HD 7700",
			"9830": "Radeon HD 8400",
			"9874": "Carrizo",
		},
	},
	"1a03": {
		Name: "ASPEED",
		Devices: map[string]string{
			// https://pci-ids.ucw.cz/read/PC/1a03/2000
			// It seems all ASPEED graphics cards use the same device id (for driver reasons?)
			"2000": "ASPEED Graphics Family",
		},
	},
	Intel: {
		Name: "Intel",
		Devices: map[string]string{
			// http://downloadmirror.intel.com/23188/eng/config.xml
			"0046": "Ironlake HD Graphics",
			"0102": "Sandy Bridge HD Graphics 2000",
			"0116": "Sandy Bridge HD Graphics 3000",
			"0166": "Ivy Bridge HD Graphics 4000",
			"0412": "Haswell HD Graphics 4600",
			"041a": "Haswell HD Graphics",
			"0a16": "Intel Haswell HD Graphics 4400",
			"0a26": "Haswell HD Graphics 5000",
			"0a2e": "Haswell Iris Graphics 5100",
			"0d26": "Haswell Iris Pro Graphics 5200",
			"0f31": "Bay Trail HD Graphics",
			"1616": "Broadwell HD Graphics 5500",
			"161e": "Broadwell HD Graphics 5300",
			"1626": "Broadwell HD Graphics 6000",
			"162b": "Broadwell Iris Graphics 6100",
			"1912": "Skylake HD Graphics 530",
			"191e": "Skylake HD Graphics 515",
			"1926": "Skylake Iris 540/550",
			"193b": "Skylake Iris Pro 580",
			"22b1": "Braswell HD Graphics",
			"3e92": "Coffee Lake UHD Graphics 630",
			"5912": "Kaby Lake HD Graphics 630",
			"591e": "Kaby Lake HD Graphics 615",
			"5926": "Kaby Lake Iris Plus Graphics 640",
		},
	},
	"102b": {
		Name: "Matrox",
		Devices: map[string]string{
			"0522": "MGA G200e",
			"0532": "MGA G200eW",
			"0534": "G200eR2",
		},
	},
	Nvidia: {
		Name: "Nvidia",
		Devices: map[string]string{
			// ftp://download.nvidia.com/XFree86/Linux-x86_64/352.21/README/README.txt
			"06fa": "Quadro NVS 450",
			"08a4": "GeForce 320M",
			"08aa": "GeForce 320M",
			"0a65": "GeForce 210",
			"0df8": "Quadro 600",
			"0fd5": "GeForce GT 650M",
			"0fe9": "GeForce GT 750M Mac Edition",
			"0ffa": "Quadro K600",
			"104a": "GeForce GT 610",
			"11c0": "GeForce GTX 660",
			"1244": "GeForce GTX 550 Ti",
			"1401": "GeForce GTX 960",
			"1ba1": "GeForce GTX 1070",
			"1cb3": "Quadro P400",
			"2184": "GeForce GTX 1660",
		},
	},
}
var vendorNamesToIDs map[string]VendorID

// init builds the static table of vendor names to vendor IDs.
func init() {
	vendorNamesToIDs = make(map[string]VendorID, len(vendorMap))
	for id, nameAndDevices := range vendorMap {
		vendorNamesToIDs[strings.ToLower(nameAndDevices.Name)] = id
	}
}

// VendorNameToID returns the vendor ID for a given GPU vendor name. If unknown, returns "".
func VendorNameToID(name string) VendorID {
	// macOS 10.13 doesn't provide the vendor ID any more, so support reverse lookups on vendor name.
	return util_generics.Get(vendorNamesToIDs, strings.ToLower(name), "")
}

// IDsToNames turns a vendor ID and device ID into human-readable names if possible, falling back to
// passed-in defaults for each one where it isn't.
func IDsToNames(vendorID VendorID, fallbackVendorName, deviceID, fallbackDeviceName string) (vendorName, deviceName string) {
	vendorID = VendorID(strings.ToLower(string(vendorID)))
	deviceID = strings.ToLower(deviceID)

	vendorDevices, ok := vendorMap[vendorID]
	if !ok {
		return fallbackVendorName, fallbackDeviceName
	}

	deviceName = util_generics.Get(vendorDevices.Devices, deviceID, fallbackDeviceName)

	return vendorDevices.Name, deviceName
}

// extract pulls out the lowercased first group of the regex from a raw device ID, returning
// "UNKNOWN" on failure.
func extract(regex *regexp.Regexp, rawDeviceID string) string {
	// Raw device ID looks like CI\VEN_15AD&DEV_0405&SUBSYS_040515AD&REV_00\3&2B8E0B4B&0&78.
	groups := regex.FindStringSubmatch(rawDeviceID)
	if groups != nil {
		return strings.ToLower(groups[1])
	}
	return "UNKNOWN"
}

var gpuVendorRegex = regexp.MustCompile(`VEN_([0-9A-F]{4})`)
var gpuDeviceRegex = regexp.MustCompile(`DEV_([0-9A-F]{4})`)

// WindowsVendorAndDeviceID extracts the vendor ID and device ID from a Windows plug-and-play device
// ID string. If either is not found, it returns "UNKNOWN" for it.
func WindowsVendorAndDeviceID(pnpDeviceID string) (VendorID, string) {
	return VendorID(extract(gpuVendorRegex, pnpDeviceID)), extract(gpuDeviceRegex, pnpDeviceID)
}
