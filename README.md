<div align="center">

<img src="assets/icon.png" width="128" height="128" alt="SCAXE-GO Logo" />

# SCAXE-GO (translatorX)

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)
![MCPE Version](https://img.shields.io/badge/MCPE-0.14.3-green?style=flat)
![Platform](https://img.shields.io/badge/Platform-Windows%20|%20Linux-blue?style=flat)

**A high-performance server core compatible with MCPE 0.14.3, built from scratch in Go**

*Featuring the AnyConvert translation engine (v0.3.9Alpha1) for PE-to-JE cross-play support*

</div>

---

## Features

- **AnyConvert Translation Engine** - Native proxy mode connecting MCPE 0.14.3 clients to Java Edition 1.7.10-compatible backend servers.
- **High-Precision World Generation** - Based on Overworld core algorithm logic, 93.77% terrain consistency.
- **Bit-Level GenLayer Precision** - Biome system achieves 99.9% bit-level accuracy.
- **1:1 Block Property Parity** - 182 registered blocks with properties matching PHP core exactly.
- **High-Performance Concurrency** - Thread-safe chunk generation powered by Go goroutines.
- **Full Protocol Implementation** - Complete MCPE 0.14.3 (Protocol 70) support.
- **Lua Plugin System** - Extensible plugin architecture with Lua scripting and hot-reload.
- **EULA Acceptance** - AGPL-3.0 license agreement on first startup.

---

## AnyConvert Translation Engine (translatorX)

The `translatorX` branch introduces an experimental native proxy engine for MCPE 0.14.3 clients connecting to Java Edition 1.7.10-compatible backend servers.

### Current Status

translatorX is an experimental branch preserved for reference and compatibility testing.

This branch is not part of the main SCAXE-GO release path. Public feature development is currently paused, and no public roadmap is provided.

Critical fixes may be reviewed at the maintainers' discretion. Feature requests, protocol expansion requests, and production-support requests may be closed if they are outside the current maintenance scope.

Production use is not recommended unless you understand the current limitations.

### Proxy Capabilities
- **Protocol Bridging**: Bidirectional translation between MCPE 0.14.3 (Protocol 70) and Java Edition 1.7.10-compatible backends.
- **World Translation**: Real-time chunk mapping and environmental time synchronization.
- **Entity Translation**: Dynamic mapping of entity IDs, metadata, and movement between Java and Bedrock.
- **Authentication**: Seamless handshake and login proxy routing.

---

## Configuration

`server.properties` key settings for both Proxy and Normal modes:

```properties
server-name=Scaxe Go Server
server-port=19132
server-ip=0.0.0.0
# Set to 'proxy' to enable AnyConvert PE-to-JE translation engine
server-mode=proxy
backend-address=127.0.0.1
backend-port=25565
max-players=20
motd=A Scaxe Go Server
gamemode=0
difficulty=1
level-name=world
level-type=gorigional
online-mode=false
view-distance=8
```

---

## Implemented Features (Native Server Mode)

*(The following features apply when the server is running in native/standalone mode instead of proxy mode)*

### Block System
Complete block property system with automated parity verification (182 registered blocks, 100% parity test pass rate).

### World Generation Engine (Gorigional)
World generation engine based on Overworld core algorithm logic with fully implemented Biomes, Density Grid Terrain, Caves, Ravines, Villages, Temples, and Mineshafts.

### Network Protocol Layer
Full implementation of MCPE 0.14.3 (Protocol 70) including BatchPacket (0x92) compression, StartGame initialization, and RakNet reliable transport layer.

### Command & Plugin Systems
45+ admin/player commands implemented. Built-in Lua scripting engine with event listener and command registration APIs.

---

## Technical Highlights

### 128-Height Squash Strategy
Adapted for the MCPE 0.14 128-block height limit by squashing the Y noise segments and base height calculation.

### Concurrency Safety
Fully stateless chunk generation, local buffer noise generation, and thread-safe random number instances.

---

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0).

---

## Unofficial Redistributions Notice

SCAXE-GO is an official SCAXE Team project. Some third-party projects may claim to be based on older SCAXE / ScaxePHP / Genisys-related code, but they are not maintained, tested, audited, endorsed, or distributed by SCAXE Team unless explicitly listed by us.

Users should be careful with third-party redistributions, especially when source code or binaries are distributed only as archive files. Archive-based distribution is not automatically a license violation, but it reduces auditability and makes it harder to verify source history, source/binary correspondence, and GPL/LGPL/AGPL compliance.

For details, see [Unofficial Redistributions Notice](docs/UNOFFICIAL_REDISTRIBUTIONS.md).

<div align="center">

**Made by SCAXE Team**

</div>
