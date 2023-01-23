# Cartographer

Use cartographer to build your own custom MLOps environment on top of Kubernetes
with tools like MLFlow, Ray, and Prefect. 

## Description

We made this project to help us manage our own machine learning infrastructure.
One of the challenges we face daily is the fact that most products offer almost
no extensibility and usally don't work on your workstation.

We like to use tools that offer a way to start on your workstation and then
move to the central machine-learning environment without changing code.

Prefect is a workflow solution for Python that you can start with from your own
machine with a few decorated methods. Cartographer hosts a Prefect controller
server and a set of agents to run workflows on. 

More often than not, machine-learning workflows grow beyond a single machine.
We use Ray to help us take machine-learning code that we develop and test on
a single machine and scale that to multiple agents across a cluster.

We need reproducable projects for our clients. To support this idea, we use 
MLFlow to track experiments, models, and associated data. You can use MLFlow
on your own workstation too. 

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use 
[KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run
against a remote cluster.

**Note:** Your controller will automatically use the current context in your
kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster

We have a sample configuration that you can use from the `config` directory. 
Please follow these instructions to set up the operator and associated components
on your cluster.

Before installing anything else, we need to set up cert-manager to make sure we
can secure communication between the operator and the rest of Kubernetes. Run
the following command to install cert-manager:

```
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml
```

To use the components deployed by cartographer, you'll need a postgres database.
We recommend using the crunchy data postgres operator. We've included the 
required installation files in the repository for your reference. To install
the postgres operator, run the following command:

```
kubectl apply --server-side -k ./config/postgres/operator/install/
```

Next, you need to install the cartographer operator:

```
kubectl apply -k ./config/default
```

### Deploying your first workspace

We've included a sample workspace in the repository. You can deploy it using
the following command:

```
kubectl apply -k ./config/workspaces/
```

This command assumes that you're using the postgres operator we included in the
repository. If you're using something else, you'll need to make sure that you
change the configuration accordingly.

After deploying the sample workspace you can access the workflow-controller by
using a port-forward. For example:

* Prefect server: `kubectl port-forward svc/workspace-sample-orion-server 4200:4200`
* MLFlow server: `kubectl port-forward svc/workspace-sample-mlflow-server 5000:5000`

## License

Copyright 2023 Willem Meints.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## Documentation

### Included components

This project relies on a number of other projects to perform its task. 

In essence, the operator only manages components not provided elsewhere. 
For example, we currently ship these components as part of the operator:

* MLFlow - We use this for experiment and model tracking
* Prefect - We use this for building ML pipelines

### Project layout

The project has the following layout:

```
├── api                          # The API definitions for the operator
├── config                       # The Kubernetes manifests to install the operator
│   ├── certmanager              # The certificate issuer and certificate to secure the operator
│   ├── crd                      # Custom resource definitions
│   ├── default                  # Default installation configuration
│   ├── manager                  # Manager deployment definitions
│   ├── postgres                 # Included postgres operator install files
│   ├── prometheus               # Metrics collection configuration
│   ├── rbac                     # RBAC configuration
│   ├── scorecard                # Validation tests
│   ├── webhook                  # Mutation and validation hooks
│   └── workspaces               # Sample workspaces
├── controllers                  # Implementation of the controllers
└─── docker                       # Customized docker images
    ├── experiment-tracking
    ├── workflow-agent
    └── workflow-controller
```


