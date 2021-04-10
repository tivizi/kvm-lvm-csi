package pkg

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (driver *Driver) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	// Check arguments
	if len(req.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name missing in request")
	}

	driver.mutex.Lock()
	defer driver.mutex.Unlock()
	if volume, err := driver.GetVolume(req.GetName()); err == nil {
		return &csi.CreateVolumeResponse{
			Volume: volume,
		}, nil
	}
	if volume, err := driver.NewVolume(req.GetName()); err == nil {
		return &csi.CreateVolumeResponse{
			Volume: volume,
		}, nil
	}

	return nil, status.Error(codes.Unavailable, "Create volume error")

}

func (driver *Driver) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	driver.mutex.Lock()
	defer driver.mutex.Unlock()
	driver.DelVolume(req.GetVolumeId())
	return &csi.DeleteVolumeResponse{}, nil
}

func (driver *Driver) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	var cl []csi.ControllerServiceCapability_RPC_Type
	cl = []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_GET_VOLUME,
		csi.ControllerServiceCapability_RPC_GET_CAPACITY,
		csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
		csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
		csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
		csi.ControllerServiceCapability_RPC_VOLUME_CONDITION,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
	}
	var csc []*csi.ControllerServiceCapability

	for _, cap := range cl {
		csc = append(csc, &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		})
	}

	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: csc,
	}, nil
}

func (driver *Driver) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeContext:      req.GetVolumeContext(),
			VolumeCapabilities: req.GetVolumeCapabilities(),
			Parameters:         req.GetParameters(),
		},
	}, nil
}

func (driver *Driver) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	err := driver.AttachDisk(req.VolumeId, req.NodeId)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	return &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{},
	}, nil
}

func (driver *Driver) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	err := driver.DetachDisk(req.VolumeId, req.NodeId)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	RemoveMeta(req.VolumeId)
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

func (driver *Driver) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return &csi.GetCapacityResponse{
		AvailableCapacity: 0xffffffff,
	}, nil
}

func (driver *Driver) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return &csi.ListVolumesResponse{
		Entries: []*csi.ListVolumesResponse_Entry{
			&csi.ListVolumesResponse_Entry{
				Volume: &csi.Volume{
					VolumeId:      "test-volume",
					CapacityBytes: 0xffffffff,
				},
				Status: &csi.ListVolumesResponse_VolumeStatus{
					PublishedNodeIds: []string{"knode1"},
					VolumeCondition: &csi.VolumeCondition{
						Abnormal: false,
						Message:  "it's ok",
					},
				},
			},
		},
	}, nil
}

func (driver *Driver) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return &csi.ControllerGetVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      "test-volume",
			CapacityBytes: 0xffffffff,
		},
		Status: &csi.ControllerGetVolumeResponse_VolumeStatus{
			PublishedNodeIds: []string{"knode1"},
			VolumeCondition: &csi.VolumeCondition{
				Abnormal: false,
				Message:  "it's ok",
			},
		},
	}, nil
}

// CreateSnapshot uses tar command to create snapshot for hostpath volume. The tar command can quickly create
// archives of entire directories. The host image must have "tar" binaries in /bin, /usr/sbin, or /usr/bin.
func (driver *Driver) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unavailable, "unsupported")
}

func (driver *Driver) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unavailable, "unsupported")
}

func (driver *Driver) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unavailable, "unsupported")
}

func (driver *Driver) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unavailable, "unsupported")
}
