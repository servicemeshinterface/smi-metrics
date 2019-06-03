k8s_yaml(local("helm template chart -f linkerd.yaml --name dev"))
watch_file('chart')

docker_build('thomasr/smi-metrics', '.')

k8s_resource('dev-smi-metrics', port_forwards=['8080:8080', '8081:8081'])
