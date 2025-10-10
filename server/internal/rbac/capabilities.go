package rbac

import "github.com/orhaniscoding/goconnect/server/internal/domain"

// Capability helpers centralize authorization role logic.
// Keep functions small & pure to simplify unit testing.

// CanManageNetwork returns true if role can update/delete network configuration.
func CanManageNetwork(role domain.MembershipRole) bool {
	return role == domain.RoleAdmin || role == domain.RoleOwner
}

// CanApproveMembership returns true if role can approve join requests.
func CanApproveMembership(role domain.MembershipRole) bool {
	return role == domain.RoleAdmin || role == domain.RoleOwner
}

// CanBanMember returns true if role can ban other members.
func CanBanMember(role domain.MembershipRole) bool { return CanApproveMembership(role) }

// CanReleaseOtherIP returns true if role can release another member's IP allocation.
func CanReleaseOtherIP(role domain.MembershipRole) bool { return CanManageNetwork(role) }

// CanViewAllNetworks indicates ability to list with visibility=all.
func CanViewAllNetworks(isAdminFlag bool, role domain.MembershipRole) bool {
	if isAdminFlag { // global admin bypass (from token)
		return true
	}
	return role == domain.RoleAdmin || role == domain.RoleOwner
}
