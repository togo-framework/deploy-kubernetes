# deploy-kubernetes — docs

**Kubernetes deploy.** Render a Deployment+Service (or apply your manifests) and `kubectl apply`.

## Install

```bash
togo install togo-framework/deploy-kubernetes
```

Registers on the [`deploy`](https://github.com/togo-framework/deploy) base; select it with **deploy.provider in togo.yaml (or DEPLOY_PROVIDER)**, then use **`togo deploy`**.

## Interface

`Deployer` — `Provision`/`Deploy`/`Destroy`/`Status` over a `Spec{App,Dir,BuildCmd,Host,User,Image,Region,Domain}` built from your `togo.yaml`.

## Configuration

| Env var | Description |
|---|---|
| `KUBE_NAMESPACE` | Kubernetes namespace to deploy into (default `default`). Uses your active `KUBECONFIG`. |

## Usage & notes

Uses your active `KUBECONFIG`. Renders a Deployment+Service for `spec.Image` into `KUBE_NAMESPACE`, or applies manifests you provide.

## Example

```bash
togo deploy --provider kubernetes --dry-run   # preview the plan
togo deploy --provider kubernetes
```

## Links

- [kubectl](https://kubernetes.io/docs/reference/kubectl/)
- [Marketplace](https://to-go.dev/marketplace)
- [Source](https://github.com/togo-framework/deploy-kubernetes)
