# GO FTP Client

[![Units Tests](https://github.com/alexZaicev/go-ftp-client/actions/workflows/unit-tests.yaml/badge.svg)](https://github.com/alexZaicev/go-ftp-client/actions/workflows/unit-tests.yaml)
[![golint](https://github.com/alexZaicev/go-ftp-client/actions/workflows/golint.yaml/badge.svg)](https://github.com/alexZaicev/go-ftp-client/actions/workflows/golint.yaml)
[![Coverage Status](https://coveralls.io/repos/github/alexZaicev/go-ftp-client/badge.svg)](https://coveralls.io/github/alexZaicev/go-ftp-client)
[![CodeQL](https://github.com/alexZaicev/go-ftp-client/actions/workflows/codeql.yaml/badge.svg)](https://github.com/alexZaicev/go-ftp-client/actions/workflows/codeql.yaml)
[![Trivy](https://github.com/alexZaicev/go-ftp-client/actions/workflows/trivy.yaml/badge.svg)](https://github.com/alexZaicev/go-ftp-client/actions/workflows/trivy.yaml)
[![Functional Tests](https://github.com/alexZaicev/go-ftp-client/actions/workflows/functional-tests.yaml/badge.svg)](https://github.com/alexZaicev/go-ftp-client/actions/workflows/functional-tests.yaml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A simple CLI utility for executing FTP client without then need of installing additional software.

## Backlog

- **Commands supported:**
  - [x] Status
  - [x] Make dir
  - [x] List
  - [x] Remove file/directory
  - [x] Move file/directory
  - [x] Upload file/directory
  - [x] Download file/directory
  - [ ] TLS support
  - [ ] Support configuration file
  - [ ] Support env variables
  - [ ] Add repos/input validator functions
  - Entry Parsers:
    - [ ] Implement missing test cases and features in UNIX and RFC3659 parsers
    - [ ] MsDOS entry parser
    - [ ] Hosted FTP entry parser
  - [ ] Missing functional tests
  - [ ] Support missing FTP server features upon discovery
  - [ ] Mechanism to disable features within config

AUTH TLS
CCC
CLNT
EPRT
EPSV
HOST
LANG fr-FR.UTF-8;fr-FR;en-US.UTF-8;en-US;it-IT.UTF-8;it-IT;es-ES.UTF-8;es-ES;bg-BG.UTF-8;bg-BG;ko-KR.UTF-8;ko-KR;zh-TW.UTF-8;zh-TW;ru-RU.UTF-8;ru-RU;ja-JP.UTF-8;ja-JP;zh-CN.UTF-8;zh-CN
MDTM
MFF modify;UNIX.group;UNIX.mode;
MFMT
MLST modify*;perm*;size*;type*;unique*;UNIX.group*;UNIX.groupname*;UNIX.mode*;UNIX.owner*;UNIX.ownername*;
PBSZ
PROT
RANG STREAM
REST STREAM
SIZE
SSCN
TVFS
UTF8
