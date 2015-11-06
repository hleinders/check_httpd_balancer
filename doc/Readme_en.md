# Nagios Plugin to monitor Apache mod_proxy Balancer Pools

The plugin ***check_httpd_balancer*** works quite similar to the standard check_http. It expects an url to a balancer manager status page (normally /balancer-manager). This page is parsed and analyzed to determine the state of the containing pools. 

One imported point is the assignment of balancer workers to their *jvmRoute*. This mapping can’t be guessed by the plugin, so it must be given on the command line. If there are several pools found in balancer manager page, you have to define all valid mappings. The mapping consists of the name or ip address of the worker (as shown in the manager page) and a unique suffix of the jvmRoute string, separated by a colon „:“.

For example, if your pool consists of these workers with their jvmRoutes
        ajp://192.168.0.11:8009	content01
        ajp://192.168.0.12:8009	content02
the mapping would be: **` -M “192.168.0.11:01 192.168.0.12:02“`**  
Here „01“ and „02“ are the unique suffixes (you also could have used „tent01“ or even „content01“ instead of „01“) 

If in a pool this mapping does not fit, the offending worker is marked as broken. Because this may lead to session hopping for end users, this immediately leads to *critical* for the whole pool. 

A worker which isn’t in an „Init OK“ state is also broken. If the ratio of broken workers to total workers is greater or equal to the warning threshold **-w NN** (where NN is a percentage, e.g. *50* for 50%), the pool is marked as *warning*. If the ratio is greater or equal to the critical threshold **-c NN**, the pool is in a critical state.

### Alerts, Warnings and Return-Codes ###
The state of a balancer pool is defined by the number of broken workers and by the state of the worker mappings to their jvmRoutes. Depending on these criteria the plugin gives an appropriate feedback:
* CRITICAL (Return-Code 2):  
The mapping of at least one worker to its jvmRoute is distorted or the number of broken workers is larger or equal to the critical threshold, given by „-c“ (default: 75). 
The affected pools are shown.
* WARNING (Return-Code 1):  
The number of broken workers is larger or equal to the waring threshold „-w“ (default: 50).
The affected pools are shown.
* OK (Return-Code 0):  
The mapping of all workers in all pools is fine and the number of broken workers is less than the warning threshold.

In all other cased a return code of 3 (Unknown) is returned. This is especially true for connection errors, e.g. the balancer manager status page can’t be reached or the access is denied.

### Usage: ###
The plugin follows the [*Nagios Plugin Development Guidelines*](https://nagios-plugins.org/doc/guidelines.html) with one exception: It doesn’t recognises the multiple usage of „-v“ to increase the verbosity level. There is a deug flag „-d“ instead. 

    $ check_httpd_balancer -h

    Version: 0.4
    Usage:   check_httpd_balancer [-h] [options] -H Hostname -M Mapping -u URL
    
    Options:
      -H=„localhost“: Host name
      -I=„127.0.0.1“: Host ip address (not implemented yet)
      -M=„192.168.0.1:01 192.168.0.2:02“: List of worker mappings (IP):(jvmRoute-suffix)
      -S=false: Connect via SSL. Port defaults to 443
      -V=false: Show version
			-C=„“: Read settings from config file (JSON)
      -a=„“: Basic Auth: password
      -c=75: Critical threshold for offline workers (in %)
      -d=false: Debug mode
      -l=„“: Basic Auth: user
      -n=false: Dry run
      -p=„“: TCP port
      -u=„/balancer-manager“: URL to check
      -v=false: Verbose mode
      -w=50: Warning threshold for offline workers (in %)

The call to an imaginary balancer pool status check of a server at web1.example.com with three backend tomcat workers with jvmRoutes „tc01“, „tc02“ and „tc03“ therefore would be:
    $ check_httpd_balancer -M „172.16.1.1:01 172.16.1.2:02 172.16.1.3:03“ -H web1.example.com -u /balancer-manager

Alternatively the values for Host, Post, SSL and Workermap can be read from a configuration file. The syntax is:
	{
	  "Host": String,
	  "UseSSL": Boolean,
	  "URL": String,
	  "Port": String,
	  "WorkerMap": [Array of Strings]
	}

An example json can be found in „/etc“. 
With a configuration file the call simplifies to:
    $ check_httpd_balancer -C config.json
