k8s_yaml('k8s/dev.yaml')

docker_build('thomasr/smi-metrics', '.')

k8s_resource('smi-metrics', port_forwards=['8080:8080', '8081:8081'])
