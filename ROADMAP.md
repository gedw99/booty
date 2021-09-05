# ROADMAP

NOTE: This is incomplete and is running notes for changes to booty that are needed for the next version.

## Why

Read me to explain why ?

For dev,ci, ops all in one tools.

Ensures that you can do to a git repo and quickly know all the moving parts.

## Config

Booty config in your repo allows multi versioning and booty itself is versioned.

## File system

Installs on global paths and does not assume golang is installed.
Todo: where is cache ?
Todo: where it places binaries needs a sub folder called booty so it’s easy to see what is installed separate from what the OS has already installed in the global path.

## Extending

On your project use what makes sense but if you notice your using it across projects and repos it handy to add it to booty.


## Install

Use the Bingo approach to include a version of booty in any project repo.
See: https://github.com/grafana/grafana/tree/main/.bingo

On server with no golang use install.ah curl

## Make files
Include my currrent ones

Git:
Use make , not the golang ones !
Todo: upgrade booty git ( gw ) to replace the make files as they provide the neede multi branch and remotes.
Provides ability to have many branches and remotes to allow working in a team using conventions.
Extend to support pulling a PR to local to make it easy to review pr’s.

## Remote config

Add vault and vault config tooling if it works out.

The env config is saved in your git repo but is encrypted allowing you to have that saved in your repo. You can have staging and production etc there too.

So you can use remote config and embed your production config inside your binary and it will at runtime ask vault for the real config one unsealed over ssh. If the server reboots it needs to be unsealed.

## DB

Add lite stream if it works out. It seems to be a great global pattern for Projects if you can stomach modernc sqlite.

## Identity

Add vault users, groups and permissions if it works out.
Still don’t have good example yet. But vault is really liking good

## Deployment

Booty has no tooling for deployment except ssh. 
This is on purpose because it is independent on k8, docker, etc. Instead it is designed to deploy inside the Remote OS itself. So you can run booty via ssh inside an already booted server to add binaries or upgrade the binaries.

Your own project binary can use the same ssh deployment technique

Todo: if Vault works out use vault for ssh auth.

TODO: The config, in your rep,o updates to tell you what you deployed where so it’s saved in your GitHub repo as an audit trail.