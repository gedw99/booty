![CI](https://github.com/amplify-edge/booty/workflows/CI/badge.svg)
[![codecov](https://codecov.io/gh/alexadhy/booty/branch/master/graph/badge.svg?token=VLMYJWAQWJ)](https://codecov.io/gh/alexadhy/booty)

# (WIP) Booty

### Spec & Design:

- [Issue 108](https://github.com/amplify-edge/main/issues/108)

## Installation

#### Linux / Mac (amd64)

`curl -fsSL https://raw.githubusercontent.com/amplify-edge/booty/master/scripts/install.sh | bash`

#### Windows (amd64) (Needs powershell)

`Invoke-WebRequest -useb https://raw.githubusercontent.com/amplify-edge/booty/master/scripts/install.ps1 | Invoke-Expression`

## Shell Completion

run

`booty completion`

or to write it to file (*nix / Darwin for example):

`booty completion > compl.bash`

it will generate your shell completion, refer to your shell documentation on how best to install it and source it on
your shell.


## Usage ( for users )

### Git

```bash
# clone your forked repro. 
# for this example the user is "johnsmith", and the repro is "shared".
git clone git@github.com-johnsmith:johnsmith/shared
cd shared

# setup git config to have remote section pointing to upstream repository.
booty gw reg share
booty gw fset amplify-edge

# make any change.
rm ./README.md

# add and commit.
booty gw add-all
booty gw commit 'blah'

# check if there were any changes on the shared remote upstream repo.
booty gw fup

# push from your local origin to your remote origin
booty gw push

# open the github repo and create a Push Request (PR) back to upstream as per normal.
open https://github.com/johnsmith/shared

```

## Usage (for devs)

1. Clone this repository
2. Run `make all`
3. Copy the binary in `bin` directory to your `PATH`
4. Run `booty`

## Tinker Tanker

To add a component, implement the `Component`
interface in [here](https://github.com/alexadhy/booty/blob/master/dep/component.go)

This program installs config under:

- Linux: `$HOME/.local/booty/etc`
- Mac: `$HOME/Library/Application\ Support/booty/etc`
- Windows: `C:\\ProgramData\booty\etc`

The file `config.reference.json` is provided as a reference should you want to change the program behaviour.

## Implemented

1. Download 3rd party binaries and libraries our project needs
2. Install said binaries
3. Also install service, managed by OS' service supervisor, preferably as user service on dev and as system-wide service
   running under unprivileged user on production.
4. Ability to run each service on the foreground like pingcap's `tiup playground`
5. Ability to self update and check for updates for third party binaries

## Short Term TODO

4. Manage configs and backups for `maintemplatev2`
