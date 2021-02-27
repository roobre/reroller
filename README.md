# reroller

Reroller is a tool that monitors deployments and daemonSets in kubernetes clusters for updates in the container images they use, and rolls an update whenever it's found.

Reroller will, by default, only process rollouts that:
- Are annotated with `reroller.roob.re/reroll: true`
- Have `image.pullPolicy: Always`

The first rule can be overridden by running reroller with the `-unannotated` flag. In this case, reroller will check and rollout updates for everything excepting rollouts annotated with `reroller.roob.re/reroll: false`.

The `image.pullPolicy == Always` check cannot be overridden, as it wouldn't make sense to redeploy without it.
