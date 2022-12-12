package main

import (
	"flag"
	"fmt"
	"sort"

	"go.skia.org/infra/go/common"
	"go.skia.org/infra/go/gce"
	"go.skia.org/infra/go/sklog"
	"go.skia.org/infra/go/util"
	compute "google.golang.org/api/compute/v0.beta"
)

var (
	instances      = common.NewMultiStringFlag("instance", nil, "Instance(s) to add scopes.")
	scopes         = common.NewMultiStringFlag("scope", nil, "Scope(s) to add.")
	serviceAccount = flag.String("service_account", "", "Override the existing service account on all instances with this one, if set.")
	project        = flag.String("project", gce.PROJECT_ID_SERVER, "GCE project.")
	zone           = flag.String("zone", gce.ZONE_DEFAULT, "GCE zone.")
)

func main() {
	common.Init()

	// Setup.
	if instances == nil || len(*instances) == 0 {
		sklog.Fatal("--instance is required.")
	}
	if scopes == nil || len(*scopes) == 0 {
		sklog.Fatal("--scope is required.")
	}

	gcloud, err := gce.NewLocalGCloud(*project, *zone)
	if err != nil {
		sklog.Fatal(err)
	}
	is := gcloud.Service().Instances

	sklog.Infof("Running on instances: %v", instances)

	// Determine the set of scopes for each instance.
	emailsByInstance := make(map[string]string, len(*instances))
	scopesByInstance := make(map[string]util.StringSet, len(*instances))
	for _, name := range *instances {
		inst, err := is.Get(*project, *zone, name).Do()
		if err != nil {
			sklog.Fatal(err)
		}
		if len(inst.ServiceAccounts) != 1 {
			sklog.Fatalf("Instances must have exactly one service account but %s has %d", name, len(inst.ServiceAccounts))
		}
		emailsByInstance[name] = inst.ServiceAccounts[0].Email
		if *serviceAccount != "" {
			emailsByInstance[name] = *serviceAccount
		}
		s := util.NewStringSet(inst.ServiceAccounts[0].Scopes)
		for _, scope := range *scopes {
			s[scope] = true
		}
		scopesByInstance[name] = s
		sklog.Infof("Scopes for %s:\n%v", name, s.Keys())
	}

	// For each instance, stop it, apply the scopes, and restart it.
	group := util.NewNamedErrGroup()
	for name, s := range scopesByInstance {
		name := name
		instanceScopes := s.Keys()
		sort.Strings(instanceScopes)
		group.Go(name, func() error {
			if err := gcloud.CheckOperation(is.Stop(*project, *zone, name).Do()); err != nil {
				return fmt.Errorf("Failed to stop %s: %s", name, err)
			}
			req := &compute.InstancesSetServiceAccountRequest{
				Email:  emailsByInstance[name],
				Scopes: instanceScopes,
			}
			if err := gcloud.CheckOperation(is.SetServiceAccount(*project, *zone, name, req).Do()); err != nil {
				return fmt.Errorf("Failed to set scopes: %s", err)
			}
			if err := gcloud.CheckOperation(is.Start(*project, *zone, name).Do()); err != nil {
				return fmt.Errorf("Failed to start %s: %s", name, err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		sklog.Fatal(err)
	}
}
