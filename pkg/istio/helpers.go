package istio

import (
	"errors"
	"regexp"
	"strings"

	"github.com/prometheus/common/log"

	"github.com/prometheus/common/model"
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

func IfWorkload(labels model.Metric) (bool, error) {

	for k, _ := range labels {
		src, _ := regexp.Match(".+_owner", []byte(k))
		dst, _ := regexp.Match(".+_owner", []byte(k))
		if src || dst {
			return true, nil
		}

		src, _ = regexp.Match(".+_namespace", []byte(k))
		dst, _ = regexp.Match(".+_namespace", []byte(k))
		if src || dst {
			return false, nil
		}
	}

	return false, errors.New("Couldn't find type of result")
}

// Return (src, dst) from the result labels
func NewResult(labels model.Metric) (*result, *result, error) {
	workload, err := IfWorkload(labels)
	if err != nil {
		log.Error(err)
		return nil, nil, err
	}
	var src, dst *result
	if workload {
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

	} else {
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

	}
	return src, dst, nil
}
