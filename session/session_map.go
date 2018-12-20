package session

import (
	"fmt"

	"sync"
)

// Map represents a Session Map (Account[Region]=Session)
type Map struct {
	m sync.Map
}

// Ensure that we have the account map initialized
func (m *Map) Ensure(account string) {
	if _, ok := m.m.Load(account); !ok {
		m.m.Store(account, &sync.Map{})
	}
}

// Put a session for a given account and region
func (m *Map) Put(account string, region string, session *AWSSession) {
	a, _ := m.m.Load(account)
	am, _ := a.(*sync.Map)
	am.Store(region, session)

}

func (m *Map) get(account string, region string) (*AWSSession, error) {
	if r, ok := m.m.Load(account); ok {
		if rs, ok := r.(*sync.Map); ok {
			if s, ok := rs.Load(region); ok {
				sess := s.(*AWSSession)
				return sess, nil
			}
		}
	}
	return nil, fmt.Errorf("session not found for account '%v' on region '%v'", account, region)
}

// ForEach allows to iterate over all the sessions
func (m *Map) forEach(fn HandlerFunc) {
	m.m.Range(func(account, regionMap interface{}) bool {
		rm := regionMap.(*sync.Map)
		rm.Range(func(region, sess interface{}) bool {
			s := sess.(*AWSSession)
			err := fn(account.(string), region.(string), s)
			return err == nil
		})
		return true
	})
}
