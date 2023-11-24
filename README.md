# patch-operator

> kubernetes operator that patches resources

**This project is depricated in favor of [Kyverno](https://kyverno.io).**
Kyverno can do essentially everything this project set out to do
and much more.

## Migrate to Kyverno

The example [below](#example) can achieve the same effect as the
following Kyverno policy.

```yaml
apiVersion: kyverno.io/v1
kind: Policy
metadata:
  name: patch-hello-configmap
spec:
  background: true
  mutateExistingOnPolicyUpdate: true
  rules:
    - name: hello-configmap
      match:
        resources:
          kinds:
            - /*/ConfigMap
          names:
            - hello
      mutate:
        targets:
          - apiVersion: v1
            kind: ConfigMap
            name: hello
        patchStrategicMerge:
          data:
            hello: world
```

## Usage

### Recalibration

The patch will be recalibrated (forced to apply again) any time the
spec changes. It is a common practice to set the value of `spec.epoch`
to the current timestamp, thus forcing the patch to recalibrate every
time a deployment is updated.

### Example

```yaml
apiVersion: patch.rock8s.com/v1alpha1
kind: Patch
metadata:
  name: patch-hello-configmap
spec:
  patches:
    - id: hello-configmap
      target:
        apiVersion: v1
        kind: ConfigMap
        name: hello
      waitForResource: true
      type: merge
      patch: |
        data:
          hello: world
```
