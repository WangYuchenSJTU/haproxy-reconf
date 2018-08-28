# Haproxy

## Introduction

### What Haproxy is
- a TCP proxy : it can accept a TCP connection from a listening socket, connect to a server and attach these sockets together allowing traffic to flow in both directions.
- an HTTP reverse-proxy (called a "gateway" in HTTP terminology) : it presents itself as a server, receives HTTP requests over connections accepted on a listening TCP socket, and passes the requests from these connections to servers using different connections.
- a TCP normalizer : since connections are locally terminated by the operating system, there is no relation between both sides, so abnormal traffic such as invalid packets, flag combinations, window advertisements, sequence numbers, incomplete connections (SYN floods), or so will not be passed to the other side. This protects fragile TCP stacks from protocol attacks, and also allows to optimize the connection parameters with the client without having to modify the servers' TCP stack settings.
- **an HTTP normalizer** : when configured to process HTTP traffic, only valid complete requests are passed. This protects against a lot of protocol-based attacks. Additionally, protocol deviations for which there is a tolerance in the specification are fixed so that they don't cause problem on the servers (e.g. multiple-line headers).
- **a content-based switch** : it can consider any element from the request to decide what server to pass the request or connection to. Thus it is possible to handle multiple protocols over a same port (e.g. HTTP, HTTPS, SSH).
- **a protection against DDoS and service abuse** : **it can maintain a wide number of statistics per IP address, URL, cookie, etc and detect when an abuse is happening, then take action (slow down the offenders, block them, send them to outdated contents, etc).**
- Others: an SSL terminator / initiator / offloader; an HTTP fixing tool; a server load balancer; a traffic regulator; an observation point for network troubleshooting; an HTTP compression offloader.

### How Haproxy works
- single-threaded, event-driven, non-blocking engine combining a very fast I/O layer with a priority-based scheduler
- layered model offering bypass mechanisms at each level ensuring data doesn't reach higher levels unless needed
- A single process can run many proxy instances (300000 at least)
- only requires the haproxy executable and a configuration file to run

Once HAProxy is started, it does exactly 3 things :

  - process incoming connections;

  - periodically check the servers' status (known as health checks);

  - exchange information with other haproxy nodes.

Processing incoming connections is by far the most complex task as it depends
on a lot of configuration possibilities, but it can be summarized as the 9 steps
below :

  - accept incoming connections from listening sockets that belong to a
    configuration entity known as a "frontend", which references one or multiple
    listening addresses;

  - apply the frontend-specific processing rules to these connections that may
    result in blocking them, modifying some headers, or intercepting them to
    execute some internal applets such as the statistics page or the CLI;

  - pass these incoming connections to another configuration entity representing
    a server farm known as a "backend", which contains the list of servers and
    the load balancing strategy for this server farm;

  - apply the backend-specific processing rules to these connections;

  - decide which server to forward the connection to according to the load
    balancing strategy;

  - apply the backend-specific processing rules to the response data;

  - apply the frontend-specific processing rules to the response data;

  - emit a log to report what happened in fine details;

  - in HTTP, loop back to the second step to wait for a new request, otherwise
    close the connection.

## Important features 
### Non-related : Proxying / SSL / Monitoring / High availability / Load balancing / Content switching / HTTP rewriting and redirection / Statistics
### Sampling and converting information
HAProxy supports information sampling using a wide set of "sample fetch
functions". The principle is to extract pieces of information known as samples,
for immediate use. This is used for stickiness, to build conditions, to produce
information in logs or to enrich HTTP headers.

Samples can be fetched from various sources :

  - constants : integers, strings, IP addresses, binary blocks;

  - the process : date, environment variables, server/frontend/backend/process
    state, byte/connection counts/rates, queue length, random generator, ...

  - variables : per-session, per-request, per-response variables;

  - the client connection : source and destination addresses and ports, and all
    related statistics counters;

  - the SSL client session : protocol, version, algorithm, cipher, key size,
    session ID, all client and server certificate fields, certificate serial,
    SNI, ALPN, NPN, client support for certain extensions;

  - request and response buffers contents : arbitrary payload at offset/length,
    data length, RDP cookie, decoding of SSL hello type, decoding of TLS SNI;

  - **HTTP (request and response)** : method, URI, path, query string arguments,
    status code, headers values, positional header value, cookies, captures,
    authentication, body elements;

A sample may then pass through a number of operators known as "converters" to
experience some transformation. A converter consumes a sample and produces a
new one, possibly of a completely different type. For example, a converter may
be used to return only the integer length of the input string, or could turn a
string to upper case. Any arbitrary number of converters may be applied in
series to a sample before final use. Among all available sample converters, the
following ones are the most commonly used :

  - arithmetic and logic operators : they make it possible to perform advanced
    computation on input data, such as computing ratios, percentages or simply
    converting from one unit to another one;

  - IP address masks are useful when some addresses need to be grouped by larger
    networks;

  - data representation : URL-decode, base64, hex, JSON strings, hashing;

  - string conversion : extract substrings at fixed positions, fixed length,
    extract specific fields around certain delimiters, extract certain words,
    change case, apply regex-based substitution;

  - date conversion : convert to HTTP date format, convert local to UTC and
    conversely, add or remove offset;

  - lookup an entry in a stick table to find statistics or assigned server;

  - map-based key-to-value conversion from a file (mostly used for geolocation).

### Maps

Maps are a powerful type of converter consisting in loading a two-columns file
into memory at boot time, then looking up each input sample from the first
column and either returning the corresponding pattern on the second column if
the entry was found, or returning a default value. The output information also
being a sample, it can in turn experience other transformations including other
map lookups. Maps are most commonly used to translate the client's IP address
to an AS number or country code since they support a longest match for network
addresses but they can be used for various other purposes.

Part of their strength comes from being updatable on the fly either from the CLI
or from certain actions using other samples, making them capable of storing and
retrieving information between subsequent accesses. Another strength comes from
the binary tree based indexation which makes them extremely fast even when they
contain hundreds of thousands of entries, making geolocation very cheap and easy
to set up.

### ACLs and conditions

Most operations in HAProxy can be made conditional. Conditions are built by
combining multiple ACLs using logic operators (AND, OR, NOT). Each ACL is a
series of tests based on the following elements :

  - a sample fetch method to retrieve the element to test ;

  - an optional series of converters to transform the element ;

  - a list of patterns to match against ;

  - a matching method to indicate how to compare the patterns with the sample

For example, the sample may be taken from the HTTP "Host" header, it could then
be converted to lower case, then matched against a number of regex patterns
using the regex matching method.

Technically, ACLs are built on the **same core as the maps**, they share the exact
same internal structure, pattern matching methods and performance. The only real
difference is that **instead of returning a sample, they only return "found" or
or "not found"**. In terms of usage, ACL patterns may be declared inline in the
configuration file and do not require their own file. ACLs may be named for ease
of use or to make configurations understandable. A named ACL may be declared
multiple times and it will evaluate all definitions in turn until one matches.

About 13 different pattern matching methods are provided, among which IP address
mask, integer ranges, substrings, regex. They work like functions, and just like
with any programming language, only what is needed is evaluated, so when a
condition involving an OR is already true, next ones are not evaluated, and
similarly when a condition involving an AND is already false, the rest of the
condition is not evaluated.

There is **no practical limit to the number** of declared ACLs, and a handful of
commonly used ones are provided. However experience has shown that setups using
a lot of named ACLs are quite hard to troubleshoot and that sometimes using
anonymous ACLs inline is easier as it requires less references out of the scope
being analyzed.

### Stick-tables

Stick-tables are commonly used to store stickiness information, that is, to keep
a reference to the server a certain visitor was directed to. The key is then the
identifier associated with the visitor (its source address, the SSL ID of the
connection, an HTTP or RDP cookie, the customer number extracted from the URL or
from the payload, ...) and the stored value is then the server's identifier.

Stick tables may use 3 different types of samples for their keys : integers,
strings and addresses. Only one stick-table may be referenced in a proxy, and it
is designated everywhere with the proxy name. Up to 8 keys may be tracked in
parallel. The server identifier is committed during request or response
processing once both the key and the server are known.

Stick-table contents may be replicated in active-active mode with other HAProxy
nodes known as "peers" as well as with the new process during a reload operation
so that all load balancing nodes share the same information and take the same
routing decision if client's requests are spread over multiple nodes.

Since stick-tables are **indexed on what allows to recognize a client**, they are
often **also used to store extra information such as per-client statistics**. The
extra statistics take some extra space and need to be explicitly declared. The
type of statistics that may be stored includes the input and output bandwidth,
the number of concurrent connections, the connection rate and count over a
period, the amount and frequency of errors, some **specific tags** and counters,
etc. In order to support keeping such information without being forced to
stick to a given server, a special "tracking" feature is implemented and allows
to track up to 3 simultaneous keys from different tables at the same time
regardless of stickiness rules. Each stored statistics may be searched, dumped
and cleared from the CLI and adds to the live troubleshooting capabilities.

While this mechanism can be used to surclass a returning visitor or to adjust
the delivered quality of service depending on good or bad behavior, **it is
mostly used to fight against service abuse and more generally DDoS as it allows
to build complex models to detect certain bad behaviors at a high processing
speed.**

### Formatted strings
There are many places where HAProxy needs to manipulate character strings, such
as logs, redirects, header additions, and so on. In order to provide the
greatest flexibility, the notion of Formatted strings was introduced, initially
for logging purposes, which explains why it's still called "log-format". These
strings contain escape characters allowing to introduce various dynamic data
including variables and sample fetch expressions into strings, and even to
adjust the encoding while the result is being turned into a string (for example,
adding quotes). This provides a powerful way to build header contents or to
customize log lines. Additionally, in order to remain simple to build most
common strings, about 50 special tags are provided as shortcuts for information
commonly used in logs.

### Server protection
HAProxy does a lot to maximize service availability, and for this it takes
large efforts to protect servers against overloading and attacks. The first
and most important point is that **only complete and valid requests are forwarded**
to the servers. The initial reason is that HAProxy needs to find the protocol
elements it needs to stay synchronized with the byte stream, and the second
reason is that until the request is complete, there is no way to know if some
elements will change its semantics. The direct benefit from this is that servers
are not exposed to invalid or incomplete requests. This is a very effective
protection against slowloris attacks, which have almost no impact on HAProxy.

Another important point is that HAProxy contains buffers to store requests and
responses, and that by only sending a request to a server when it's complete and
by reading the whole response very quickly from the local network, the server
side connection is used for a very short time and this preserves server
resources as much as possible.

A direct extension to this is that HAProxy can artificially limit the number of
concurrent connections or outstanding requests to a server, which guarantees
that the server will never be overloaded even if it continuously runs at 100% of
its capacity during traffic spikes. All excess requests will simply be queued to
be processed when one slot is released. In the end, this huge resource savings
most often ensures so much better server response times that it ends up actually
being faster than by overloading the server. Queued requests may be redispatched
to other servers, or even aborted in queue when the client aborts, which also
protects the servers against the "reload effect", where each click on "reload"
by a visitor on a slow-loading page usually induces a new request and maintains
the server in an overloaded state.

The slow-start mechanism also protects restarting servers against high traffic
levels while they're still finalizing their startup or compiling some classes.

Regarding the protocol-level protection, it is possible to relax the HTTP parser
to accept non standard-compliant but harmless requests or responses and even to
fix them. This allows bogus applications to be accessible while a fix is being
developed. In parallel, offending messages are completely captured with a
detailed report that help developers spot the issue in the application. The most
dangerous protocol violations are properly detected and dealt with and fixed.
For example malformed requests or responses with two Content-length headers are
either fixed if the values are exactly the same, or rejected if they differ,
since it becomes a security problem. Protocol inspection is not limited to HTTP,
it is also available for other protocols like TLS or RDP.

**When a protocol violation or attack is detected, there are various options to
respond to the user, such as returning the common "HTTP 400 bad request",
closing the connection with a TCP reset, or faking an error after a long delay
("tarpit") to confuse the attacker.** All of these contribute to protecting the
servers by discouraging the offending client from pursuing an attack that
becomes very expensive to maintain.

HAProxy also proposes some more advanced options to protect against accidental
data leaks and session crossing. Not only it can log suspicious server responses
but it will also log and optionally block a response which might affect a given
visitors' confidentiality. One such example is a cacheable cookie appearing in a
cacheable response and which may result in an intermediary cache to deliver it
to another visitor, causing an accidental session sharing.

### Logging
Logging is an extremely important feature for a load balancer, first because a
load balancer is often wrongly accused of causing the problems it reveals, and
second because it is placed at a critical point in an infrastructure where all
normal and abnormal activity needs to be analyzed and correlated with other
components.

HAProxy provides very detailed logs, with millisecond accuracy and the exact
connection accept time that can be searched in firewalls logs (e.g. for NAT
correlation). **By default, TCP and HTTP logs are quite detailed an contain
everything needed for troubleshooting**, such as source IP address and port,
frontend, backend, server, timers (request receipt duration, queue duration,
connection setup time, response headers time, data transfer time), global
process state, connection counts, queue status, retries count, detailed
stickiness actions and disconnect reasons, header captures with a safe output
encoding. It is then possible to extend or replace this format to include any
sampled data, variables, captures, resulting in very detailed information. **For
example it is possible to log the number of cumulative requests or number of
different URLs visited by a client.**

**The log level may be adjusted per request using standard ACLs**, so it is possible
to automatically silent some logs considered as pollution and instead raise
warnings when some abnormal behavior happen for a small part of the traffic
(e.g. too many URLs or HTTP errors for a source address). Administrative logs
are also emitted with their own levels to inform about the loss or recovery of a
server for example.

Each frontend and backend may use multiple independent log outputs, which eases
multi-tenancy. Logs are preferably sent over UDP, maybe JSON-encoded, and are
truncated after a configurable line length in order to guarantee delivery.

## Haproxy Management

A command line interface (CLI) is available as a UNIX or TCP socket, to perform
a number of operations and to retrieve troubleshooting information. Everything
done on this socket doesn't require a configuration change, so it is mostly used
for temporary changes. Using this interface it is possible to change a server's
address, weight and status, to consult statistics and clear counters, dump and
clear stickiness tables, possibly selectively by key criteria, dump and kill
client-side and server-side connections, dump captured errors with a detailed
analysis of the exact cause and location of the error, dump, add and remove
entries from ACLs and maps, update TLS shared secrets, apply connection limits
and rate limits on the fly to arbitrary frontends (useful in shared hosting
environments), and disable a specific frontend to release a listening port
(useful when daytime operations are forbidden and a fix is needed nonetheless).
