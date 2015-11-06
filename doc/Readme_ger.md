# Nagios Plugin zur Überwachung von Apache mod_proxy Balancer Pools

Das Plugin ***check_httpd_balancer*** funktioniert ähnlich wie das zum Standard-Umfang gehörende check_http. Es erwartet als URL eine Balancer-Manager-Seite (normalerweise /balancer-manager). Diese Seite wird analysiert und es wird versucht, den Status der hier enthaltenen Pools festzustellen. Dabei ist wichtig, dass es eine Regel in Bezug auf die verwendeten Worker (Hostnamen oder IP-Adressen) und den jvmRouten gibt. Dies kann dadurch erreicht werden, dass ein Worker, der in verschiedenen Pools immer wieder auftaucht, bei der jvmRoute immer die gleiche Endung bekommt.

Der Zusammenhang zwischen Worker-Adresse und jvmRoute-Endung ist das so genannte **Mapping**, welches auf der Kommandozeile angegeben werden muss. Das Mapping besteht aus der Worker-Adresse (ohne Port) und einem Suffix der jvmRoute, getrennt durch einen Doppelpunkt.

Bei dem folgenden Test-Pool
        ajp://192.168.0.11:8009	content01
        ajp://192.168.0.12:8009	content02
wäre ein Mapping beispielsweise: **` -M “192.168.0.11:01 192.168.0.12:02“`**  
Dabei ist nur wichtig, dass ein hinreichend langer und eindeutiger Suffix der jvmRoute angegeben wird, **` -M “192.168.0.11:tent01 192.168.0.12:tent02“`** würde genauso gut funktionieren.

Falls bei einem Pool die Zuordnung von Worker zu jvmRoute von dem vorgegebenen Mapping abweicht, wird der Pool als „gestört“ markiert. Das gleiche gilt, wenn *alle* Worker eines Pools im Fehlerzustand sind (also *nicht* den Status „Init OK“ haben).

Falls ein einzelner Worker in einem Fehlerzustand ist oder nicht im Mapping gefunden wird, wird er als defekt markiert.

Für die tolerierte Anzahl defekter Worker pro Pool kann mit **-w NN** ein Wert in Prozent angegeben werden (Default: 50). Eine kritische Grenze für diese Anzahl wird mit **-c NN** festgelegt (Default: 75).

### Alarme, Warnungen und Return-Codes ###
Der Zustand der Balancer-Pools wird zum Einen über die Anzahl funktionierender Worker definiert, zum Andern darüber, ob alle Worker dem vorgegebenen Mapping gehorchen. Je nach Zustand der Pools gibt das Plugin eine entsprechende Rückmeldung:
* CRITICAL (Return-Code 2):  
Falls die Zuordnung von Workern zu jvmRouten durcheinander geraten ist, so wird dieser Zustand sofort als kritisch definiert, da hier Nutzer Sessions verlieren oder in fremde Session einspringen können. Das Gleiche gilt, wenn die Anzahl defekter Worker die kritische Grenze (-c) übersteigt oder gar kein funktionierender Worker vorhanden ist. Der oder die betroffenen Pools werden angezeigt.
* WARNING (Return-Code 1):  
Übersteigt die Anzahl der defekten Worker in einem beliebigen Pool den entsprechenden Grenzwert (-w), so eine Warnung ausgegeben, zusammen mit dem oder den betroffenen Pools.  
* OK (Return-Code 0):  
Sofern alle funktionierenden Worker dem Mapping entsprechen und die Anzahl der defekten Worker unterhalb der Warnschwelle (-w) liegt, ist der Zustand zufriedenstellend. Dies muss für alle Pools gelten.

In allen anderen Fällen und Fehlern wird der Return-Code 3 (Unknown) ausgegeben. Dies gilt z.B., wenn der Zugriff auf die Webseite verboten ist oder die angegebene URL nicht auf eine Balancer-Manager-Seite verweist.

### Benutzung des Plugins: ###
Das Plugin folgt den Vorgaben für Nagios-Plugins für die Standard-Optionen und Return-Codes. Einziger Unterschied ist, dass es entgegen der Empfehlung keine mehrfachen „-v“-Optionen zur Erhöhung des Verbosity-Levels gibt. Dafür existiert eine Debug-Ausgabe (-d), welche erweiterte Informationen anzeigt.

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

Der Aufruf zur Prüfung der Balancer-Pools auf einer angenommenen  web1.example.com mit drei angeschlossenen Tomcats wäre demnach:
    $ check_httpd_balancer -M „172.16.1.1:01 172.16.1.2:02 172.16.1.3:03“ -H web1.example.com -u /balancer-manager

Alternativ können die Werte für Host, Post, SSL und vor allem die Workermap auch aus einer Konfigurationsdatei gelesen werden. Das Format ist:
	{
	  "Host": String,
	  "UseSSL": Boolean,
	  "URL": String,
	  "Port": String,
	  "WorkerMap": [Array of Strings]
	}

Eine Beispieldatei finden Sie im Verzeichnis „/etc“. Damit reduziert sich der Aufruf zu:
    $ check_httpd_balancer -C config.json
