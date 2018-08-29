# Realizing Service Function Chain and Group Base Policy in HAProxy

## Option 1: stick table
declare a [stick-table](https://cbonte.github.io/haproxy-dconv/1.8/configuration.html#4.2-stick-table%20type) for the frontend
```
stick-table type ip size 1m expire 5m store gpc0
```
notice that the final parameter:
  - `gpc0` : first General Purpose Counter. It is a positive 32-bit integer
    integer which may be used for anything. Most of the time it will be used
    to put a special tag on some entries, for instance to note that a
    specific behavior was detected and must be known for future matches.

```
global
    # enable Unix Socket commands (runtime API)
    stats socket /var/run/haproxy.sock mode 600 level admin
    # enbale another runtime API socket listening to a TCP port (dangerous) 
    stats socket ipv4@192.168.0.1:9999 level admin
    stats timeout 2m
frotend fe
    log global
    log-format "%ci:%cp [%t] %ft %b/%s %Tq/%Tw/%Tc/%Tr/%Tt %ST %B %CC %CS %tsc %ac/%fc/%bc/%sc/%rc %sq/%bq %hr %hs {%[ssl_c_verify],%{+Q}[ssl_c_s_dn],%{+Q}[ssl_c_i_dn]} %{+Q}r"

    # declare the table to store all incomming source IPs in 5 minutes, the maximum size is set to 1 million IPs
    stick-table type ip size 1m expire 5m store gpc0

    # use acl to define user role according to gpc0
    acl suspect src_get_gpc0 eq 1
    acl attacker src_get_gpc0 eq 2

    # for attacker, we can simply reject 
    tcp-request connection reject if attacker

    # for suspect, we can redirect them to a slower backend
    use_backend be_429_slow_down if suspect

    # any other user goes to normal backend
    default_backend be_app_stable

backend be_app_stable
    log global
    http-request replace-value Host .* mf-chsdi3.int.bgdi.ch
    server upstream_server mf-chsdi3.int.bgdi.ch:80 maxconn 512

 backend be_429_slow_down    
    timeout tarpit 2s
    errorfile 500 /var/local/429.http
    http-request tarpit
```
Then we can manipulate the `gpc0` through HAProxy [runtime API](https://cbonte.github.io/haproxy-dconv/1.8/management.html#9.3) [set table](https://cbonte.github.io/haproxy-dconv/1.8/management.html#9.3-set%20table) command, which does not require reload to take effect.

For example, we can analyse HAProxy frontend log using ML model on another server in real time. Once attack deteceted, block the cooresponding IP by setting `gpc0 = 2` through runtime API.

## Option 2: ACL
For each group, store all the IPs in a .lst file and call the file in configuration.
```
acl group-name src -f group-name.lst
```
Another runtime script should be used to operate the .lst file providing basic features like add, delete and expire. 

Each time the .lst file is changed, a HAProxy reload is needed to take effect.

Considering a IPv4 address normally takes 32 bytes to store, **every 30k (IP,role) pair will cost 1m file I/O for each reload**. This might not be negligible since we also need to reload very frequently.