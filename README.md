# informer-test

This repo is used to investigate the use of informers for interacting with k8s resources.

## Getting started
Be logged into a cluster that has the namespace `default` and run `make run`.
Minicube is recommended as the example is so simple the resource requirements are very small.

## What happens
In the default namespace `configmaps` are managed. 
These configmaps are specified in the `assets/configmap` folder.
Adding extra resources will be created on restart of the application.

Deleting the resources from the cluster will cause them to be recreated.
There is no validation done on the contains of the configmaps but log messages are given when they are updated.


The aim of this poc is to show how reconcile loops are not required to maintain the resources.