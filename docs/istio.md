# SMI Metrics with Istio

The SMI Metrics Extension APIService makes use of prometheus metrics exposed by Istio to resopond to metric queries like p90, request counts, etc.

There is a configuration change required on the Istio metrics for SMI Metrics to work, as the default metrics dosen't have enough information to relate thosoe metrics with that of Kubernetes resources.

SMI Metrics makes use of the following Metric Instances in Istio:

- `istio_request_duration_seconds_bucket`
- `istio_requests_total`

As SMI Metrics allows users to query for metrics based on the resource `kind` i.e pod, deployment, etc. These metrics should have corresponding labels that provide this information.

The labels required by SMI Metrics on the above metrics are:

- `source_owner`, `destination_owner` : To identify the workload types that these metrics belong to i.e Deployment, Daemonset, Job, etc.
- `source_uid`, `destination_uid` : To identify the pod instance that these metrics belong to.
- `source_workload_namespace`,  `destination_workload_namespace` : To identify the namespace that a particular metric belongs to.

The first four labels are not present in the default installation of Istio. So, For SMI to work with Istio, these four labels have to be added manually to the above mentioned Metric Instances.

This can be done by adding those labels to the `dimensions` field of the corresponding Metric Instance definitions and the `label_names` field of Handler definitions.

```diff
diff --git install/kubernetes/istio-demo-auth.yaml install/kubernetes/istio-demo-auth.yaml
index 5175ae1..fef8c6b 100644
--- install/kubernetes/istio-demo-auth.yaml
+++ install/kubernetes/istio-demo-auth.yaml
 ---
 apiVersion: "config.istio.io/v1alpha2"
 kind: instance
 metadata:
   name: requestcount
   namespace: istio-system
   labels:
     app: mixer
     chart: mixer
     heritage: Tiller
     release: istio
 spec:
   compiledTemplate: metric
   params:
     value: "1"
     dimensions:
       reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
       source_workload: source.workload.name | "unknown"
       source_workload_namespace: source.workload.namespace | "unknown"
+      source_owner: source.owner | "unknown"
       source_principal: source.principal | "unknown"
       source_app: source.labels["app"] | "unknown"
+      source_uid: source.uid | "unknown"
       source_version: source.labels["version"] | "unknown"
       destination_workload: destination.workload.name | "unknown"
       destination_workload_namespace: destination.workload.namespace | "unknown"
+      destination_owner: destination.owner | "unknown"
       destination_principal: destination.principal | "unknown"
       destination_app: destination.labels["app"] | "unknown"
+      destination_uid: destination.uid | "unknown"
       destination_version: destination.labels["version"] | "unknown"
       destination_service: destination.service.host | "unknown"
       destination_service_name: destination.service.name | "unknown"
       destination_service_namespace: destination.service.namespace | "unknown"
       request_protocol: api.protocol | context.protocol | "unknown"
       response_code: response.code | 200
       response_flags: context.proxy_error_code | "-"
       permissive_response_code: rbac.permissive.response_code | "none"
       permissive_response_policyid: rbac.permissive.effective_policy_id | "none"
       connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
     monitored_resource_type: '"UNSPECIFIED"'
 ---
 apiVersion: "config.istio.io/v1alpha2"
 kind: instance
 metadata:
   name: requestduration
   namespace: istio-system
   labels:
     app: mixer
     chart: mixer
     heritage: Tiller
     release: istio
 spec:
   compiledTemplate: metric
   params:
     value: response.duration | "0ms"
     dimensions:
       reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
       source_workload: source.workload.name | "unknown"
       source_workload_namespace: source.workload.namespace | "unknown"
+      source_owner: source.owner | "unknown"
       source_principal: source.principal | "unknown"
       source_app: source.labels["app"] | "unknown"
+      source_uid: source.uid | "unknown"
       source_version: source.labels["version"] | "unknown"
       destination_workload: destination.workload.name | "unknown"
       destination_workload_namespace: destination.workload.namespace | "unknown"
+      destination_owner: destination.owner | "unknown"
       destination_principal: destination.principal | "unknown"
       destination_app: destination.labels["app"] | "unknown"
+      destination_uid: destination.uid | "unknown"
       destination_version: destination.labels["version"] | "unknown"
       destination_service: destination.service.host | "unknown"
       destination_service_name: destination.service.name | "unknown"
       destination_service_namespace: destination.service.namespace | "unknown"
       request_protocol: api.protocol | context.protocol | "unknown"
       response_code: response.code | 200
       response_flags: context.proxy_error_code | "-"
       permissive_response_code: rbac.permissive.response_code | "none" 
       permissive_response_policyid: rbac.permissive.effective_policy_id | "none"
       connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
     monitored_resource_type: '"UNSPECIFIED"'
 ---
 apiVersion: "config.istio.io/v1alpha2"
 kind: handler
 metadata:
   name: prometheus
   namespace: istio-system
   labels:
     app: mixer
     chart: mixer
     heritage: Tiller
     release: istio
 spec:
   compiledAdapter: prometheus
   params:
     metricsExpirationPolicy:
       metricsExpiryDuration: "10m"
     metrics:
     - name: requests_total
       instance_name: requestcount.instance.istio-system
       kind: COUNTER
       label_names:
       - reporter
       - source_app
+      - source_uid
       - source_principal
       - source_workload
       - source_workload_namespace
+      - source_owner
       - source_version
       - destination_app
+      - destination_uid
       - destination_principal
       - destination_workload
       - destination_workload_namespace
+      - destination_owner
       - destination_version
       - destination_service
       - destination_service_name
       - destination_service_namespace
       - request_protocol
       - response_code
       - response_flags
       - permissive_response_code
       - permissive_response_policyid
       - connection_security_policy
     - name: request_duration_seconds
       instance_name: requestduration.instance.istio-system
       kind: DISTRIBUTION
       label_names:
       - reporter
       - source_app
+      - source_uid
       - source_principal
       - source_workload
       - source_workload_namespace
+      - source_owner
       - source_version
       - destination_app
+      - destination_uid
       - destination_principal
       - destination_workload
       - destination_workload_namespace
+      - destination_owner
       - destination_version
       - destination_service
       - destination_service_name
       - destination_service_namespace
       - request_protocol
       - response_code
       - response_flags
       - permissive_response_code
       - permissive_response_policyid
       - connection_security_policy
       buckets:
         explicit_buckets:
           bounds: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
.
.
.
```
The above updation in the manifest would add configure Istio to add the labels required by SMI to those labels.

A full Installation manifest with those updates can be found [here](https://gist.github.com/Pothulapati/4a236196141ac1aa4acbb117b7c2e4c8) 

To verify if the labels are added correctly, prometheus metrics can be checked to see if those metrics have the configured labels.


## Installation of SMI

Once Istio is installed with the above mentioned configuration, SMI-Metrics can be installed by cloning the repo and running

```bash
helm template chart -f istio.yaml --name dev | kubectl apply -f -
```