# Cartographer

This project implements a custom operator for Kubernetes that manages the MLOps environments for Aigency.
We've started this project, because we're with a small number of people working on various projects for
various customers.

We want to spend our time building and deploying models. The rest of our stuff should just work. Therefore we're trying
to automate as much of the operations work as possible for our environments.

------------------------------------------------------------------------------------------------------------------------

**Note** This is an experiment. Feel free to give it a try on your own Kubernetes environment. 
Currently, only the workflow bits work.

------------------------------------------------------------------------------------------------------------------------

## Description

Cartographer allows you to create a `Workspace` in your kubernetes cluster. The controller automatically deploys 
a number of components for the workspaces it manages:

- Prefect orion server to manage pipelines
- One or more pools of Prefect agents to run the pipelines

You can scale the resources in the workspace by editing the properties of the workspace accordingly.
For now, you'll need to look at the YAML files in the `samples` directory to learn more about
the structure of a workspace definition.

After the workspace is configured, you can forward the orion service by executing the following command:

```
kubectl port-forward svc/<environment>-orion-server 4200:4200
```

Make sure to use the name of the workspace to forward to the correct namespace. 
After forwarding the port, you can access the orion server on `http://localhost:4200`.

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Installing the operator

Unpack the release zip into a dedicated folder and verify the settings in `install/kustomization.yaml`. 
When you're satisfied with the settings in the kustomization file, run the following command
to install the operator:

```
kubectl apply -k install
```

### Uninstalling the operator

When you no longer want to use the operator, use the following command to uninstall
the operator:

```
kubectl delete -k install
```

### Deploying your first workspace

After installing the operator, you can create a new workspace. For example,
you can run the following command to create a very basic workspace:

```
kubectl apply -f samples/basic-workspace.yaml
```

When you've created the workspace, you can connect it by forwarding the port
to the prefect server in the workspace:

```
kubectl port-forward svc/test-environment-orion-server 4200:4200
```

## Documentation
TODO: Describe how the operator works internally