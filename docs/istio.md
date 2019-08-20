# SMI Metrics with Istio

The SMI Metrics Extension APIService makes use of prometheus metrics exposed by Istio to resopond to metric queries like p90, request counts, etc.

The following steps require Istio to be installed along with mixer as it generate metrics.

## Out of the Box Installation

As Istio allows metrics to be configurable, SMI with Istio can be installed by directly running

```bash
helm template chart --set adapter=istio | k apply -f -
```
This installs the necessary instances, handlers and rules for Istio to emit those metrics along with SMI-Metrics APIServer.

## How it works

SMI Metrics makes use of the following Metric Instances:

- `istio_smi_request_duration_seconds_bucket`
- `istio_smi_requests_total`

As SMI Metrics allows users to query for metrics based on the resource `kind` i.e pod, deployment, etc. These metrics should have corresponding labels that provide this information.

The labels required by SMI Metrics on the above metrics are:

- `source_owner`, `destination_owner` : To identify the workload types that these metrics belong to i.e Deployment, Daemonset, Job, etc.
- `source_uid`, `destination_uid` : To identify the pod instance that these metrics belong to.
- `source_workload_namespace`,  `destination_workload_namespace` : To identify the namespace that a particular metric belongs to.

The first four labels are not present in the default metrics installation of Istio. So, For SMI to work with Istio, SMI needs separate metrics with labels that it requires.

The manifests that add these required metrics can be found [here](https://github.com/deislabs/smi-metrics/tree/master/chart/templates/crds.yaml)

To verify if the labels are added correctly, prometheus metrics can be checked to see if those metrics have the configured labels.
