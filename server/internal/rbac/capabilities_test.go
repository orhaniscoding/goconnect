package rbac

import (
    "testing"
    "github.com/orhaniscoding/goconnect/server/internal/domain"
)

func TestCapabilitiesMatrix(t *testing.T) {
    roles := []domain.MembershipRole{domain.RoleMember, domain.RoleAdmin, domain.RoleOwner}
    for _, r := range roles {
        // Manage network
        manage := CanManageNetwork(r)
        approve := CanApproveMembership(r)
        ban := CanBanMember(r)
        relOther := CanReleaseOtherIP(r)
        switch r {
        case domain.RoleMember:
            if manage || approve || ban || relOther { t.Fatalf("member should have no elevated perms") }
        case domain.RoleAdmin, domain.RoleOwner:
            if !manage || !approve || !ban || !relOther { t.Fatalf("%s should have all elevated perms", r) }
        }
    }
    // Global admin flag scenario for view all
    if !CanViewAllNetworks(true, domain.RoleMember) { t.Fatalf("global admin flag should allow view all") }
    if CanViewAllNetworks(false, domain.RoleMember) { t.Fatalf("member without admin flag should not view all") }
}
