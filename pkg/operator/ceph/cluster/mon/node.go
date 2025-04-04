/*
Copyright 2016 The Rook Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mon

import (
	"github.com/pkg/errors"
	opcontroller "github.com/rook/rook/pkg/operator/ceph/controller"
	"github.com/rook/rook/pkg/operator/k8sutil"
	v1 "k8s.io/api/core/v1"
)

const (
	monIPAnnotation = "network.rook.io/mon-ip"
)

func getNodeInfoFromNode(n v1.Node) (*opcontroller.MonScheduleInfo, error) {
	nr := &opcontroller.MonScheduleInfo{
		Name:     n.Name,
		Hostname: n.Labels[k8sutil.LabelHostname()],
	}

	// If the host networking is setup such that a different IP should be used
	// than the one that is to the K8s node.
	if customIP, ok := n.Annotations[monIPAnnotation]; ok {
		logger.Infof("found %s annotation on node %q --> %q", monIPAnnotation, n.Name, customIP)
		nr.Address = customIP
		return nr, nil
	}

	for _, ip := range n.Status.Addresses {
		if ip.Type == v1.NodeInternalIP {
			logger.Debugf("using internal IP %s for node %s", ip.Address, n.Name)
			nr.Address = ip.Address
			break
		}
	}

	// If no internal IP found try to use an external IP
	if nr.Address == "" {
		for _, ip := range n.Status.Addresses {
			if ip.Type == v1.NodeExternalIP {
				logger.Debugf("using external IP %s for node %s", ip.Address, n.Name)
				nr.Address = ip.Address
				break
			}
		}
	}

	if nr.Address == "" {
		return nil, errors.Errorf("failed to find any IP on node %s", nr.Name)
	}
	return nr, nil
}
