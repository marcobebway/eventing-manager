# Eventing Manager

## Overview

This is a PoC for the Eventing manager to provision and deprovision Kyma Eventing resources.

> Note: This is not an official implementation of the Eventing module manager.
 
> Note: Scaffolding is inspired by the Kyma [template-operator](https://github.com/kyma-project/template-operator) and
> the [Keda manager](https://github.com/kyma-project/keda-manager).

## Tasks

- ✅ Eventing manager scaffolding.
- ✅ Pass installation overrides from the Eventing manager to the Eventing charts.
- ✅ Provision NATS.
- ✅ Deprovision NATS.
- ✅ Update the Eventing CR status.
- ✅ Control the naming of the Eventing operator and the Eventing resources.
- ✅ Write unit-tests (used for consistency checks and fast feedback loop).
- ✅ Implement graceful shutdown to deprovision created resources.
- 🚧 [React on the availability of the Eventing dependencies](#react-on-the-availability-of-the-eventing-dependencies).

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

> Note: The Eventing CRD included in this PoC is for demo purposes only. The actual Eventing CRD should be covered 
> in this design proposal [ticket](https://github.com/kyma-project/kyma/issues/16248).

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

- If the Eventing backend is changed from `eventmesh` to `nats`. 

  When the Eventing CR has `spec.backend.type: "nats"`, there should be a statefulset for `nats`:
 
  ```bash
  $ kubectl get statefulsets.apps -n kyma-system
  
  NAME            READY   AGE
  eventing-nats   0/1     27s
  ```

  When the Eventing backend is changed to `spec.backend.type: "eventmesh"`, the `nats` statefulset will be deleted:
  
  ```bash
  $ kubectl get statefulsets.apps -n kyma-system
  
  No resources found in kyma-system namespace.
  ```

- If the Eventing module is added/removed from the Kyma CR `spec.modules`.

  > Note: Currently this is not working, but there is a [PR](https://github.com/kyma-project/module-manager/pull/195) 
  to fix it on the module manager.

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

Inject the following `declarative.ReconcilerOption` in the Eventing controller (see [controller](controllers/eventing_controller.go)).

```
declarative.WithPostRenderTransform # to add prune labels
declarative.WithPostRun             # to remove resource by prune labels
```

### React on the availability of the Eventing dependencies

This is not yet decided:
- Either it will be done globally by the module manager for all Kyma modules, or
- The Eventing manager should take care of this.

Either way, this seems to not be a blocker for the Eventing manager.

## Followup issues

- https://github.com/kyma-project/module-manager/issues/188
- https://github.com/kyma-project/module-manager/issues/190
- https://github.com/kyma-project/module-manager/issues/191

## Known issues

- Currently, the Eventing manager can react to multiple instances of the Eventing CR. [We should limit it to react to
  exactly one instance](https://github.tools.sap/kyma/backlog/issues/3298).

## References

- [Module manager](https://github.com/kyma-project/module-manager).
- [Lifecycle manager](https://github.com/kyma-project/lifecycle-manager).
- [Template operator](https://github.com/kyma-project/template-operator).
