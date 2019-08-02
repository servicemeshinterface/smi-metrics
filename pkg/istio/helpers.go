package istio

import (
	"errors"
	"strings"

	"github.com/prometheus/common/log"

	"github.com/prometheus/common/model"
)

type ResultType int

const (
	Pod       ResultType = 1
	Workload  ResultType = 2
	Namespace ResultType = 3
)

type result struct {
	Name      string
	Namespace string
	Kind      string
}

var sourceOwner = model.LabelName("source_owner")
var destinationOwner = model.LabelName("destination_owner")
var sourceNamespace = model.LabelName("source_workload_namespace")
var destinationNamespace = model.LabelName("destination_workload_namespace")
var sourcePod = model.LabelName("source_uid")
var destinationPod = model.LabelName("destination_uid")

func GetType(labels model.Metric) (ResultType, error) {
	metric := labels.String()
	if strings.Contains(metric, string(sourcePod)) || strings.Contains(metric, string(destinationPod)) {
		return Pod, nil
	} else if strings.Contains(metric, string(sourceOwner)) || strings.Contains(metric, string(destinationOwner)) {
		return Workload, nil
	} else if strings.Contains(metric, string(sourceNamespace)) || strings.Contains(metric, string(destinationNamespace)) {
		return Namespace, nil
	}
	return -1, errors.New("Couldn't find type of result")
}

// Return (src, dst) from the result labels
func NewResult(labels model.Metric) (*result, *result, error) {
	resultType, err := GetType(labels)
	if err != nil {
		log.Error(err)
		return nil, nil, err
	}
	var src, dst *result
	if resultType == Workload {
		if val, ok := labels[sourceOwner]; ok && !strings.Contains(string(val), ".+") {
			//Present at Source
			values := strings.Split(string(val), "/")[6:]
			if len(values) < 3 {
				return nil, nil, errors.New("Wrong Pattern of Workload")
			}
			src = &result{
				Name:      values[2],
				Namespace: values[0],
				Kind:      values[1],
			}
		}

		if val, ok := labels[destinationOwner]; ok && !strings.Contains(string(val), ".+") {
			//Present at  Destination
			values := strings.Split(string(val), "/")[6:]
			if len(values) < 3 {
				return nil, nil, errors.New("Wrong Pattern of Workload")
			}
			dst = &result{
				Name:      values[2],
				Namespace: values[0],
				Kind:      values[1],
			}
		}

	} else if resultType == Namespace {
		// It is Namespace
		if val, ok := labels[sourceNamespace]; ok {
			//Present at Source
			src = &result{
				Name:      string(val),
				Namespace: "",
				Kind:      "Namespace",
			}
		}

		if val, ok := labels[destinationNamespace]; ok {
			//Present at  Destination
			dst = &result{
				Name:      string(val),
				Namespace: "",
				Kind:      "Namespace",
			}
		}
	} else if resultType == Pod {
		// Resource is Pod
		log.Info("PODDDDDD")
		if val, ok := labels[sourcePod]; ok {
			//Present at Source
			values := strings.Split(string(val), "//")
			log.Info("Value is", values[1])
			subVal := strings.Split(values[1], ".")
			log.Info("Sub Values", subVal[0], subVal[1])
			src = &result{
				Name:      subVal[0],
				Namespace: subVal[1],
				Kind:      "Pod",
			}
		}

		if val, ok := labels[destinationPod]; ok {
			//Present at Destination
			values := strings.Split(string(val), "//")
			subVal := strings.Split(values[1], ".")
			log.Info("Value is", values[1])
			log.Info("Sub Values", subVal[0], subVal[1])
			dst = &result{
				Name:      subVal[0],
				Namespace: subVal[1],
				Kind:      "Pod",
			}
		}
	}
	return src, dst, nil
}
