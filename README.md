# Biren GPU Device plugin

## About
The Biren GPU device plugin is as Daemonset that allows you to automatically:

 1. Expose the number of GPUs on each nodes for you cluster
 2. Keep track of the health of your GPUs
 3. Run GPU enabled containers in your k8s cluster

This repository contains Biren's official implementation of the [k8s device plugin](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/resource-management/device-plugin.md)
## Prerequisites
The list of prerequisites for running the Biren device plugin is described below:
 1. Biren GPU Driver >= 1.2.2
 2. Kubernetes >=1.13
 3. if need mount dri device, need run `modprobe -v vgem` in host which have gpus

## SVI in Device plugin
1. SVI devices will not be created dynamically anywhere within the k8s software stack (GPU must be configured into svi card and split into svi devices priori)


## SR-IOV in device plugin
1. setup SR-IOV vfio driver
2. run device plugin with --container-runtime kata


## Quick Start
### Deploy
`kubectl create -f deploy/biren-device-plugin.yaml`
### Running GPU Pods
```
$ cat <<EOF | kubectl apply -f 
apiVersion: v1
kind: Pod
metadata:
  name: gpu-pod
spec:
  restartPolicy: Never
  containers:
  - image: ubuntu:20.04
    name: pod1-ctr
    command: ["sleep"]
    args: ["infinity"]
    resources:
      limits:
        birentech.com/gpu: 1
EOF
```


## Command
```
Biren gpu device plugin

Usage:
  br-gpu-device-plugin [flags]

Flags:
      --cdi-feature                enable cdi feature
      --container-runtime string   the container runtime;runc or kata, default is runc
  -h, --help                       help for br-gpu-device-plugin
      --mount-host-path            mount lib and bin folder in host to container, default is false
      --overwrite-cdi-config       overwrite cdi config
      --pulse int                  heart beating every seconds
```

## How to use it 
requests 
`birentech.com/gpu: num`
`birentech.com/1-4-gpu: num`
`birentech.com/1-2-gpu: num`

## CDI (container device interface) Feature

- https://github.com/cncf-tags/container-device-interface

### Version requirements

- kubelet >= 1.28
- containerd >= 1.7.0

### How to use it

#### kubelet

In kubelet version 1.28, the CDI feature is in alpha state, so it needs to be enabled manually. To do this, add the `--feature-gates=DevicePluginCDIDevices=true` argument to the kubelet startup command.

#### containerd

Modify the containerd configuration file as follows:

```toml
[plugins."io.containerd.grpc.v1.cri"]
  cdi_spec_dirs = ["/etc/cdi", "/var/run/cdi"]
  enable_cdi = true
```

#### k8s-device-plugin

Add the startup command parameter `--cdi-feature` to enable the CDI feature. If the CDI feature is enabled, this will generate a biren.yaml file in the node's `/etc/cdi` directory, which defines the configuration of CDI. If the startup command parameter includes `--overwrite-cdi-config`, the configuration file will be overwritten each time it starts. Otherwise, if the biren.yaml configuration file already exists, it will not be overwritten.

k8s-device-plugin startup command example:

```yaml
        command: 
        - "/root/k8s-device-plugin"
        args: 
        - "--pulse" 
        - "300"
        - "--container-runtime"
        - "runc"
        - "--cdi-feature" # enable cdi feature
        - "--overwrite-cdi-config" # overwrite cdi config
```
