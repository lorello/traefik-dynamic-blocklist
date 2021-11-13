# Dynamic IPs Blacklist for Traefik 

This small software implements the possibility to block access to all the services
behind your Traefik service.

The basic idea was explained in [this post](https://scaleup.us/2020/06/21/how-to-block-ips-in-your-traefik-proxy-server/): 
I basically implemented a small API to make the blocklist dynamic.

## How it runs

This small API can be used behind a Traefik web router to block a list of IPs,
for example list of remote hosts that are trying to violate the integrity of
your applications.

This service must be configured as ForwardAuthentication for all websites
that you want to protect.

The list is dynamic so that you can easly add/check/remove IPs to this list
without the need to restart anything.

## How to populate list of IPs to be blocked

Extract attackers IPs from traefik logs itself: a lot of times in my logs I
find remote hosts checking for URLs like 'admin.asp' (and I've no asp applications)
of `myadmin.php` or `/phpMyAdmin` but I don't use phpmyadmin on my server.
All those IPs are good candidates to be added to a blacklist.


