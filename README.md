# Proglog
Commit logs - append-only data structure, sequenced by time.

Each handler consists of three steps:
1. Unmarshal the request's JSON body into a struct.
2. Complete that endpoint's logic with the request, obtain a result.
3. Marshal and write the result to the response.
If a handler becomes much more complicated than this, then move req & resp handling to middleware, business logic further down the stack.

The order of building / operations in this project:
1. We defined the model of a Log and access methods. 
2. We defined an Http server, a method to create it, routes, and handler's names and signatures.
3. Request and response structs (since we're receiving requests and sending responses, that have to be marshalled/unmarshalled.)
4. Implement the handlers
5. main.go logic to run the server