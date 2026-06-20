## Why

Oktopus needs a reliable way to execute jobs on local or distributed workers without duplicate execution, lost work, hidden state, or unsafe capability access. Workers may run agent runtimes, shell commands, tool adapters, verifiers, inference evaluators, or review personas. The controller must remain the source of truth while workers stay replaceable and disposable.

A worker lease protocol provides the operational contract: how workers register, advertise capabilities, request work, atomically accept jobs, heartbeat, stream logs/events, upload artifacts, complete, fail, cancel, and recover from crashes.

## What Changes

- Define worker roles and lifecycle.
- Define polling, matching, atomic lease acquisition, heartbeat, completion, failure, expiration, and cancellation semantics.
- Define worker capability labels and resource constraints.
- Define local-first worker behavior and distributed deployment behavior.
- Define idempotency, retries, and exactly-once-vs-at-least-once expectations.
- Define event and artifact responsibilities for workers.
- Define initial worker kinds: `shell`, `agent`, `tool`, `verifier`, `review`, `inference`, and `synthesis`.

## Out of Scope

- Implementing the network API or CLI.
- Implementing the first worker runtime.
- Choosing final queue technology.
- Implementing sandbox backends such as Docker, OpenShell, microVMs, or Kubernetes.
- Full policy language for secrets, network, and filesystem access.

## Impact

This change establishes the distributed execution contract. Future implementation can begin with an in-process/local worker and later move to HTTP/gRPC workers, container workers, and remote worker pools without changing run/job semantics.
