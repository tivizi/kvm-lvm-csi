package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sync"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
)

// Driver Driver
type Driver struct {
	name              string
	version           string
	nodeID            string
	maxVolumesPerNode int64
	mutex             sync.Mutex
}

func NewDriver(nodeId string) (*Driver, error) {
	return &Driver{
		name:              "kvm-lvm-csi",
		version:           "1.0.0",
		nodeID:            nodeId,
		maxVolumesPerNode: 1000,
	}, nil
}

func (driver *Driver) GetVolume(name string) (*csi.Volume, error) {
	fmt.Println("GetVolume:", name)
	cmdPath, err := exec.LookPath("lvs")
	if err != nil {
		return nil, fmt.Errorf("findmnt not found: %w", err)
	}

	out, err := exec.Command(cmdPath, "--reportformat", "json").CombinedOutput()
	if err != nil {
		glog.V(3).Infof("failed to execute command: %+v", cmdPath)
		return nil, err
	}
	var result map[string]interface{}
	json.Unmarshal(out, &result)
	volList := result["report"].([]interface{})[0].(map[string]interface{})["lv"].([]interface{})
	for _, vol := range volList {
		lvName := vol.(map[string]interface{})["lv_name"].(string)
		if lvName == "k8s-"+name {
			return &csi.Volume{
				VolumeId: lvName,
			}, nil
		}
	}
	return nil, errors.New("not found")
}

func (driver *Driver) NewVolume(name string) (*csi.Volume, error) {
	fmt.Println("NewVolume", name)
	cmdPath, err := exec.LookPath("lvcreate")
	if err != nil {
		return nil, fmt.Errorf("findmnt not found: %w", err)
	}

	lvName := name
	_, err = exec.Command(cmdPath, "storages", "-n", lvName, "-L", "10G").CombinedOutput()
	if err != nil {
		glog.V(3).Infof("failed to execute command: %+v", cmdPath)
		return nil, err
	}

	return &csi.Volume{
		VolumeId: lvName,
	}, nil
}

func (driver *Driver) DelVolume(volumeId string) error {
	fmt.Println("DelVolume:", volumeId)
	cmdPath, err := exec.LookPath("lvremove")
	if err != nil {
		return fmt.Errorf("findmnt not found: %w", err)
	}

	_, err = exec.Command(cmdPath, "/dev/storages/"+volumeId, "-y").CombinedOutput()
	if err != nil {
		glog.V(3).Infof("failed to execute command: %+v", cmdPath)
		return err
	}
	return nil
}

func (driver *Driver) AttachDisk(volumeId, nodeId string) error {
	fmt.Println("AttachDisk:", volumeId, nodeId)
	cmdPath, err := exec.LookPath("virsh")
	if err != nil {
		return fmt.Errorf("virsh not found: %w", err)
	}

	meta, err := NewMeta(volumeId, nodeId)
	glog.Infof("VolumeMeta: %s", meta)
	if err != nil {
		return err
	}

	_, err = exec.Command(cmdPath, "attach-disk", nodeId, "/dev/storages/"+volumeId, meta.Name).CombinedOutput()
	if err != nil {
		glog.V(3).Infof("failed to execute command: %+v", cmdPath)
		return err
	}
	return nil
}

func (driver *Driver) DetachDisk(volumeId, nodeId string) error {
	fmt.Println("DetachDisk: ", volumeId, nodeId)
	cmdPath, err := exec.LookPath("virsh")
	if err != nil {
		return fmt.Errorf("virsh not found: %w", err)
	}

	_, err = exec.Command(cmdPath, "detach-disk", nodeId, "/dev/storages/"+volumeId).CombinedOutput()
	if err != nil {
		glog.V(3).Infof("failed to execute command: %+v", cmdPath)
		return err
	}
	return nil
}
