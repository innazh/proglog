package auth

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Authorizer is a wrapper class for casbin's enforcer that helps us to restrict access to the api
type Authorizer struct {
	enforcer *casbin.Enforcer
}

// model - our authorization model config, in this case - ACL
// policy - the fiile containing the ACL table
func New(model, policy string) (*Authorizer, error) {
	enforcer, err := casbin.NewEnforcer(model, policy)
	if err != nil {
		return nil, err
	}
	return &Authorizer{enforcer: enforcer}, nil
}

func (a *Authorizer) Authorize(subject, object, action string) error {
	rule, err := a.enforcer.Enforce(subject, object, action)
	if err != nil {
		return err
	}
	if !rule {
		msg := fmt.Sprintf("%s not permitted to %s", subject, action)
		st := status.New(codes.PermissionDenied, msg)
		return st.Err()
	}
	return nil
}
