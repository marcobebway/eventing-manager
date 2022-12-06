# Eventing Manager

## Overview

This is a PoC for the Eventing manager to provision and deprovision Kyma Eventing resources.

> Note: This is not an official implementation of the Eventing module manager.
 
> Note: Scaffolding is inspired by the Kyma [template-operator](https://github.com/kyma-project/template-operator) and
> the [Keda manager](https://github.com/kyma-project/keda-manager).

## Tasks

- [x] Eventing manager scaffolding.
- [x] Pass installation overrides from the Eventing manager to the Eventing charts.
- [x] Provision NATS.
- [ ] Deprovision NATS.
- [x] Update the Eventing CR status.
- [ ] Add/Remove resources and update the status when dependencies come and go.
- [x] Control the naming of the Eventing operator and the Eventing resources.
- [x] Write unit-tests.
- [ ] Implement graceful shutdown to deprovision created resources.

## Setup

### Provision Kyma and the Eventing manager on k3d

```bash
$ cd eventing-manager/hack/local/eventing/
$ make run-dev
```

> Note: Since the module-manager, lifecycle-manager and Kyma CLI is still in alpha development phase, there is a lot
> of resource patching to correctly set up the k3d cluster. Currently, this is abstracted using the make target 
> `make run-dev` which is suitable only for local development on k3d.

> Note: Always refer to the latest version of the Kyma [template-operator](https://github.com/kyma-project/template-operator)
> as a reference implementation of Kyma module managers.

### Switch Eventing backends

**The Eventing backend can be configured in the Eventing CR**

```bash
$ kubectl edit eventings.operator.kyma-project.io -n kcp-system default-kyma-eventing

# spec:
#   backend:
#     type: "eventmesh"
```

**The backend changes are reported in the Eventing manager logs**

```bash
$ stern -n kyma-system eventing-controller-manager

{
   "controller":"eventing",
   "controllerGroup":"operator.kyma-project.io",
   "controllerKind":"Eventing",
   "eventing":{
      "name":"default-kyma-eventing",
      "namespace":"kcp-system"
   },
   "namespace":"kcp-system",
   "name":"default-kyma-eventing",
   "reconcileID":"7a9a8e34-ff18-4f82-bc2b-f79e41d2e883",
   "flags":{
      "backend":{
         "type":"nats"
      }
   },
   "backend":"nats"
}
```

### Add/Remove the Eventing module

**The active modules can be configured in the Kyma CR `spec.modules`**

```bash
$ kubectl edit kymas.operator.kyma-project.io -n kcp-system default-kyma

# spec:
#   modules:
#   - channel: alpha
#     name: eventing
```

### Verification

**The Eventing CR status should be ready**

```bash
$ kubectl get eventings.operator.kyma-project.io -n kcp-system default-kyma-eventing -oyaml

apiVersion: operator.kyma-project.io/v1alpha1
kind: Eventing
metadata:
  creationTimestamp: "2022-12-05T19:53:13Z"
  finalizers:
  - eventing-manager.kyma-project.io/deletion-hook
  generation: 1
  labels:
    operator.kyma-project.io/kyma-name: eventing-sample
  name: default-kyma-eventing
  namespace: kcp-system
  resourceVersion: "883"
  uid: 9b7c0c06-aa3e-43ba-88b6-23a589a34fcc
spec:
  backend:
    type: nats
status:
  state: Ready
```

**The Kyma CR status should be ready**

```bash
$ kubectl get kymas.operator.kyma-project.io -n kcp-system default-kyma -oyaml

apiVersion: operator.kyma-project.io/v1alpha1
kind: Kyma
metadata:
  annotations:
    cli.kyma-project.io/source: deploy
  creationTimestamp: "2022-12-05T19:53:03Z"
  finalizers:
  - operator.kyma-project.io/Kyma
  generation: 3
  labels:
    operator.kyma-project.io/managed-by: lifecycle-manager
  name: default-kyma
  namespace: kcp-system
  resourceVersion: "677"
  uid: b94836f0-b5a8-45af-b8fa-0bbb9161b1b5
spec:
  channel: regular
  modules:
  - channel: alpha
    name: eventing
  sync:
    enabled: false
    moduleCatalog: true
    noModuleCopy: true
    strategy: secret
status:
  activeChannel: regular
  conditions:
  - lastTransitionTime: "2022-12-05T19:53:13Z"
    message: all modules are in ready state
    observedGeneration: 3
    reason: ModulesAreReady
    status: "True"
    type: Ready
  moduleStatus:
  - generation: 1
    moduleName: eventing
    name: default-kyma-eventing
    namespace: kcp-system
    state: Ready
    templateInfo:
      channel: regular
      generation: 1
      gvk:
        group: operator.kyma-project.io
        kind: Manifest
        version: v1alpha1
      name: moduletemplate-eventing
      namespace: kcp-system
      version: 0.0.4
  state: Ready
```

> Note: The Eventing module should be reported as ready in the Kyma CR status.

**The Eventing resources are provisioned/deprovisioned**

- When the Eventing module is added/removed.

  TBD

- When the Eventing backend is switched from `eventmesh` to `nats` and vice-versa. 

  TBD

### Cleanup

```bash
$ cd eventing-manager/hack/local/eventing/
$ make stop
```

## Compatibility issues

- The module-manager, lifecycle-manager and Kyma CLI can go out of sync or have breaking changes at any time, this is 
  why we use the `kustomization` flag to pass their correct templates to the Kyma CLI. Also, the `template` flag is
  used for development only and can be removed at anytime.

  ```bash
  $ kyma alpha deploy \
      --template=module-template.yaml \
      --kustomization https://github.com/kyma-project/lifecycle-manager/config/default@main \
      --kustomization https://github.com/kyma-project/module-manager/config/default@main
  ```

## How-to

### Pass installation overrides from the Eventing manager to the charts

The installation overrides are passed as InstallationSpec flags to the module manager reconciler which internally
applies them on the Eventing module charts. The flags are passed as follows:

```
types.InstallationSpec{
    ChartPath: m.chartPath,
    ChartFlags: types.ChartFlags{
        ConfigFlags: types.Flags{
            "Namespace":       chartNs,
            "CreateNamespace": true,
        },
        SetFlags: types.Flags{
            "nats": map[string]interface{}{
                "enabled": eventing.Spec.BackendSpec.Type == v1alpha1.BackendTypeNats,
            },
        },
    },
}
```

Where `nats` is the chart name, and `enabled` is the value to be overridden 
(see [controller](controllers/eventing_controller.go) and [chart](module-chart/charts/nats/values.yaml)).

### Deprovision Eventing resources

TBD.

### React on the availability of the Eventing dependencies

TBD.

## Followup issues

- https://github.com/kyma-project/module-manager/issues/188
- https://github.com/kyma-project/module-manager/issues/190
- https://github.com/kyma-project/module-manager/issues/191

## References

- [Module manager](https://github.com/kyma-project/module-manager).
- [Lifecycle manager](https://github.com/kyma-project/lifecycle-manager).
- [Template operator](https://github.com/kyma-project/template-operator).
