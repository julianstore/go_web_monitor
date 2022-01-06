### Step

1. Download “go installer” and install on your machine.
2. Open VPN.
3. Go to “web-monitor” directory.
4. Open terminal.
5. Run “go build” command to make executable file.
6. Run web-monitor executable file on terminal.
   The time interval of this monitor is 1 hour.
   “config.conf” file is config file to run this web-monitor. The website url which have to be monitored, vcenter ip, venter username & password, download baseurl are saved as json type in this file.
   You have to place this config file in the same directory as executable file.
7. You can see the logs of web-monitor and find out downloaded ova file in your directory.
   If Error occurs , the error message will be logged.
8. After uploading to vcenter finished, you could open browser and go to vcenter.
9. Log in to VCenter, you could find the content library and ovf item that created.

### Dependencies

github.com/cavaliercoder/grab v2.0.0+incompatible
github.com/cpuguy83/go-md2man/v2 v2.0.1
github.com/dustin/go-humanize v1.0.0
github.com/fatih/color v1.13.0
github.com/hashicorp/go-multierror v1.1.1
github.com/mattn/go-isatty v0.0.14
github.com/mattn/go-runewidth v0.0.13
github.com/mgechev/revive v1.1.2
github.com/pkg/errors v0.9.1
github.com/pmezard/go-difflib v1.0.0
golang.org/x/crypto v0.0.0-20211202192323-5770296d904e
golang.org/x/net v0.0.0-20211205041911-012df41ee64c
golang.org/x/sys v0.0.0-20211205182925-97ca703d548d
golang.org/x/term v0.0.0-20210927222741-03fcf44c2211
golang.org/x/tools v0.1.8
honnef.co/go/tools v0.2.2

### Config

The constants for config are stored in “config.conf” file with JSON type.
If you want to change the config , open “config.conf” file and find the fields and replace the values as you want.
Fields and current values are below;
Fields | Value | Description
--- | --- | ---
websiteurl | https://wiki.ubuntu.com/Releases?_ga=2.179241575.2128997974.1638430885-220022704.1638430885 | the url of web site of which this web-monitor watches
downloadbaseurl | https://cloud-images.ubuntu.com/ | the base url of download file
Timeinterval | 1 | Timer interval for monitoring( unit: Hour)
vcenterip | https://51.255.152.252 | Ip address vcenter server
vcenterusername | administrator@vsphere.local | User name for authenticating
vcenteruserpwd | %p:)GBJ23\*d/T<KCCMMJ | User password for authenticating
