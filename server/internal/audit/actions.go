package audit

// Action constants centralize audit action names to avoid typos.
// NOTE: Do not log PII in details; actor/object are redacted downstream.
const (
    ActionNetworkCreated     = "NETWORK_CREATED"
    ActionNetworkUpdated     = "NETWORK_UPDATED"
    ActionNetworkDeleted     = "NETWORK_DELETED"
    ActionNetworkJoinApprove = "NETWORK_JOIN_APPROVE"
    ActionNetworkMemberBan   = "NETWORK_MEMBER_BAN"
    ActionIPAllocated        = "IP_ALLOCATED"
    ActionIPReleased         = "IP_RELEASED"
)
