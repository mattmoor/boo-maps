# Kubernetes `MutableMap` and `ImmutableMap` CRDs.

## Overview

This repository implements an EXPERIMENTAL substitute for Kubernetes
`ConfigMaps`.  The idea behind these resources is to enable users to "freeze"
their `ConfigMap` state at the point they create a Deployment, StatefulSet,
Job, etc.

> Trivia: These live under `boos.mattmoor.io` as I have named them after the
> ghosts in the Super Mario video game franchise that freeze when you are
> looking at them, but move when you look away.

## Installation

You can install this via:

```
kubectl apply -f release.yaml
```


## Lifecycle of a MutableMap

You create a `MutableMap` resource much like a `ConfigMap`, but with the content
under `spec:` in place of `data:`

```
apiVersion: boos.mattmoor.io/v1alpha1
kind: MutableMap
metadata:
  name: my-config
spec:
  foo: bar
```

Each generation of a `MutableMap` will create an immutable snapshot of itself, e.g.

```
apiVersion: boos.mattmoor.io/v1alpha1
kind: ImmutableMap
metadata:
  name: my-config-00001
spec:
  foo: bar
```

Which in turn will stamp out `ConfigMap` resources, e.g.

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config-00001
data:
  foo: bar
```

The `ImmutableMap` disallows mutations via webhook, and the controller will
revert any changes to the underlying `ConfigMap` as they are observed.


## Using `MutableMaps` with resources containing a `PodSpec`

There are several ways that a `PodSpec` can reference a `ConfigMap`, e.g.
a Kubernetes Deployment may project a particular key into an environment variable:

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
spec:
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
      - name: example
        image: docker.io/mattmoor/something:interesting
        env:
        - name: BLAH
          valueFrom:
            configMapKeyRef:
              name: "foo"
              key: "bar"
```

The webhook this registers to aid with our own map types also registers to
mutate the array of built-in Kubernetes types containing a `PodSpec`, so
when applying the above if "foo" is actually a `MutableMap` with generation
`36` what would actually be applied is:

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
spec:
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
      - name: example
        image: docker.io/mattmoor/something:interesting
        env:
        - name: BLAH
          valueFrom:
            configMapKeyRef:
              name: "foo-00036"  # Updated to the frozen ConfigMap
              key: "bar"
```
