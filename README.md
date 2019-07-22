# Tiles

## Introduction
Tiles is a solution for [tile38](https://github.com/tidwall/tile38) sharding.
It shards write requests using geohash and acts like a proxy so clients can connect to it without any modification.

## Supported Commands
- SET
- SCAN
- WITHIN
- GET (only returns the first result)
- PIPELINE (This is a client feature not a tile feature!)
