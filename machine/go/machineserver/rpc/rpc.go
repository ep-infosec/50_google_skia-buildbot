package rpc

import (
	"time"

	"go.skia.org/infra/machine/go/machine"
)

// URL paths.
const (
	APIPrefix = "/json/v1"

	MachineDescriptionRelativeURL           = "/machine/description/{id:.+}"
	MachineEventRelativeURL                 = "/machine/event/"
	PowerCycleCompleteRelativeURL           = "/powercycle/complete/{id:.+}"
	PowerCycleListRelativeURL               = "/powercycle/list"
	PowerCycleStateUpdateRelativeURL        = "/powercycle/state/update"
	SSEMachineDescriptionUpdatedRelativeURL = "/machine/sse/description/updated"

	MachineDescriptionURL           = APIPrefix + MachineDescriptionRelativeURL
	MachineEventURL                 = APIPrefix + MachineEventRelativeURL
	PowerCycleCompleteURL           = APIPrefix + PowerCycleCompleteRelativeURL
	PowerCycleListURL               = APIPrefix + PowerCycleListRelativeURL
	PowerCycleStateUpdateURL        = APIPrefix + PowerCycleStateUpdateRelativeURL
	SSEMachineDescriptionUpdatedURL = APIPrefix + SSEMachineDescriptionUpdatedRelativeURL
)

type SupplyChromeOSRequest struct {
	SSHUserIP          string
	SuppliedDimensions machine.SwarmingDimensions
}

type SetNoteRequest struct {
	Message string
	// User and Timestamp will be added by the server
}

type SetAttachedDevice struct {
	AttachedDevice machine.AttachedDevice
}

type PowerCycleStateForMachine struct {
	MachineID       string
	PowerCycleState machine.PowerCycleState
}

type UpdatePowerCycleStateRequest struct {
	Machines []PowerCycleStateForMachine
}

// FrontendDescription is the frontend representation of machine.Description.
// See that struct for details on the fields.
type FrontendDescription struct {
	MaintenanceMode     string
	IsQuarantined       bool
	Recovering          string
	AttachedDevice      machine.AttachedDevice
	Annotation          machine.Annotation
	Note                machine.Annotation
	Version             string
	PowerCycle          bool
	PowerCycleState     machine.PowerCycleState
	LastUpdated         time.Time
	Battery             int
	Temperature         map[string]float64
	RunningSwarmingTask bool
	LaunchedSwarming    bool
	DeviceUptime        int32
	SSHUserIP           string
	Dimensions          machine.SwarmingDimensions
}

// ListMachinesResponse is the full list of all known machines.
type ListMachinesResponse []FrontendDescription

// ListPowerCycleResponse is the list of machine ids that need powercycling.
type ListPowerCycleResponse []string

// ToListPowerCycleResponse converts the response from store.ListPowerCycle to a
// ListPowerCycleResponse.
func ToListPowerCycleResponse(machineIDs []string) ListPowerCycleResponse {
	return machineIDs
}

// ToFrontendDescription converts a machine.Description into a FrontendDescription.
func ToFrontendDescription(d machine.Description) FrontendDescription {
	return FrontendDescription{
		MaintenanceMode:     d.MaintenanceMode,
		IsQuarantined:       d.IsQuarantined,
		Recovering:          d.Recovering,
		AttachedDevice:      d.AttachedDevice,
		Annotation:          d.Annotation,
		Note:                d.Note,
		Version:             d.Version,
		PowerCycle:          d.PowerCycle,
		PowerCycleState:     d.PowerCycleState,
		LastUpdated:         d.LastUpdated,
		Battery:             d.Battery,
		Temperature:         d.Temperature,
		RunningSwarmingTask: d.RunningSwarmingTask,
		LaunchedSwarming:    d.LaunchedSwarming,
		DeviceUptime:        d.DeviceUptime,
		SSHUserIP:           d.SSHUserIP,
		Dimensions:          d.Dimensions,
	}
}

// ToListMachinesResponse converts []machine.Description into []FrontendDescription.
func ToListMachinesResponse(descriptions []machine.Description) []FrontendDescription {
	rv := make([]FrontendDescription, 0, len(descriptions))
	for _, d := range descriptions {
		rv = append(rv, ToFrontendDescription(d))
	}
	return rv
}
