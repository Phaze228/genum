# genum #


## Features ##

DNS Enumeration
```bash
genum dns -d google.com -t A,MX
genum dns -d zonetransfer.me 
```
Example Output -- 
```bash
[DNS ENUMERATION]
 Domain: zonetransfer.me
 Nameserver: 8.8.8.8
 Time Start: Sun, 2025-01-05 13:21:15

------------[PROGRESS]---------------------
[ Record Check Results ]
  [ SOA ]
  |_____zonetransfer.me.        4102    IN      SOA     nsztm1.digi.ninja. robin.digi.ninja. 2019100801 172800 900 1209600 3600

  [ NS ]
  |     zonetransfer.me.        4102    IN      NS      nsztm1.digi.ninja.
  |_____zonetransfer.me.        4102    IN      NS      nsztm2.digi.ninja.

  [ A ]
  |_____zonetransfer.me.        3289    IN      A       5.196.105.14

  [ MX ]
  |     zonetransfer.me.        3831    IN      MX      10 ALT2.ASPMX.L.GOOGLE.COM.
  |     zonetransfer.me.        3831    IN      MX      20 ASPMX3.GOOGLEMAIL.COM.
  |     zonetransfer.me.        3831    IN      MX      20 ASPMX5.GOOGLEMAIL.COM.
  |     zonetransfer.me.        3831    IN      MX      0 ASPMX.L.GOOGLE.COM.
  |     zonetransfer.me.        3831    IN      MX      20 ASPMX4.GOOGLEMAIL.COM.
  |     zonetransfer.me.        3831    IN      MX      20 ASPMX2.GOOGLEMAIL.COM.
  |_____zonetransfer.me.        3831    IN      MX      10 ALT1.ASPMX.L.GOOGLE.COM.

  [ TXT ]
  |_____zonetransfer.me.        301     IN      TXT     "google-site-verification=tyP28J7JAUHA9fw2sHXMgcCC0I6XBmmoVi04VlMewxA"

  [ HINFO ]
  |_____zonetransfer.me.        300     IN      HINFO   "Casio fx-700G" "Windows XP"

[ Zone Transfer Results ]
[------ internal.zonetransfer.me.@nsztm1.digi.ninja. ------]
  [ SOA ]
  |     internal.zonetransfer.me.       7200    IN      SOA     intns1.zonetransfer.me. robin.digi.ninja. 2014101601 172800 900 1209600 3600
  |_____internal.zonetransfer.me.       7200    IN      SOA     intns1.zonetransfer.me. robin.digi.ninja. 2014101601 172800 900 1209600 3600

  [ NS ]
  |_____internal.zonetransfer.me.       300     IN      NS      intns1.zonetransfer.me.

  [ A ]
  |     cisco1.internal.zonetransfer.me.        300     IN      A       10.1.1.254
  |     cisco2.internal.zonetransfer.me.        300     IN      A       10.1.1.253
  |     dc.internal.zonetransfer.me.    300     IN      A       10.1.1.1
  |     fileserv.internal.zonetransfer.me.      300     IN      A       10.1.1.4
  |_____mail.internal.zonetransfer.me.  300     IN      A       10.1.1.3

```

SMTP Enumeration
```bash
genum smtp -U <user or file list> -d <domain> 

```
