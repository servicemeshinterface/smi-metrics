package istio

import (
	"errors"
	"fmt"
	"strings"

	"github.com/prometheus/common/log"
	"github.com/prometheus/common/model"
	v1 "k8s.io/api/core/v1"
)

var (
	sourceOwner          = model.LabelName("source_owner")
	destinationOwner     = model.LabelName("destination_owner")
	sourceNamespace      = model.LabelName("source_workload_namespace")
	destinationNamespace = model.LabelName("destination_workload_namespace")
	sourcePod            = model.LabelName("source_uid")
	destinationPod       = model.LabelName("destination_uid")
)

const (
	Pod       string = "Pod"
	Workload  string = "Workload"
	Namespace string = "Namespace"
)

func newObjectReference(name, namespace, kind string) *v1.ObjectReference {
	return &v1.ObjectReference{
		Name:      name,
		Namespace: namespace,
		Kind:      strings.Title(strings.TrimSuffix(kind, "s")),
	}
}

func GetType(labels fmt.Stringer) (string, error) {
	metric := labels.String()
	if strings.Contains(metric, string(sourcePod)) || strings.Contains(metric, string(destinationPod)) {
		return Pod, nil
	}
	if strings.Contains(metric, string(sourceOwner)) || strings.Contains(metric, string(destinationOwner)) {
		return Workload, nil
	}
	if strings.Contains(metric, string(sourceNamespace)) || strings.Contains(metric, string(destinationNamespace)) {
		return Namespace, nil
	}
	return "", errors.New("couldn't find type of result")
}

// Return (src, dst) from the result labels
func GetObjectsReference(labels model.Metric) (source, destination *v1.ObjectReference, err error) {
	result, err := GetType(labels)
	if err != nil {
		log.Error(err)
		return nil, nil, err
	}
	var src, dst *v1.ObjectReference

	switch result {
	case Pod:
		// Resource is a Pod
		src, dst, err = objectReferencesFromPodLabels(labels)
	case Namespace:
		// Resource is a Namespace
		src, dst, err = objectReferencesFromNamespaceLabels(labels)
	case Workload:
		// Resource is a Workload i.e Deployment, Daemonset, Job, etc
		src, dst, err = objectReferencesFromWorkloadLabels(labels)
	}

	if err != nil {
		return nil, nil, err
	}
	return src, dst, nil
}

func objectReferencesFromNamespaceLabels(labels model.Metric) (source, destination *v1.ObjectReference, err error) {
	var src, dst *v1.ObjectReference
	if val, ok := labels[sourceNamespace]; ok {
		// Present at Source
		src = newObjectReference(string(val), "", Namespace)
	}

	if val, ok := labels[destinationNamespace]; ok {
		// Present at  Destination
		dst = newObjectReference(string(val), "", Namespace)
	}

	return src, dst, nil
}

func objectReferencesFromWorkloadLabels(labels model.Metric) (source, destination *v1.ObjectReference, err error) {
	var src, dst *v1.ObjectReference
	if val, ok := labels[sourceOwner]; ok && !strings.Contains(string(val), ".+") {
		// Present at Source
		src, err = objectReferenceFromWorkloadLabel(val)
		if err != nil {
			return src, dst, err
		}
	}

	if val, ok := labels[destinationOwner]; ok && !strings.Contains(string(val), ".+") {
		// Present at  Destination
		dst, err = objectReferenceFromWorkloadLabel(val)
		if err != nil {
			return src, dst, err
		}
	}

	return src, dst, nil
}

func objectReferencesFromPodLabels(labels model.Metric) (source, destination *v1.ObjectReference, err error) {
	var src, dst *v1.ObjectReference
	if val, ok := labels[sourcePod]; ok {
		// Present at Source
		src, err = ObjectReferenceFromPodLabel(val)
		if err != nil {
			return nil, nil, err
		}
	}

	if val, ok := labels[destinationPod]; ok {
		// Present at Destination
		dst, err = ObjectReferenceFromPodLabel(val)
		if err != nil {
			return nil, nil, err
		}
	}

	return src, dst, nil
}

func ObjectReferenceFromPodLabel(value model.LabelValue) (*v1.ObjectReference, error) {
	values := strings.Split(string(value), "//")
	subVal := strings.Split(values[1], ".")
	if len(subVal) < 2 {
		return nil, errors.New("wrong pattern of Pod label value")
	}

	return newObjectReference(subVal[0], subVal[1], Pod), nil
}

func objectReferenceFromWorkloadLabel(value model.LabelValue) (*v1.ObjectReference, error) {
	values := strings.Split(string(value), "/")[6:]
	if len(values) < 3 {
		return nil, errors.New("wrong pattern of Workload label value")
	}

	return newObjectReference(values[2], values[0], values[1]), nil
}
