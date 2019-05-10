k8s_yaml(local("helm template chart -f dev.yaml --name dev"))
watch_file('chart')

docker_build('thomasr/smi-metrics', '.', build_args={
    'NETRC': str(local('cat ~/.netrc'))
})

k8s_resource('dev-smi-metrics', port_forwards=['8080:8080', '8081:8081'])
