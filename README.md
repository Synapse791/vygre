# Vygre
**Keep your docker containers running**

### Overview
Vygre is a service designed to ensure that docker containers stay running. Maybe you have a web service running in a docker container and you want to make sure that if for some reason the container dies, your site doesn't go down. Vygre will spot that container isn't running and start up another.
The program reads JSON config files describing your containers and creates them accordingly. In these files you can specify the image, volumes, ports and how many instances of the container there should be.

### Application config
The default location for the config.json file is `/etc/vygre/config.json`. This can be altered using the `-c` flag and passing the filepath of the config.json file when running the program.

```sh
vygre -c /tmp/config.json
```

You can also test the application config using the `-t` flag.

```sh
vygre -c /tmp/config.json -t

# Or to test the default location
vygre -t
```

```json
{
  "install_dir"     : "/etc/vygre",
  "docker_endpoint" : "unix:///var/run/docker.sock",
  "check_interval"  : 5,
  "auth"            : {
    "username" : "groot",
    "password" : "letmein",
    "email"    : "groot@gmail.com"
  }
}
```
* **install_dir** - the location of the vygre config files. Inside this directory should be the config.json file and conf.d directory. *DEFAULT: /etc/vygre*
* **docker_endpoint** - The docker endpoint to query. *DEFAULT: unix:///var/run/docker.sock*
* **check_interval** - The interval to check containers in seconds *DEFAULT: 3*
* **auth** - The authorization information for docker (as if you ran `sudo docker login`) *DEFAULT: nil*

### Example container config file
Container configuration files should be stored in the conf.d directory inside the `install_dir` directory from the config.json file. 
```json
{
  "image": "mysql:latest",
  "instances": 1,
  "hostname": "app-db",
  "env": [
    "MYSQL_ROOT_PASSWORD=letmein"
  ],
  "ports": {
    "3306/tcp": [{"HostPort": "3306"}]
  },
  "volumes": [
    "/tmp/www:/usr/share/nginx/html"
  ]
}
```