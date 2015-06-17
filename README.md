## check_httpd_balancer ##
This plugin monitors the state of an apache balancer pool established by *mod_proxy*. Due to the fact that there’s still no really working xml api, the plugin parses the /balancer-manager status page with some RegEx.

The main reason for developing this plugin is the bad behaviour of tomcats as backend systems even with deactivated failover settings for *mod_proxy_ajp*, which sometimes lead to a quite complete confusion regarding the *jvmRoutes*, so some really ugly things happen as session mixing or session loss. 

It is written in [**Go**](https://golang.org/project/), which is my preferred and really beloved language these days. I’m not connected to Google in any other way. :-)

I’ve put this plugin under the GPLv2 with the kind permission of my employer [**denkwerk GmbH**](http://www.denkwerk.com), because some parts of the sources were coded at work (and for work, of course :-)).

A more detailed Readme (in [english](https://github.com/hleinders/check_httpd_balancer/blob/master/doc/Readme_en.md) or [german](https://github.com/hleinders/check_httpd_balancer/blob/master/doc/Readme_ger.md)) can be found in the docs folder.

I’ve provided some binary builds in the bin folder, for Linux (64 and 32 bit) and an Apple darwin 64 bit executable.