# topologyinfo - go package to extract hardware topology from sysfs

## Goals

- provide a trivial to vendor (few if any non-stdlib deps) package
- provide easy access to data exported by sysfs
- consolidate all the common tasks (e.g. learn about NUMA distances, online CPUs, NUMA affinities...)
  needed to work in the openshift-kni areas

## Non-Goals

- provide representation different than the sysfs one
- compete with more comprehensive packages like [cadvisor](https://github.com/google/cadvisor)

## license
(C) 2020 Red Hat Inc and licensed under the Apache License v2

