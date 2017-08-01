# kcache: kubernetes object cache [![Build Status](https://travis-ci.org/boz/kcache.svg?branch=master)](https://travis-ci.org/boz/kcache)

Kcache is a replacement for [k8s.io/client-go/tools/cache](https://github.com/kubernetes/client-go/tree/master/tools/cache).

See example usage [here](_example/main.go)

## Key differences

 * Uses goprocs and channels instead of mutexes and condition variables (no need to poll when ready)
 * Allows multiple subscribers
 * Filtering, with in-place filter update.
 * Does not emit "add" events on every resync
 * no "indexes", but easy to roll your own.
 * Types for common objects (currently Pod,Ingress,Service,Secret)

## Status

WIP; not ready for production.

### TODO

 * Tests
 * Documentation
 * Add more generated types
