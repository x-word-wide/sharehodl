package types

import (
	"encoding/binary"
)

// Key prefixes for agent module storage
var (
	AgentPrefix       = []byte{0x01}
	AgentByKeyPrefix  = []byte{0x02}
	AgentByOwnerPrefix = []byte{0x03}
	ActionPrefix      = []byte{0x10}
	ActionByAgentPrefix = []byte{0x11}
	SubscriptionPrefix = []byte{0x20}
	AgentCounterKey   = []byte{0x90}
	ActionCounterKey  = []byte{0x91}
)

// GetAgentKey returns the store key for an agent by ID
func GetAgentKey(agentID uint64) []byte {
	key := make([]byte, len(AgentPrefix)+8)
	copy(key, AgentPrefix)
	binary.BigEndian.PutUint64(key[len(AgentPrefix):], agentID)
	return key
}

// GetAgentByAPIKeyKey returns the index key for looking up agents by API key
func GetAgentByAPIKeyKey(apiKey string) []byte {
	return append(AgentByKeyPrefix, []byte(apiKey)...)
}

// GetAgentByOwnerKey returns the index key for looking up agents by owner
func GetAgentByOwnerKey(owner string, agentID uint64) []byte {
	ownerBytes := []byte(owner)
	key := make([]byte, len(AgentByOwnerPrefix)+len(ownerBytes)+1+8)
	copy(key, AgentByOwnerPrefix)
	copy(key[len(AgentByOwnerPrefix):], ownerBytes)
	key[len(AgentByOwnerPrefix)+len(ownerBytes)] = 0x00 // separator
	binary.BigEndian.PutUint64(key[len(AgentByOwnerPrefix)+len(ownerBytes)+1:], agentID)
	return key
}

// GetActionKey returns the store key for an action by ID
func GetActionKey(actionID uint64) []byte {
	key := make([]byte, len(ActionPrefix)+8)
	copy(key, ActionPrefix)
	binary.BigEndian.PutUint64(key[len(ActionPrefix):], actionID)
	return key
}

// GetActionByAgentKey returns the index key for looking up actions by agent
func GetActionByAgentKey(agentID, actionID uint64) []byte {
	key := make([]byte, len(ActionByAgentPrefix)+16)
	copy(key, ActionByAgentPrefix)
	binary.BigEndian.PutUint64(key[len(ActionByAgentPrefix):], agentID)
	binary.BigEndian.PutUint64(key[len(ActionByAgentPrefix)+8:], actionID)
	return key
}

// GetSubscriptionKey returns the store key for a subscription
func GetSubscriptionKey(owner string) []byte {
	return append(SubscriptionPrefix, []byte(owner)...)
}

// OwnerIteratorKey returns the prefix for iterating over agents by owner
func OwnerIteratorKey(owner string) []byte {
	ownerBytes := []byte(owner)
	key := make([]byte, len(AgentByOwnerPrefix)+len(ownerBytes)+1)
	copy(key, AgentByOwnerPrefix)
	copy(key[len(AgentByOwnerPrefix):], ownerBytes)
	key[len(AgentByOwnerPrefix)+len(ownerBytes)] = 0x00 // separator
	return key
}

// AgentActionIteratorKey returns the prefix for iterating over actions by agent
func AgentActionIteratorKey(agentID uint64) []byte {
	key := make([]byte, len(ActionByAgentPrefix)+8)
	copy(key, ActionByAgentPrefix)
	binary.BigEndian.PutUint64(key[len(ActionByAgentPrefix):], agentID)
	return key
}

// PrefixEndBytes returns the end bytes for a prefix iterator
func PrefixEndBytes(prefix []byte) []byte {
	end := make([]byte, len(prefix))
	copy(end, prefix)
	for i := len(end) - 1; i >= 0; i-- {
		if end[i] < 0xff {
			end[i]++
			return end[:i+1]
		}
	}
	return nil // All 0xff bytes, no end
}
