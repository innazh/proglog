# Proglog
Commit logs - append-only data structure, sequenced by time.

Each handler consists of three steps:
1. Unmarshal the request's JSON body into a struct.
2. Complete that endpoint's logic with the request, obtain a result.
3. Marshal and write the result to the response.
If a handler becomes much more complicated than this, then move req & resp handling to middleware, business logic further down the stack.

Protobuf pros:
- type safety
- prevents schema violations
- fast serialization
- backward compatibility (new code can read old data structures)

The magic internal packages are used to restrict access to certain packages within a project. Packages inside an internal directory can only be imported by code within the parent directory or its subdirectories. This helps in encapsulating code and preventing it from being used outside its intended scope.

Types of gRPC Streaming RPCs:
Unary: Single request and single response.
Server streaming: Single request and multiple responses (streamed from server to client).
Client streaming: Multiple requests (streamed from client to server) and single response.
Bidirectional streaming: Both client and server send a sequence of messages using a read-write stream.
note: Rcv() is a blocking call, waits until a msg is received or the stream is closed.

Fav quote about securing an application: "Whenever I'm building a service, I think about what it'd be like if the data I'm trying to protect was publicly posted all over planet. Picturing this gives me the motivation to make sure that sort of thing doesn't happen to me, ..."

Secrutiy of distrtibuted services in three steps:
1. Encrypt the transmitted/in-flight data against MITM (man in the middle)
    - TLS - successor to SSL
    - Typically web services user one-way auth and only auth the server through the handshake that's initiated when the client and server connect. It's up to the app to auth the user
    - Certs for internal distributed systems don't need to come from a third party, one can operate a CA (cert authority) themself
2. Authentication to identify clients (who is who)
    - There's also a two-way auth or TLS mutual auth, whcih is used in machine-to-machine communication, or distributed systems. Both client and server use a cert to auth itself
3. Authorization to assign the right permissions to the ID-ed clients.
    - when you have a shared resource with varying levels of ownership (read/write permissions)

## The order of building / operations in this project:
### Chapter 1:
1. We defined the model of a Log and access methods. 
2. We defined an Http server, a method to create it, routes, and handler's names and signatures.
3. Request and response structs (since we're receiving requests and sending responses, that have to be marshalled/unmarshalled.)
4. Implement the handlers
5. main.go logic to run the server
### Chapter 2: Protocol Buffers
1. Define protos & make sure it compiles
learning opportunity: can write a protobuf extensions/plugins
### Chapter 3: Write a Log Package
1. Create an store for our log files (a wrapper around a 'file' in our case)
2. Code up the read and write methods to persist our records
3. Test file
4. Write out the index struct and logic, test file
5. Segment logic (so that we can split our log into segmentes when one gets too big), test file
6. Code the Log + test
### Chapter 4: Add gRPC service
1. Add grpc Log service, declare methods, response and request objects
2. Compile the code and see it generate log_grpc.pb.go
3. Implement a grpc server that will implement the Log Service and define its methods
4. Error handling
5. Swap out the concrete Log structure / object our server depends on to an interface
6. Create a gRPC server and register it (NewGRPCServer)
7. Tests!
### Chapter 5: Security
1. Create a cert issuer authority using CloudFlare's open source lib
2. Define the configs and write out the makefile cmds to generate certs
3. Add a /config dir to take care of retrieving the cert files and parsing them
4. Add grpc opts to our server so it can handle a creds opt to handle tls conns
5. 