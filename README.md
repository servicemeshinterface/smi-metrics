## Running

Take a look at [k8s/dev.yaml](k8s/dev.yaml). The image will either need to be
local or hosted elsewhere (not currently public). Change `prometheus-url` to
point to where it is installed.

## Development

- Get [tilt](https://tilt.dev/), run `tilt up`.
- The prometheus API client has been mocked out with a tool `mockery`. When
  bumping the API client version, a new mock will need to be generated. This can
  be done by checking out the correct version of the API client repo, running
  `mockery -name API` and copying the `mocks` folder into `pkg/metrics/mocks`.

## Admin

- `/debug`
- `/metrics`
- `/status`

## TODO

### API Questions

- ObjectMeta includes OwnerReferences, Labels and Annotations. Should any of
  these be included as part of TrafficMetrics?
- ObjectReference has ResourceVersion and APIVersion, pull these in?

### Internal details

- export prometheus for client-go
- integrate swagger with apiservice (OpenAPI AggregationController)
