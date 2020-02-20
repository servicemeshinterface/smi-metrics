k8s_yaml(local("helm template chart --set adapter=linkerd --name dev"))
watch_file('chart')
docker_build('deislabs/smi-metrics', '.')

k8s_resource('dev-smi-metrics', port_forwards=['8080:8080', '8081:8081'])
