package cluster

import (
	"errors"
	"sort"
	"strings"

	"github.com/accuknox/auto-policy-discovery/src/config"
	"github.com/accuknox/auto-policy-discovery/src/types"
)

func GetPods(clusterName string) []types.Pod {
	var pods []types.Pod

	if config.GetCfgClusterInfoFrom() == "k8sclient" { // get from k8s client api
		pods = GetPodsFromK8sClient()
	} else {
		clusterInstance := GetClusterFromClusterName(clusterName)
		if clusterInstance.ClusterID == 0 { // cluster not onboarded
			pods = nil
		} else {
			pods = GetPodsFromCluster(clusterInstance)
		}
	}

	// Append VM pod type to pods
	pods = append(pods, types.Pod{
		Namespace: types.PolicyDiscoveryVMNamespace,
		PodName:   types.PolicyDiscoveryVMPodName,
	})

	pods = append(pods, types.Pod{
		Namespace: types.PolicyDiscoveryContainerNamespace,
		PodName:   types.PolicyDiscoveryContainerPodName,
	})

	return pods
}

func GetAllClusterResources(cluster string) ([]string, []types.Service, []types.Endpoint, []types.Pod, error) {
	clusterMgmt := config.GetCfgClusterInfoFrom()

	if clusterMgmt == "k8sclient" { // get from k8s client api
		namespaces := GetNamespacesFromK8sClient()
		services := GetServicesFromK8sClient()
		endpoints := GetEndpointsFromK8sClient()
		pods := GetPodsFromK8sClient()

		return namespaces, services, endpoints, pods, nil
	} else if clusterMgmt == "kvmservice" {
		namespaces, pods := GetResourcesFromKvmService()
		return namespaces, nil, nil, pods, nil
	} else {
		clusterInstance := GetClusterFromClusterName(cluster)
		if clusterInstance.ClusterID == 0 { // cluster not onboarded
			return nil, nil, nil, nil, errors.New("Cluster " + cluster + " not onboarded")
		}

		namespaces := GetNamespacesFromCluster(clusterInstance)
		services := GetServicesFromCluster(clusterInstance)
		endpoints := GetEndpointsFromCluster(clusterInstance)
		pods := GetPodsFromCluster(clusterInstance)

		return namespaces, services, endpoints, pods, nil
	}
}

// ExtractPodSvcInfoFromIP -- Extract respective podname/ns/labels from pod/svc ip
func ExtractPodSvcInfoFromIP(ip, clustername string) (string, string, string) {
	podSvcName := ip

	_, services, _, pods, err := GetAllClusterResources(clustername)
	if err != nil {
		return podSvcName, "", ""
	}

	for _, pod := range pods {
		if pod.PodIP == ip {
			return "pod/" + pod.PodName, strings.Join(sort.StringSlice(pod.Labels), ","), pod.Namespace
		}
	}
	for _, svc := range services {
		if svc.ClusterIP == ip {
			return "svc/" + svc.ServiceName, strings.Join(svc.Labels, ","), svc.Namespace
		}
	}

	return podSvcName, "", ""
}
