## References
API Docs: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/
Operator SDK User Guide: https://github.com/operator-framework/operator-sdk/blob/master/doc/user-guide.md

## Issues
- Will need to account for https://github.com/containers/libpod/issues/4666 (stuck stopped containers)
Workaround is to run:
```shell script
$ oc debug node/ip-10-0-136-149.us-west-2.compute.internal -- chroot /host podman rm -f toolbox-
```
- How to communicate sosreport path to controller?
  - Some standard JSON logging? Update ConfigMap or something?
  - Any other standard way to pass status to a controller?

## Documentation Needs
- Create install script that deploys yamls 
- Must add privileged scc to sosreport-worker service account!