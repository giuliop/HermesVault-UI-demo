package memstore

import (
	"fmt"
	"sync"
	"time"

	"github.com/giuliop/HermesVault-frontend/models"

	"github.com/algorand/go-algorand-sdk/v2/types"
)

type depositData struct {
	depositData *models.DepositData
	createdAt   time.Time
}

// MemoryStoreWithCleanup encapsulates a sync.Map and cleanup configuration
type MemoryStoreWithCleanup struct {
	data            sync.Map
	ttl             time.Duration
	cleanupInterval time.Duration
}

// Singleton instance and initialization
var UserSessions *MemoryStoreWithCleanup

func init() {
	UserSessions = &MemoryStoreWithCleanup{
		ttl:             10 * time.Minute, // how long to keep data in memory
		cleanupInterval: 5 * time.Minute,  // how often to run cleanup
	}
	go UserSessions.startCleanup()
}

// StoreDeposit adds a new TxnGroup to the store and returns its group ID
func (s *MemoryStoreWithCleanup) StoreDeposit(d *models.DepositData) (types.Digest, error) {
	groupId := d.Txns[0].Group
	if groupId == (types.Digest{}) {
		return types.Digest{}, fmt.Errorf("missing group ID")
	}
	s.data.Store(groupId, depositData{
		depositData: d,
		createdAt:   time.Now(),
	})
	return groupId, nil
}

// RetrieveDeposit fetches the TxnGroup associated with the given group ID
func (s *MemoryStoreWithCleanup) RetrieveDeposit(groupId types.Digest) (*models.DepositData, error) {
	value, ok := s.data.Load(groupId)
	if !ok {
		return nil, fmt.Errorf("depositID not found: %v", groupId)
	}
	data, ok := value.(depositData)
	if !ok {
		return nil, fmt.Errorf("invalid data type")
	}
	return data.depositData, nil
}

// DeleteDeposit removes the TransactionGroup associated with the given group ID
func (s *MemoryStoreWithCleanup) DeleteDeposit(groupId types.Digest) {
	s.data.Delete(groupId)
}

// startCleanup periodically removes expired entries from the store.
func (s *MemoryStoreWithCleanup) startCleanup() {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()
	for {
		<-ticker.C
		now := time.Now()
		s.data.Range(func(key, value interface{}) bool {
			tg, ok := value.(depositData)
			if !ok {
				s.data.Delete(key)
				return true
			}
			if now.Sub(tg.createdAt) > s.ttl {
				s.data.Delete(key)
			}
			return true
		})
	}
}
