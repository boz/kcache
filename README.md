# kcache: kubernetes object cache

Kcache is a replacement for [k8s.io/client-go/tools/cache](https://github.com/kubernetes/client-go/tree/master/tools/cache).

See example usage [here](_example/main.go)

## Key differences

 * Uses goprocs and channels instead of mutexes and condition variables (no need to poll when ready)
 * Allows multiple subscribers
 * Does not emit "add" events on every resync
 * Does not (currently) support indexes.
 * Kcache does not handle connection/timeout errors well

## Status

WIP; not ready for production.

### TODO

 * Tests
 * Indexes (will be subscribers) and filters
 * Documentation

