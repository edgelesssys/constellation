---
slug: /
id: intro
---
# Introduction

Welcome to the documentation of Constellation! Constellation is a Kubernetes engine that aims to provide the best possible data security.

![Constellation concept](/img/concept.svg)

 Constellation shields your entire Kubernetes cluster from the underlying cloud infrastructure. Everything inside is always encrypted, including at runtime in memory. For this, Constellation leverages a technology called *confidential computing* and more specifically Confidential VMs.

:::tip
See the 📄[whitepaper](https://content.edgeless.systems/hubfs/Confidential%20Computing%20Whitepaper.pdf) for more information on confidential computing.
:::

## Goals

From a security perspective, Constellation is designed to keep all data always encrypted and to prevent any access from the underlying (cloud) infrastructure. This includes access from datacenter employees, privileged cloud admins, and attackers coming through the infrastructure. Such attackers could be malicious co-tenants escalating their privileges or hackers who managed to compromise a cloud server.

From a DevOps perspective, Constellation is designed to work just like what you would expect from a modern Kubernetes engine.

## Use cases

Constellation provides unique security [features](overview/confidential-kubernetes.md) and [benefits](overview/security-benefits.md). The core use cases are:

* Increasing the overall security of your clusters
* Increasing the trustworthiness of your SaaS offerings
* Moving sensitive workloads from on-prem to the cloud
* Meeting regulatory requirements

## Next steps

You can learn more about the concept of Confidential Kubernetes, features, security benefits, and performance of Constellation in the *Basics* section. To jump right into the action head to *Getting started*.
