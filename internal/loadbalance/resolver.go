package loadbalance

import (
	"context"
	"fmt"
	"sync"

	api "github.com/innazh/proglog/api/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
)

//gRPC uses the builder pattern for resolvers and pickers.

const Name = "proglog"

// Resolver
type Resolver struct {
	mu           sync.Mutex
	clientConn   resolver.ClientConn //user's client conn, so that the Resolver can upd it with servers it discovers
	resolverConn *grpc.ClientConn    //resolver's conn to the server used to call GetServers()
	serviceConf  *serviceconfig.ParseResult
	logger       *zap.Logger
}

// implementing the Builder interface below
var _ resolver.Builder = (*Resolver)(nil)

// Build receives the data needed to build a resolver that can discover the servers, connects to the server (for the GetServers API call later on)
func (r *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.logger = zap.L().Named("resolver")
	r.clientConn = cc

	var dialOpts []grpc.DialOption
	if opts.DialCreds != nil {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(opts.DialCreds))
	}
	r.serviceConf = r.clientConn.ParseServiceConfig(fmt.Sprintf(`{"loadBalancingConfig":[{"%s":{}}]}`, Name)) // clients need to place "prglog" at the start of target address

	var err error
	fmt.Println(target.Endpoint())
	r.resolverConn, err = grpc.NewClient(target.Endpoint(), dialOpts...) //parses out the scheme from the target address and tries to find a resolver that matches, defaulting to DNS resolver
	if err != nil {
		return nil, err
	}
	r.ResolveNow(resolver.ResolveNowOptions{})

	return r, nil
}

// Scheme returns the resolver's scheme identifier
func (r *Resolver) Scheme() string {
	return Name
}

// init registers the resolver builder to the resolver map, so grpc knows about this resolver when its looking for resolvers that match the target's scheme
func init() {
	resolver.Register(&Resolver{})
}

// implementing the Resolver interface below
var _ resolver.Resolver = (*Resolver)(nil)

// ResolveNow resolves the target, discovers the servers, and updates the client conn with servers
func (r *Resolver) ResolveNow(resolver.ResolveNowOptions) {
	//gRPC may call this method concurrently, hence:
	r.mu.Lock()
	defer r.mu.Unlock()

	//The discovery part below depends on the service we're working with; For example, a resolver built for Kubernetes, could call Kubernetes' API to get the list of endpoints
	client := api.NewLogClient(r.resolverConn)
	//get cluster and then set cc on attributes
	ctx := context.Background()
	res, err := client.GetServers(ctx, &api.GetServersRequest{})
	if err != nil {
		r.logger.Error("failed to resolve server", zap.Error(err))
		return
	}

	var addrs []resolver.Address
	for _, server := range res.Servers {
		addrs = append(addrs, resolver.Address{
			Addr: server.RpcAddr,
			Attributes: attributes.New(
				"is_leader",
				server.IsLeader,
			),
		})
	}

	err = r.clientConn.UpdateState(resolver.State{
		Addresses:     addrs,         //inform the load balancer what servers it can choose from
		ServiceConfig: r.serviceConf, // specifies how clients should balance their calls to the service
	})
	if err != nil {
		r.logger.Error("failed update client's connection", zap.Error(err))
		return
	}
}

// Close closes resolver's connection to the server that we setup in Build()
func (r *Resolver) Close() {
	if err := r.resolverConn.Close(); err != nil {
		r.logger.Error("failed to close conn", zap.Error(err))
	}
}
