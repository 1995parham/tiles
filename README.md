# Tiles
[![Drone (cloud)](https://img.shields.io/drone/build/1995parham/tiles.svg?style=flat-square)](https://cloud.drone.io/1995parham/tiles)

## Introduction
Tiles is a solution for [tile38](https://github.com/tidwall/tile38) sharding.
It shards write requests using [geohash](https://en.wikipedia.org/wiki/Geohash) and acts like a proxy so clients can connect to it without any modification.

## Supported Commands
- SET
- SCAN
- WITHIN
- GET (only returns the first result)
- PIPELINE (This is a client feature not a tile feature!)
