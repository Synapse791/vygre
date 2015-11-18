# Vygre
**Start new docker containers when your old ones die**

### Overview
Vygre is a service designed to ensure that docker containers are always running. Maybe you have a web service running in a docker container and you want to make sure that if for some reason the container dies, your site doesn't go down. Vygre will spot that container isn't running and start up another based on your configuration.
The program reads JSON config files describing your containers and creates them accordingly. In these files you can specify the image, ports, environment variables, volumes, and how many instances of the container there should be.

### Vygre configuration
The location for the JSON configuration file is `/etc/vygre/config.json`

#### Structure
```json
{
  "log_level"       : "warning",
  "check_interval"  : 5,
  "auth"            : {
    "serveraddress"   : "https://myregistry.net",
    "username"        : "registry_user",
    "password"        : "Registry_user_password",
    "email"           : "registry_user_email"
  }
}
```

Parameter          | Description                                                                                                                 | Default
-------------------|-----------------------------------------------------------------------------------------------------------------------------|---------
log_level          | level of log output messages to display. This parameter will be overwritten by the `-d` flag. [debug\|info\|warning\|error] | warning
check_interval     | number of seconds between checks                                                                                            | 5
auth               | authorization object for private docker registry                                                                            | nil
auth.serveraddress | address, including scheme, of the private docker registry to authenticate against                                           | nil
auth.username      | username to use for authenticating                                                                                          | nil
auth.password      | password to use for authenticating                                                                                          | nil
auth.email         | email to use for authenticating                                                                                             | nil

### Container Configuration
The container configuration files should be stored in `/etc/vygre/conf.d/` in JSON format

#### Structure
```json
{
  "image": "myregistry.net/my-app:latest",
  "instances": 2,
  "env": {
    "MY_APP_ENV" : "production"
  },
  "ports": [
    "3000",
    "80:80",
    "10.100.1.1:443:443"
  ],
  "volumes": [
    "/var/cache/my-app:/var/cache/my-app",
    "/tmp:/tmp/host_tmp:ro"
  ]
}
```

Parameter | Description
----------|-------------------------------------------------------------------------------------------------------
image     | docker image name, including registry and tag if required
instances | the number of instances you require running. If specifying host ports, this **must** be set to 1
env       | map environment variables for the container as key value pairs
ports     | array of ports to expose. If no host port is specified, a random host port will be assigned
volumes   | array of volumes to mount in the format of `host_path:container_path`. Optional `:ro` or `:rw` suffix