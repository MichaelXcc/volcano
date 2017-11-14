/*
Copyright 2017 The Kubernetes Authors.

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

package proportion

import (
	"github.com/golang/glog"

	apiv1 "github.com/kubernetes-incubator/kube-arbitrator/pkg/apis/v1"
	"github.com/kubernetes-incubator/kube-arbitrator/pkg/schedulercache"

	"k8s.io/apimachinery/pkg/api/resource"
)

// PolicyName is the name of proportion policy; it'll be use for any case
// that need a name, e.g. default policy, register proportion policy.
var PolicyName = "proportion"

type proportionScheduler struct {
}

func New() *proportionScheduler {
	return &proportionScheduler{}
}

func (ps *proportionScheduler) Name() string {
	return PolicyName
}

func (ps *proportionScheduler) Initialize() {
	// TODO
}

func (ps *proportionScheduler) Group(
	jobs []*schedulercache.QueueInfo,
) map[string][]*schedulercache.QueueInfo {
	groups := make(map[string][]*schedulercache.QueueInfo)
	for _, job := range jobs {
		groups[job.Queue().Namespace] = append(groups[job.Queue().Namespace], job)
	}

	return groups
}

func (ps *proportionScheduler) Allocate(
	jobGroup map[string][]*schedulercache.QueueInfo,
	nodes []*schedulercache.NodeInfo,
) map[string]*schedulercache.QueueInfo {
	totalCPU := int64(0)
	totalMEM := int64(0)
	for _, node := range nodes {
		if cpu, ok := node.Node().Status.Capacity["cpu"]; ok {
			if capacity, ok := cpu.AsInt64(); ok {
				totalCPU += capacity
			}
		}
		if memory, ok := node.Node().Status.Capacity["memory"]; ok {
			if capacity, ok := memory.AsInt64(); ok {
				totalMEM += capacity
			}
		}
	}
	totalWeight := int64(0)
	for _, jobs := range jobGroup {
		for _, job := range jobs {
			totalWeight += int64(job.Queue().Spec.Weight)
		}
	}
	glog.V(4).Infof("proportion scheduler, total cpu %d, total memory %d, total weight %d", totalCPU, totalMEM, totalWeight)

	if totalCPU == 0 || totalMEM == 0 || totalWeight == 0 {
		glog.V(4).Info("there is no resources or allocators in cluster")
		return nil
	}

	totalAllocatedCPU := int64(0)
	totalAllocatedMEM := int64(0)
	allocatedResult := make(map[string]*schedulercache.QueueInfo)
	for _, jobs := range jobGroup {
		for _, job := range jobs {
			allocatedCPU := int64(job.Queue().Spec.Weight) * totalCPU / totalWeight
			allocatedMEM := int64(job.Queue().Spec.Weight) * totalMEM / totalWeight
			totalAllocatedCPU += allocatedCPU
			totalAllocatedMEM += allocatedMEM

			allocatedResult[job.Name()] = job.Clone()
			allocatedResult[job.Name()].Queue().Status.Deserved = apiv1.ResourceList{
				Resources: map[apiv1.ResourceName]resource.Quantity{
					"cpu":    *resource.NewQuantity(allocatedCPU, resource.DecimalSI),
					"memory": *resource.NewQuantity(allocatedMEM, resource.BinarySI),
				},
			}
			// clear Used resources
			allocatedResult[job.Name()].Queue().Status.Used = apiv1.ResourceList{
				Resources: make(map[apiv1.ResourceName]resource.Quantity),
			}
		}
	}

	// assign the left resources to queues one by one
	// leftCPU and leftMEM is less than the size of allocatedResult
	// so all of them can be allocated in below
	leftCPU := totalCPU - totalAllocatedCPU
	leftMEM := totalMEM - totalAllocatedMEM
	for _, queue := range allocatedResult {
		resList := queue.Queue().Status.Deserved.DeepCopy()
		if leftCPU > 0 {
			leftCPU -= 1
			result := resList.Resources["cpu"].DeepCopy()
			result.Add(*resource.NewQuantity(1, resource.DecimalSI))
			queue.Queue().Status.Deserved.Resources["cpu"] = result
		}
		if leftMEM > 0 {
			leftMEM -= 1
			result := resList.Resources["memory"].DeepCopy()
			result.Add(*resource.NewQuantity(1, resource.BinarySI))
			queue.Queue().Status.Deserved.Resources["memory"] = result
		}
	}

	return allocatedResult
}

func (ps *proportionScheduler) Assign(
	jobs []*schedulercache.QueueInfo,
	alloc *schedulercache.QueueInfo,
) *schedulercache.Resource {
	// TODO
	return nil
}

func (ps *proportionScheduler) Polish(
	job *schedulercache.QueueInfo,
	res *schedulercache.Resource,
) []*schedulercache.QueueInfo {
	// TODO
	return nil
}

func (ps *proportionScheduler) UnInitialize() {
	// TODO
}
