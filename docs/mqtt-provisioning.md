# MQTT Provisioning Between SparkEdgeGo and SparkAPI

This document explains how MQTT provisioning works in the current Spark stack.

The important point is that the old `SparkEdge` flow is deprecated.
The official edge runtime is `SparkEdgeGo`, and MQTT provisioning is now handled through:

- `SparkEdgeGo` for the edge runtime
- `SparkAPI` for cloud-side provisioning and MQTT authorization
- `EMQX` as the MQTT broker

## High-level flow

The provisioning flow is:

1. The operator creates a pairing token in `SparkAPI` for a Unit.
2. In `SparkEdgeGo`, the operator enters the pairing token during setup.
3. `SparkEdgeGo` calls the cloud pairing endpoint.
4. `SparkAPI` validates the token and creates MQTT credentials for the edge.
5. `SparkAPI` returns:
   - edge id
   - MQTT broker URL
   - MQTT username
   - MQTT password
   - topic base and standard topics
6. `SparkEdgeGo` stores those credentials in its local SQLite database.
7. `SparkEdgeGo` connects to `EMQX` using the returned credentials.
8. `EMQX` delegates auth and ACL checks back to `SparkAPI`.

## Responsibilities

### SparkAPI

`SparkAPI` is responsible for:

- creating pairing tokens
- pairing the edge to a Unit
- generating MQTT credentials
- exposing HTTP auth and ACL hooks for EMQX
- validating publish and subscribe permissions

### SparkEdgeGo

`SparkEdgeGo` is responsible for:

- collecting local onboarding data
- calling the cloud pairing endpoint
- storing edge identity and MQTT credentials locally
- connecting to EMQX
- publishing status, heartbeat, stats, meta and response data
- retrying queued MQTT messages when needed

### EMQX

`EMQX` is the MQTT broker used by the platform.

It is configured to:

- authenticate clients by calling `SparkAPI`
- authorize pub/sub actions by calling `SparkAPI`
- deny access when the API does not allow the request

## Required configuration

### SparkAPI

The cloud side must expose EMQX and the auth hooks expected by the broker.

Typical environment values:

- `MQTT_BROKER_URL=mqtt://emqx:1883`
- `MQTT_SUPERUSER_ID=sparkapi-superuser`
- `MQTT_SUPERUSER_PASSWORD=<strong-secret>`

The EMQX container must call:

- `POST /mqtt/auth`
- `POST /mqtt/acl`

### SparkEdgeGo

`SparkEdgeGo` reads its runtime configuration from:

- `config.yml`
- environment variables
- built-in defaults

Important config entries:

- `cloud.url`: SparkAPI base URL
- `cloud.mqtt_url`: fallback MQTT broker URL
- `db.file`: local SQLite file
- `auth.jwt_secret`: JWT secret for local API access
- `server.port`: local HTTP port

If the environment variable `MQTT_URL` is set, it overrides the MQTT URL received from the cloud.

## Pairing endpoint

The pairing endpoint in `SparkAPI` returns the data needed by `SparkEdgeGo` to connect to the broker.

The response includes:

- `edge_id`
- `edge_name`
- `mqtt.url`
- `mqtt.username`
- `mqtt.password`
- standard topic names

Example topic base:

```text
spark/{edge_id}
```

Example topics:

- `spark/{edge_id}/status`
- `spark/{edge_id}/heartbeat`
- `spark/{edge_id}/stats`
- `spark/{edge_id}/meta`
- `spark/{edge_id}/commands`
- `spark/{edge_id}/response`
- `spark/{edge_id}/context`
- `spark/{edge_id}/metrics`
- `spark/{edge_id}/logs`

## Topic namespace rules

The expected convention is:

```text
spark/{edge_id}/{subject}
```

The edge should only publish and subscribe inside its own namespace.

That means an edge with id `edge-123` should use topics like:

- `spark/edge-123/status`
- `spark/edge-123/commands`
- `spark/edge-123/response`

## SparkAPI auth and ACL behavior

`SparkAPI` exposes HTTP hooks for EMQX:

- `/mqtt/auth` validates username, password and client id
- `/mqtt/acl` validates publish and subscribe access

The broker authorization logic is:

- allow the SparkAPI superuser for internal subscription and monitoring
- allow a normal edge only inside its own topic namespace
- deny any topic outside `spark/{username}/#`

## SparkEdgeGo connection behavior

After provisioning, `SparkEdgeGo`:

- loads the saved edge identity from SQLite
- loads the saved MQTT credentials from SQLite
- connects to the broker
- subscribes to the command topic
- publishes status and metadata
- starts heartbeat and stats publishing
- retries queued messages when needed

If `MQTT_URL` is present in the environment, it replaces the broker URL returned by the cloud.

## Local development checklist

1. Start `SparkAPI`.
2. Ensure `EMQX` is running and reachable.
3. Generate a pairing token for a Unit in the cloud UI.
4. Open `SparkEdgeGo`.
5. Fill onboarding data if needed.
6. Pair the edge with the token.
7. Verify the saved MQTT credentials in the local config or SQLite state.
8. Confirm the edge connects and publishes on `spark/{edge_id}/#`.

## Troubleshooting

### The edge does not connect

Check:

- the pairing token is valid and unused
- `SparkAPI` is reachable from `SparkEdgeGo`
- the returned MQTT URL points to the correct broker
- EMQX is running
- the MQTT username and password were saved correctly

### EMQX denies authentication

Check:

- `MQTT_SUPERUSER_ID` and `MQTT_SUPERUSER_PASSWORD`
- the edge credential record in `SparkAPI`
- whether the credential was revoked

### EMQX denies topic access

Check:

- the topic is inside `spark/{edge_id}/#`
- the publish or subscribe action matches the ACL rule
- the edge id used by the client matches the stored credential username

## Relevant source files

- `SparkAPI/src/modules/pairing/pairing.service.ts`
- `SparkAPI/src/modules/mqtt-broker/mqtt-broker.service.ts`
- `SparkAPI/src/modules/mqtt-broker/mqtt-broker.controller.ts`
- `SparkEdgeGo/internal/edge/cloud.go`
- `SparkEdgeGo/internal/edge/service.go`
- `SparkEdgeGo/internal/mqtt/client.go`
- `SparkEdgeGo/config.yml`

