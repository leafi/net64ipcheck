# Open IP/port tester

Service to tell a user their IP & whether their server port is open.

## Usage

Server: Run the binary. It listens on port :6460 UDP.

## Protocol

Consider a user attempting to start a server on port 8000. Their IP is 1.2.3.4.

aaaa, bbbb etc. are dynamically chosen/irrelevant ports.

**PAY ATTENTION TO THE PORT DANCE.**

The first UDP packet sent to api.?????.com:6460 must be 400 bytes.

It must contain the UTF-8 string `net64ipc0000`, followed by 2 bytes encoding the server port (typically 8000), and then the remaining 386 bytes can be anything.

**Wrong version scenario**

1.2.3.4:aaaa -> api.?????.com:6460   (sends garbage that's 400 bytes long)

api.?????.com:bbbb -> 1.2.3.4:aaaa   `BADVER`

**Correct version but port closed**

1.2.3.4:aaaa -> api.?????.com:6460    `net64ipc0000(binary short representing 8000)`

api.?????.com:bbbb -> 1.2.3.4:8000    `TEST`

(... packet never gets through to 1.2.3.4:8000 ...)

Session ends. Client should rely on timeout to detect failure. Service will wait **2 seconds** for a reply.

The client should consider sending 3 attempts 500ms apart for every "test" to deal with network flakiness lost packets etc.

**Correct version and port open**

1.2.3.4:aaaa -> api.?????.com:6460    `net64ipc0000(binary short representing 8000)`

api.?????.com:bbbb -> 1.2.3.4:8000    `TEST`

1.2.3.4:8000 -> api.?????.com:bbbb    `TOAST000`

api.?????.com:cccc -> 1.2.3.4:aaaa    `OK 1.2.3.4`

Session ends.

As UDP is lossy, the client should rely on timeout to detect failure but should probably send out 3 requests spaced 500ms apart or so & if any OK reply comes back then consider the port open.

1.2.3.4:8000 must send back TOAST000 within **2 seconds**.

(If 3 (or 5, whatever) attempts aren't enough or packet takes >2s, the user's connection is so shit we may as well say "hey, this ain't gonna work".)
