# SMI Metrics API

The `smi-metrics` api follows the format of the official [Metrics API](https://github.com/kubernetes-incubator/metrics-server) being built on top of kubernetes to get metrics about pods and nodes but here the metrics are oriented towards the inbound and outbound requests i.e their Golden Metrics like p99_latency, P55_latency, success_count, etc for a particular workload.

## Working

The SMI metrics api is a Kubernetes [APIService](https://kubernetes.io/docs/tasks/access-kubernetes-api/setup-extension-api-server/) more info [here](https://github.com/deislabs/smi-metrics/blob/94dec57fdabc680cc60e4961db1707609f6b81ed/chart/templates/apiservice.yaml#L5), which is a way of extending the Kubernetes API.

We will perform installation of the SMI Metrics API w.r.t linkerd. Make sure linkerd is installed and is running as per the instructions [here](https://linkerd.io/2/getting-started/), This API can be installed by running the following command
```
helm template chart -f dev.yaml -f linkerd.yaml --name dev
```

The APIService first informs the kubernetes API about itself and the resource types that it exposes. The following command can used to see the resource types that SMI exposes.

```
kubectl api-resources | grep smi
NAME                              SHORTNAMES   APIGROUP                       NAMESPACED   KIND
daemonsets                                     metrics.smi-spec.io            true         TrafficMetrics
deployments                                    metrics.smi-spec.io            true         TrafficMetrics
namespaces                                     metrics.smi-spec.io            false        TrafficMetrics
pods                                           metrics.smi-spec.io            true         TrafficMetrics
statefulsets                                   metrics.smi-spec.io            true         TrafficMetrics
```

This means that, the APIService can be queried regarding the above mentioned resource types.

Now that the SMI APIService is installing, metric queries can be done through the kubernetes API.
To Access the Kubernetes API, we can use the `proxy` command present in `kubectl`
```
kubectl proxy --port=8087 &
```
Now the Kubernetes API can be accessed at port `8087`. As Linkerd also attaches proxies to the linkerd control-plane components, we should be able to query their metrics too. 

Be sure to access the linkerd dashboard by using the command `linkerd dashboard` before querying, as the metrics are configured to only return the last 30s by default.

```
curl http://localhost:8087/apis/metrics.smi-spec.io/v1alpha1/namespaces/linkerd/deployments/linkerd-web | jq
```
Output:
```
{
  "kind": "TrafficMetrics",
  "apiVersion": "metrics.smi-spec.io/v1alpha1",
  "metadata": {
    "name": "linkerd-web",
    "namespace": "linkerd",
    "selfLink": "/apis/metrics.smi-spec.io/v1alpha1/namespaces/linkerd/deployments/linkerd-web",
    "creationTimestamp": "2019-06-17T14:26:22Z"
  },
  "timestamp": "2019-06-17T14:26:22Z",
  "window": "30s",
  "resource": {
    "kind": "Deployment",
    "namespace": "linkerd",
    "name": "linkerd-web"
  },
  "edge": {
    "direction": "from",
    "resource": null
  },
  "metrics": [
    {
      "name": "p99_response_latency",
      "unit": "ms",
      "value": "296875m"
    },
    {
      "name": "p90_response_latency",
      "unit": "ms",
      "value": "268750m"
    },
    {
      "name": "p50_response_latency",
      "unit": "ms",
      "value": "162500m"
    },
    {
      "name": "success_count",
      "value": "73492m"
    },
    {
      "name": "failure_count",
      "value": "0"
    }
  ]
}
```
As we can see, the golden metrics of a particular resource can be retrieved by querying the API with a path format `/apis/metrics.smi-spec.io/v1alpha1/namespaces/{Namespace}/{Kind}/{ResourceName}`

Queries for golden metrics on edges i.e paths associated with a particular resource, for example linkerd-controller can also be done by adding `/edges` to the path.
```
curl http://localhost:8087/apis/metrics.smi-spec.io/v1alpha1/namespaces/linkerd/deployments/linkerd-controller/edges | jq
```
Output:
```
{
  "kind": "TrafficMetricsList",
  "apiVersion": "metrics.smi-spec.io/v1alpha1",
  "metadata": {
    "selfLink": "/apis/metrics.smi-spec.io/v1alpha1/namespaces/linkerd/deployments/linkerd-controller/edges"
  },
  "resource": {
    "kind": "Deployment",
    "namespace": "linkerd",
    "name": "linkerd-controller"
  },
  "items": [
    {
      "kind": "TrafficMetrics",
      "apiVersion": "metrics.smi-spec.io/v1alpha1",
      "metadata": {
        "name": "linkerd-controller",
        "namespace": "linkerd",
        "selfLink": "/apis/metrics.smi-spec.io/v1alpha1/namespaces/linkerd/deployments/linkerd-controller/edges",
        "creationTimestamp": "2019-06-17T14:51:57Z"
      },
      "timestamp": "2019-06-17T14:51:57Z",
      "window": "30s",
      "resource": {
        "kind": "Deployment",
        "namespace": "linkerd",
        "name": "linkerd-controller"
      },
      "edge": {
        "direction": "from",
        "resource": {
          "kind": "Deployment",
          "namespace": "linkerd",
          "name": "linkerd-web"
        }
      },
      "metrics": [
        {
          "name": "p99_response_latency",
          "unit": "ms",
          "value": "294"
        },
        {
          "name": "p90_response_latency",
          "unit": "ms",
          "value": "240"
        },
        {
          "name": "p50_response_latency",
          "unit": "ms",
          "value": "150"
        },
        {
          "name": "success_count",
          "value": "28580m"
        },
        {
          "name": "failure_count",
          "value": "0"
        }
      ]
    },
    {
      "kind": "TrafficMetrics",
      "apiVersion": "metrics.smi-spec.io/v1alpha1",
      "metadata": {
        "name": "linkerd-controller",
        "namespace": "linkerd",
        "selfLink": "/apis/metrics.smi-spec.io/v1alpha1/namespaces/linkerd/deployments/linkerd-controller/edges",
        "creationTimestamp": "2019-06-17T14:51:58Z"
      },
      "timestamp": "2019-06-17T14:51:57Z",
      "window": "30s",
      "resource": {
        "kind": "Deployment",
        "namespace": "linkerd",
        "name": "linkerd-controller"
      },
      "edge": {
        "direction": "to",
        "resource": {
          "kind": "Deployment",
          "namespace": "linkerd",
          "name": "linkerd-prometheus"
        }
      },
      "metrics": [
        {
          "name": "p99_response_latency",
          "unit": "ms",
          "value": "368"
        },
        {
          "name": "p90_response_latency",
          "unit": "ms",
          "value": "247199m"
        },
        {
          "name": "p50_response_latency",
          "unit": "ms",
          "value": "120731m"
        },
        {
          "name": "success_count",
          "value": "1008100m"
        },
        {
          "name": "failure_count",
          "value": "0"
        }
      ]
    }
  ]
}
```

The support for istio and consul connect is being worked up on right now and the API and responses will have the same structure unless there are no changes to the spec.