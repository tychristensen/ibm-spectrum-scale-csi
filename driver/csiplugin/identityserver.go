/**
 * Copyright 2019 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package scale

import (
	"context"
	"io/ioutil"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ScaleIdentityServer struct {
	Driver *ScaleDriver
}

func (is *ScaleIdentityServer) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}, nil
}

func (is *ScaleIdentityServer) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	glog.V(4).Infof("Probe called with args: %#v", req)

	// Determine plugin health
	// If unhealthy return gRPC error code
	// more on error codes https://github.com/container-storage-interface/spec/blob/master/spec.md#probe-errors

	// Check /proc/mounts if gpfs filesystem is listed
	procMounts, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		glog.Error(err)
		return &csi.ProbeResponse{Ready: &wrappers.BoolValue{Value: false}},
			status.Error(codes.FailedPrecondition, "Probe was unable to open /proc/mounts")
	}
	lines := strings.Split(string(procMounts), "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[2] == "gpfs" {
			// If gpfs is listed return plugin is ready
			return &csi.ProbeResponse{Ready: &wrappers.BoolValue{Value: true}}, nil
		}
	}

	// If gpfs is not found return plugin isn't ready with gRPC error
	return &csi.ProbeResponse{Ready: &wrappers.BoolValue{Value: false}},
		status.Error(codes.FailedPrecondition, "GPFS is not listed in /proc/mounts")
}

func (is *ScaleIdentityServer) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	glog.V(5).Infof("Using default GetPluginInfo")

	if is.Driver.name == "" {
		return nil, status.Error(codes.Unavailable, "Driver name not configured")
	}

	return &csi.GetPluginInfoResponse{
		Name:          is.Driver.name,
		VendorVersion: is.Driver.vendorVersion,
	}, nil
}
