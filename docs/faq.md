# Frequently Asked Questions

Depending on the context you're using hcdiag in, you may have some questions about the tool. Here are questions we see frequently.

### Why is hcdiag checking for all HashiCorp products when I only want one?

We heard feedback from `hcdiag` users that the default behavior, with no flags provided, could be confusing. In that case,
it only gathered host details from the machine where it ran. This could cause users to think that data had been
collected about their HashiCorp products when it actually hadn't.

Beginning in version `0.4.0`, the default behavior - when no product flags are provided (such as `-consul`, `-vault`, etc.) -
is to detect which product CLIs are installed on the system. Any that are found are added, automatically, to the list of
products for which to pull diagnostics.

In the event that you run `hcdiag` with no product flags on a system where the CLI for multiple HashiCorp products are installed,
but any of them are not configured, this can lead to failures. For example, if you had both Terraform and Vault CLIs installed,
but only Vault was configured, then `hcdiag` might fail on Terraform Enterprise checks. A log message is displayed when auto-detection
is used, and it will indicate which products were found on your system. If you find that you wish to limit which products
are executed, please use the appropriate product flag. In the previous example, you would want to run `hcdiag -vault` instead
of just `hcdiag` because you have both CLIs on your system, but Terraform is not actually configured for use.

### How do I use hcdiag with Kubernetes?

Although Kubernetes is a complex topic with many configuration options, the key to remember is that hcdiag must be able
to communicate with your pod(s) via a network connection in order to get diagnostic details from the products running in
those pods. If the management interface is not already exposed externally from k8s, you may consider setting up a port-forward
when collecting diagnostics. The command would be similar to `kubectl -n consul port-forward consul-server-0 8500:8500`; 
in this example, we are in the namespace `consul` (noted by `-n consul`), and we are forwarding the external port `8500`
(the port before the `:`) to port `8500` (the port after the `:`) on the pod `consul-server-0`.

If you would like to experiment with a setup like this, assuming you have both minikube and helm installed on your machine,
you could use the following as a reference:

```shell
minikube start
helm repo add hashicorp https://helm.releases.hashicorp.com
helm search repo hashicorp/consul
helm install consul hashicorp/consul --set global.name=consul --create-namespace -n consul
kubectl -n consul port-forward consul-server-0 8500:8500

# Now, run hcdiag
hcdiag -consul
```
