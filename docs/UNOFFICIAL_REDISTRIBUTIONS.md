# Unofficial Redistributions Notice

This document explains how SCAXE Team treats third-party projects that claim to be based on older SCAXE / ScaxePHP / Genisys-related code.

## Official status

SCAXE-GO and repositories published under the `ScaxeTeam` GitHub organization are official SCAXE Team projects unless stated otherwise.

Third-party redistributions are not maintained, tested, audited, endorsed, or distributed by SCAXE Team unless they are explicitly listed by us.

## License compliance

If a third-party project redistributes code or binaries derived from SCAXE / ScaxePHP / Genisys / PocketMine-MP, it should comply with the applicable upstream open-source licenses. Depending on the actual upstream code and license chain, this may include GPL, LGPL, or AGPL requirements.

These requirements may include:

- preserving original license texts;
- preserving copyright notices and attribution;
- preserving warranty disclaimers;
- clearly stating that the work was modified;
- providing modification dates or other clear modification notices;
- providing the complete corresponding source code for distributed binaries;
- providing build scripts, dependency information, and installation materials needed to reproduce or modify the distributed binaries;
- clearly stating that the project is not an official SCAXE Team release.

## Archive-based distribution

Some third-party redistributions may provide source code or binaries mainly through archive files such as `src.zip`, `bin(Windows).zip`, or `bin(Linux).7z`.

Archive-based distribution is not automatically a license violation. However, it reduces auditability and can make it harder for users and developers to verify:

- the real source history;
- whether the published source code is complete;
- whether the provided binaries were built from the published source code;
- whether upstream license texts and copyright notices were preserved;
- whether required GPL/LGPL/AGPL obligations are being followed.

For security and license-compliance reasons, users should be careful with third-party binary archives that cannot be independently reproduced from complete source code.

## User recommendation

Users should prefer official SCAXE Team repositories and releases.

Do not run third-party binaries unless the project provides complete source code, clear license information, upstream attribution, and reproducible build information.

SCAXE Team supports compliant open-source forks, modifications, and redistributions. SCAXE Team does not endorse redistributions that remove attribution, omit license notices, fail to provide corresponding source code, or package modified code and binaries in a way that makes independent verification difficult.