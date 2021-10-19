# patch-operator

> kubernetes operator that patches resources

### Usage

#### Recalibration

The patch will be recalibrated (forced to apply again) any time the
spec changes. It is a common practice to set the value of `spec.epoch`
to the current timestamp, thus forcing the patch to recalibrate every
time a deployment is updated.
