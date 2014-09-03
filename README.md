Spool
=====

A docker container deployment tool. Define your services
with a simple config file and let spool create your environment.
Written in golang. For the time being, look at /test/project/spool.json for a sample config.

*Current Version: 0.1.0*

## Usage

```bash
# deploy an environment
spool up [env]

# stop an environment
spool stop [env]

# stop and remove an environment
spool destroy [env]

# help
spool
```

## Install

To install, ensure `$GOPATH/bin` is in your PATH and:

```bash
$ go get github.com/jeffturcotte/spool
```

## Development

There is a docker/go dev environment available as
a vagrant box. To run and connect to it:

```bash
$ vagrant up
$ vagrant ssh
```

Look at the Vagrantfile for more details.

## TODO

- Tests
- Image tag clean up
- Sync command
- Symlink command
- Config validation
- Port mapping option
- Persist option
- Env specific service overrides
- Custom cobra help template
- Tag pulled images
- Lots more

## Author

[Jeff Turcotte](https://github.com/jeffturcotte)
