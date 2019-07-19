package mesh

import (
	"github.com/deislabs/smi-sdk-go/pkg/apis/metrics"
	v1 "k8s.io/api/core/v1"
)

type ResourceDetails struct {
	Kind       string
	Namespaced bool
}

func GetResourceDetails(name string) (*ResourceDetails, bool) {
	for kind, r := range metrics.AvailableKinds {
		if r.Name == name {
			return &ResourceDetails{
				Kind:       kind,
				Namespaced: r.Namespaced,
			}, true
		}
	}

	return nil, false
}

// listKey constructs a key that is useful for getting elements from a
// TrafficMetricsList
func ListKey(kind, name, namespace string) *v1.ObjectReference {
	var namespaced bool

	if details, ok := metrics.AvailableKinds[kind]; ok {
		namespaced = details.Namespaced
	} else {
		namespaced = true
	}

	obj := &v1.ObjectReference{
		Kind: kind,
		Name: name,
	}

	if namespaced {
		obj.Namespace = namespace
	}

	return obj
}
