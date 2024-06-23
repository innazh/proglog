package discovery

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/serf/serf"
	"github.com/travisjeffery/go-dynaport"

	"github.com/stretchr/testify/require"
)

// TestMembership setsup a cluster with multiple servers and checks that the Membership returns all the servers that join and updates after a leave.
func TestMembership(t *testing.T) {
	memberArray, handler := setupMember(t, nil)
	memberArray, _ = setupMember(t, memberArray)
	memberArray, _ = setupMember(t, memberArray)

	require.Eventually(t, func() bool {
		return 2 == len(handler.joins) &&
			3 == len(memberArray[0].Members()) &&
			0 == len(handler.leaves)
	}, 3*time.Second, 250*time.Millisecond)

	require.NoError(t, memberArray[2].Leave())

	require.Eventually(t, func() bool {
		return 2 == len(handler.joins) &&
			3 == len(memberArray[0].Members()) &&
			serf.StatusLeft == memberArray[0].Members()[2].Status &&
			1 == len(handler.leaves)
	}, 3*time.Second, 250*time.Millisecond)

	require.Equal(t, fmt.Sprintf("%d", 2), <-handler.leaves)
}

type handler struct {
	joins  chan map[string]string
	leaves chan string
}

func (h *handler) Join(id, addr string) error {
	if h.joins != nil {
		h.joins <- map[string]string{
			"id":   id,
			"addr": addr,
		}
	}
	return nil
}

func (h *handler) Leave(id string) error {
	if h.leaves != nil {
		h.leaves <- id
	}
	return nil
}

func setupMember(t *testing.T, members []*Membership) ([]*Membership, *handler) {
	id := len(members)
	ports := dynaport.Get(1)
	addr := fmt.Sprintf("%s:%d", "127.0.0.1", ports[0])
	tags := map[string]string{
		"rpc_addr": addr,
	}
	c := Config{
		NodeName: fmt.Sprintf("%d", id),
		BindAddr: addr,
		Tags:     tags,
	}
	h := &handler{}
	if len(members) == 0 { //if this is a newly born cluster, then init the channels
		h.joins = make(chan map[string]string, 3)
		h.leaves = make(chan string, 3)
	} else {
		c.StartJoinAddrs = []string{
			members[0].BindAddr,
		}
	}
	m, err := NewMembership(h, c)
	require.NoError(t, err)
	members = append(members, m)
	return members, h
}
