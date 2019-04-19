# Mailboat: a verified mail server

[![Build Status](https://travis-ci.com/tchajed/mailboat.svg?branch=master)](https://travis-ci.com/tchajed/mailboat)

Mailboat is a qmail-like mail server with a proof in [Armada](https://github.com/mit-pdos/armada) of functional correctness, including linearizability in the presence of crashes and concurrency.

## Benchmarking

Run `./run-mailboat.sh` to run a mixed-workload benchmark, which is output from the test `TestMixedLoad`.
