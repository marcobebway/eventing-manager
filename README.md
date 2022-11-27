# Eventing Manager

## Overview

Eventing Manager is a module compatible with Lifecycle Manager that allows you to add EVENTING Event Driven Autoscaler to the Kyma ecosystem. 

See also:
- [lifecycle-manager documetation](https://github.com/kyma-project/lifecycle-manager)

## Prerequisites

- Access to a k8s cluster
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [kubebuilder](https://book.kubebuilder.io/)

```bash
# you could use one of the following options

# option 1: using brew
brew install kubebuilder

# option 2: fetch sources directly
curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)
chmod +x kubebuilder && mv kubebuilder /usr/local/bin/
```

## Manual `eventing-manager` installation


1. Clone project

```bash
git clone https://github.com/kyma-project/eventing-manager.git && cd eventing-manager/
```

2. Set `eventing-manager` image name

```bash
export IMG=custom-eventing-manager:0.0.1
export K3D_CLUSTER_NAME=eventing-manager-demo
```

3. Build project

```bash
make build
```

4. Build image

```bash
make docker-build
```

5. Push image to registry

<div tabs name="Push image" group="eventing-installation">
  <details>
  <summary label="k3d">
  k3d
  </summary>

   ```bash
   k3d image import $IMG -c $K3D_CLUSTER_NAME
   ```
  </details>
  <details>
  <summary label="Docker registry">
  Globally available Docker registry
  </summary>

   ```bash
   make docker-push
   ```

  </details>
</div>

6. Deploy

```bash
make deploy
```

## Using `eventing-manager`

- Create Eventing instance

```bash
kubectl apply -f config/samples/operator_v1alpha1_eventing.yaml
```

- Delete Eventing instance

```bash
kubectl delete -f config/samples/operator_v1alpha1_eventing.yaml
```

- Update Eventing properties

This example shows how you can modify the Eventing logging level using the `eventing.operator.kyma-project.io` CR

```bash
cat <<EOF | kubectl apply -f -
apiVersion: operator.kyma-project.io/v1alpha1
kind: Eventing
metadata:
  name: eventing-sample
spec:
  backend:
    type: "nats"
EOF
```

## Installation in modular Kyma on the local k3d cluster

>>NOTE: The following steps are scripted in form of make targets. You can execute all the steps by calling `make run` target from `hack/local` folder.

1. Setup local k3d cluster and local Docker registry

# Provision k3d cluster

```bash
kyma provision k3d 
```
2. Add the `etc/hosts` entry for the local `k3d-kyma-registry`.

```
127.0.0.1 k3d-kyma-registry.localhost
127.0.0.1 k3d-kyma-registry
```

3. Build Eventing Manager image
```bash
make module-image IMG_REGISTRY=k3d-kyma-registry:5001/unsigned/operator-images
```

This builds a Docker image for Eventing Manager and pushes it to the registry and path, as defined in `IMG_REGISTRY`.

4. Build Eventing module
```bash
make module-build IMG_REGISTRY=k3d-kyma-registry:5001/unsigned/operator-images MODULE_REGISTRY=k3d-kyma-registry.localhost:5001/unsigned
```

This builds an OCI image for Eventing module and pushes it to the registry and path, as defined in `MODULE_REGISTRY`.

5. Verify if the module and the manager's image are pushed to the local registry

```bash
curl k3d-kyma-registry.localhost:5001/v2/_catalog
{"repositories":["unsigned/component-descriptors/kyma.project.io/module/eventing","unsigned/operator-images/eventing-operator"]}
```

7. Inpect the generated module template

The following are temporary workarounds.

Edit the `template.yaml` file and:

 - change `target` to `control-plane`
>**NOTE:** This is only required in the single cluster mode

```yaml
spec:
  target: control-plane
```

- change the existing repository context in `spec.descriptor.component`:
>**NOTE:** Because Pods inside the k3d cluster use the docker-internal port of the registry, it tries to resolve the registry against port 5000 instead of 5001. K3d has registry aliases but module-manager is not part of k3d and thus does not know how to properly alias `k3d-kyma-registry.localhost:5001`

```yaml
repositoryContexts:                                                                           
- baseUrl: k3d-kyma-registry.localhost:5000/unsigned                                                   
  componentNameMapping: urlPath                                                               
  type: ociRegistry
```

8. Install modular Kyma on the k3d cluster

This installs the latest versions of `module-manager` and `lifecycle-manager`

You can use the `--template` flag to deploy the Eventing module manifest from the beginning or apply it using kubectl later.

```bash
kyma alpha deploy  --template=./template.yaml

- Kustomize ready
- Lifecycle Manager deployed
- Module Manager deployed
- Modules deployed
- Kyma CR deployed
- Kyma deployed successfully!

Kyma is installed in version:
Kyma installation took:		18 seconds

Happy Kyma-ing! :)
```

Kyma installation is ready, but no module is activated yet
```bash
kubectl get kymas.operator.kyma-project.io -A
NAMESPACE    NAME           STATE   AGE
kcp-system   default-kyma   Ready   71s
```

Eventing Module is a known module, but not activated
```bash
kubectl get moduletemplates.operator.kyma-project.io -A 
NAMESPACE    NAME                  AGE
kcp-system   moduletemplate-eventing   2m24s
```

9. Give Module Manager permission to install CustomResourceDefinition (CRD) cluster-wide

>**NOTE:** This is a temporary workaround and is only required in the single-cluster mode

Module-manager must be able to apply CRDs to install modules. In the remote mode (with control-plane managing remote clusters) it gets an administrative kubeconfig, targeting the remote cluster to do so. But in local mode (single-cluster mode), it uses Service Account and does not have permission to create CRDs by default.

Run the following to make sure the module manager's Service Account becomes an administrative role:

```bash
kubectl edit clusterrole module-manager-manager-role
```
add
```yaml
- apiGroups:                                                                                                                  
  - "*"                                                                                                                       
  resources:                                                                                                                  
  - "*"                                                                                                                       
  verbs:                                                                                                                      
  - "*"
```

10. Enable Eventing in Kyma

Edit Kyma CR ...

```bash
kubectl edit kymas.operator.kyma-project.io -n kcp-system default-kyma
```
..to add Eventing module

```yaml
spec:
  modules:
  - name: eventing
```
