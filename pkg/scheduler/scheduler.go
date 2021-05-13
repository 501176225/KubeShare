package scheduler

import (
	"math"
	"sync"

	corev1 "k8s.io/api/core/v1"
	kubesharev1 "github.com/NTHU-LSALAB/KubeShare/pkg/apis/kubeshare/v1"
)

func scheduleSharePod(isGPUPod bool, gpu_request float64, gpu_mem int64, sharepod *kubesharev1.SharePod, nodeList []*corev1.Node, podList []*corev1.Pod, sharePodList []*kubesharev1.SharePod) (string, string) {

	// Implement custom scheduling algorithm and replace the function assignment
	// Prototype: FUNC(bool, string, *kubesharev1.SharePod, NodeResources) (string, string, error)
	ap := ScheduleAlgorithmBestFit

	nodeResources := syncClusterResources(nodeList, podList, sharePodList)
	for _, filter := range filters {
		filter(nodeResources, sharepod)
	}
	return ap(isGPUPod, gpu_request, gpu_mem, sharepod, nodeResources)
}

func ScheduleAlgorithmBestFit(isGPUPod bool, gpu_request float64, gpu_mem int64, sharepod *kubesharev1.SharePod, nodeResources NodeResources) (schedNodeName string, schedGPUID string) {
	type candidateNodeGPU struct {
		NodeName string
		GPUID    string
		Point    int64
	}

	bestNode := candidateNodeGPU{
		Point:    2147483647,
		NodeName: "",
		GPUID:    "",
	}
	var bestNodeMux sync.Mutex
	tryBestNode := func(point int64, nodeName, GPUID string) {
		bestNodeMux.Lock()
		if point < bestNode.Point {
			bestNode.Point = point
			bestNode.NodeName = nodeName
			bestNode.GPUID = GPUID
		}
		bestNodeMux.Unlock()
	}

	var cpuReqTotal, memReqTotal int64 = 0, 0
	for _, container := range sharepod.Spec.Containers {
		cpuReqTotal += container.Resources.Requests.Cpu().MilliValue()
		memReqTotal += container.Resources.Requests.Memory().MilliValue()
	}

	gpu_request_millivalue := int64(math.Ceil(gpu_request * (float64)(1000.0)))

	var wait sync.WaitGroup\

	scheduleNode := func(nodeName string, nodeRes *NodeResource) {
		//如果CPU和Memory资源不够，直接返回
		if nodeRes.CpuFree < cpuReqTotal || nodeRes.MemFree < memReqTotal {
			wait.Done()
			return
		}

		if isGPUPod {
			findTheHole := false
			//遍历结点中的每一个GPU
			for id, gpu := range nodeRes.GpuFree {
				//资源不够，跳过
				if gpu.GPUFreeReq < gpu_request_millivalue || gpu.GPUFreeMem < gpu_mem {
					continue
				}
				//找到满足的GPU
				findTheHole = true
				//分配后剩余资源越多，得分越多
				//可在此更改评分
				nodeScore := gpu.GPUFreeReq-gpu_request_millivalue
				tryBestNode(nodeScore, nodeName, id)
			}
			if !findTheHole {
				//如果有剩余的GPU，分配虚拟GPU ID
				if nodeRes.GpuFreeCount > 0 {
					nodeScore := 1000-gpu_request_millivalue
					tryBestNode(nodeScore, nodeName, kubesharev1.NewGPUID(5))
				}

				//
			}
		} else {
			//不需要GPU，按照CPU排序
			nodeScore := nodeRes.CpuFree-cpuReqTotal
			tryBestNode(nodeRes.CpuFree-cpuReqTotal, nodeName, "")
		}
		wait.Done()
	}
	wait.Add(len(nodeResources))

	for nodeName, nodeRes := range nodeResources {
		go scheduleNode(nodeName, nodeRes)
	}
	wait.Wait()

	return bestNode.NodeName, bestNode.GPUID
}
