# Examples

This folder contains a number of example configurations, service scripts, and packaging details may be useful to the end users.

## Example `nomnomlog` Configurations

### nomnomlog-config.example.yml

This is a simple configuration file example.  Use it as a template for your configuration.  This file should be placed at `/etc/nomnomlog-config.yml`.

### nomnomlog-config.example.advanced.yml

More advanced example of above using all the available yaml configuration options. These options *do not* reflect the default values.

## Service Configuration Scripts

### nomnomlog.init.d

This is an init.d script.  Use this if your system uses init.d for startup scripts.  Place this file at `/etc/init.d/nomnomlog` and then run `chmod +x /etc/init.d/nomnomlog`.  To start the service, run `service nomnomlog start` and to run on startup, run `update-rc.d nomnomlog defaults`.

### nomnomlog.supervisor.conf

This is a supervisor configuration file.

### nomnomlog.systemd.service

This is a systemd service configuration file.  Place this file at `/etc/systemd/system/nomnomlog.service` and then run `systemctl enable nomnomlog.service` to enable the service and `systemctl start nomnomlog.service` to start it.

### nomnomlog.upstart.conf

This is an upstart configuration file.  Place this file at `/etc/init/nomnomlog.conf` and then run `sudo start nomnomlog` to start the service.

## Packaging Configurations

## com.papertrailapp.nomnomlog.plist

This is an example Mac OS X plist file.  This file should be placed at `/Library/LaunchDaemons/com.papertrailapp.nomnomlog.plist`.
