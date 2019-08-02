package istio

import (
	"errors"
	"strings"

	"github.com/prometheus/common/log"

	"github.com/prometheus/common/model"
)

var (
	sourceOwner          = model.LabelName("source_owner")
	destinationOwner     = model.LabelName("destination_owner")
	sourceNamespace      = model.LabelName("source_workload_namespace")
	destinationNamespace = model.LabelName("destination_workload_namespace")
	sourcePod            = model.LabelName("source_uid")
	destinationPod       = model.LabelName("destination_uid")
)

type ResultType int

const (
	Pod       ResultType = 1
	Workload  ResultType = 2
	Namespace ResultType = 3
)

type Result struct {
	Name      string
	Namespace string
	Kind      string
}

func GetType(labels model.Metric) (ResultType, error) {
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
	return -1, errors.New("couldn't find type of result")
}

// Return (src, dst) from the result labels
func NewResult(labels model.Metric) (source, destination *Result, err error) {
	resultType, err := GetType(labels)
	if err != nil {
		log.Error(err)
		return nil, nil, err
	}
	var src, dst *Result

	switch resultType {
	case Pod:
		// Resource is Pod
		if val, ok := labels[sourcePod]; ok {
			// Present at Source
			values := strings.Split(string(val), "//")
			subVal := strings.Split(values[1], ".")
			src = &Result{
				Name:      subVal[0],
				Namespace: subVal[1],
				Kind:      "Pod",
			}
		}

		if val, ok := labels[destinationPod]; ok {
			// Present at Destination
			values := strings.Split(string(val), "//")
			subVal := strings.Split(values[1], ".")
			dst = &Result{
				Name:      subVal[0],
				Namespace: subVal[1],
				Kind:      "Pod",
			}
		}
	case Namespace:
		// It is Namespace
		if val, ok := labels[sourceNamespace]; ok {
			// Present at Source
			src = &Result{
				Name:      string(val),
				Namespace: "",
				Kind:      "Namespace",
			}
		}

		if val, ok := labels[destinationNamespace]; ok {
			// Present at  Destination
			dst = &Result{
				Name:      string(val),
				Namespace: "",
				Kind:      "Namespace",
			}
		}
	case Workload:
		if val, ok := labels[sourceOwner]; ok && !strings.Contains(string(val), ".+") {
			// Present at Source
			values := strings.Split(string(val), "/")[6:]
			if len(values) < 3 {
				return nil, nil, errors.New("wrong Pattern of Workload")
			}
			src = &Result{
				Name:      values[2],
				Namespace: values[0],
				Kind:      values[1],
			}
		}

		if val, ok := labels[destinationOwner]; ok && !strings.Contains(string(val), ".+") {
			// Present at  Destination
			values := strings.Split(string(val), "/")[6:]
			if len(values) < 3 {
				return nil, nil, errors.New("wrong Pattern of Workload")
			}
			dst = &Result{
				Name:      values[2],
				Namespace: values[0],
				Kind:      values[1],
			}
		}
	}
	return src, dst, nil
}
