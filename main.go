package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dougm/pretty"

	"github.com/vmware/govmomi/find"
	//"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"

	"github.com/vmware/govmomi"
	cnstypes "github.com/vmware/govmomi/cns/types"


	"github.com/vmware/govmomi/cns"

)

func main() {

	datastoreForMigration := os.Getenv("TARGET_DATASTORE")
	datacenter := os.Getenv("DC")
	volumeId := os.Getenv("VOLUME_ID")

	url := os.Getenv("CNS_VC_URL") // example: export CNS_VC_URL='https://username:password@vc-ip/sdk'

	u, err := soap.ParseURL(url)

	ctx := context.Background()
	c, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		panic(err)
	}
	// UseServiceVersion sets soap.Client.Version to the current version of the service endpoint via /sdk/vsanServiceVersions.xml
	c.UseServiceVersion("vsan")
	cnsClient, err := cns.NewClient(ctx, c.Client)
	if err != nil {
		panic(err)
	}

	//finder := find.NewFinder(cnsClient.vim25Client, false)
	finder := find.NewFinder(c.Client, false)
	dc, err := finder.Datacenter(ctx, datacenter)
	if err != nil {
		panic(err)
	}

	finder.SetDatacenter(dc)

	fmt.Printf("version: %v+", cnsClient.Version)

	if cnsClient.Version != cns.ReleaseVSAN67u3 && cnsClient.Version != cns.ReleaseVSAN70 && datastoreForMigration != "" {
		migrationDS, err := finder.Datastore(ctx, datastoreForMigration)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Relocating volume %v to datastore %+v", pretty.Sprint(volumeId), migrationDS.Reference())
		relocateSpec := cnstypes.NewCnsBlockVolumeRelocateSpec(volumeId, migrationDS.Reference())
		relocateTask, err := cnsClient.RelocateVolume(ctx, relocateSpec)
		if err != nil {
			fmt.Printf("Failed to migrate volume with Relocate API. Error: %+v \n", err)
			panic(err)
		}
		relocateTaskInfo, err := cns.GetTaskInfo(ctx, relocateTask)
		if err != nil {
			fmt.Printf("Failed to get info of task returned by Relocate API. Error: %+v \n", err)
			panic(err)
		}
		taskResults, err := cns.GetTaskResultArray(ctx, relocateTaskInfo)
		if err != nil {
			panic(err)
		}
		for _, taskResult := range taskResults {
			res := taskResult.GetCnsVolumeOperationResult()
			if res.Fault != nil {
				fmt.Printf("Relocation failed due to fault: %+v", res.Fault)
			}
			fmt.Printf("Successfully Relocated volume. Relocate task info result: %+v", pretty.Sprint(taskResult))
		}
	}

}
