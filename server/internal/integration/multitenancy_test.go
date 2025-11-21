package integration

import (
	"context"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/config"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

func TestMultiTenancyIsolation(t *testing.T) {
	// Setup
	networkRepo := repository.NewInMemoryNetworkRepository()
	deviceRepo := repository.NewInMemoryDeviceRepository()
	userRepo := repository.NewInMemoryUserRepository()
	idempotencyRepo := repository.NewInMemoryIdempotencyRepository()
	membershipRepo := repository.NewInMemoryMembershipRepository()
	joinRepo := repository.NewInMemoryJoinRequestRepository()
	ipamRepo := repository.NewInMemoryIPAM()

	networkService := service.NewNetworkService(networkRepo, idempotencyRepo)
	peerRepo := repository.NewInMemoryPeerRepository()
	wgConfig := config.WireGuardConfig{}
	deviceService := service.NewDeviceService(deviceRepo, userRepo, peerRepo, networkRepo, wgConfig)
	membershipService := service.NewMembershipService(networkRepo, membershipRepo, joinRepo, idempotencyRepo)
	ipamService := service.NewIPAMService(networkRepo, membershipRepo, ipamRepo)

	ctx := context.Background()

	// Tenant A Data
	tenantA := "tenant-a"
	userA := &domain.User{ID: "user-a", TenantID: tenantA, Email: "a@example.com"}
	_ = userRepo.Create(ctx, userA)

	// Tenant B Data
	tenantB := "tenant-b"
	userB := &domain.User{ID: "user-b", TenantID: tenantB, Email: "b@example.com"}
	_ = userRepo.Create(ctx, userB)

	// 1. Network Isolation
	t.Run("Network Isolation", func(t *testing.T) {
		// User A creates a PUBLIC network
		netA, err := networkService.CreateNetwork(ctx, &domain.CreateNetworkRequest{
			Name:       "Network A",
			CIDR:       "10.0.0.0/24",
			Visibility: "public",
		}, userA.ID, tenantA, "")
		if err != nil {
			t.Fatalf("Failed to create network for Tenant A: %v", err)
		}

		// User B tries to get Network A
		_, err = networkService.GetNetwork(ctx, netA.ID, userB.ID, tenantB)
		if err == nil {
			t.Errorf("Tenant B should not be able to get Tenant A's network")
		}

		// User B tries to list public networks (should not see Network A)
		nets, _, err := networkService.ListNetworks(ctx, &domain.ListNetworksRequest{Visibility: "public"}, userB.ID, tenantB, false)
		if err != nil {
			t.Fatalf("Failed to list networks for Tenant B: %v", err)
		}
		for _, n := range nets {
			if n.ID == netA.ID {
				t.Errorf("Tenant B should not see Tenant A's network in list")
			}
		}
	})

	// 2. Device Isolation
	t.Run("Device Isolation", func(t *testing.T) {
		// User A registers a device
		devA, err := deviceService.RegisterDevice(ctx, userA.ID, tenantA, &domain.RegisterDeviceRequest{
			Name:     "Device A",
			PubKey:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 44 chars
			Platform: "linux",
		})
		if err != nil {
			t.Fatalf("Failed to register device for Tenant A: %v", err)
		}

		// User B tries to get Device A
		_, err = deviceService.GetDevice(ctx, devA.ID, userB.ID, tenantB, false)
		if err == nil {
			t.Errorf("Tenant B should not be able to get Tenant A's device")
		}

		// User B tries to list devices (should not see Device A)
		devs, _, err := deviceService.ListDevices(ctx, userB.ID, tenantB, false, domain.DeviceFilter{})
		if err != nil {
			t.Fatalf("Failed to list devices for Tenant B: %v", err)
		}
		for _, d := range devs {
			if d.ID == devA.ID {
				t.Errorf("Tenant B should not see Tenant A's device in list")
			}
		}
	})

	// 3. IPAM Isolation
	t.Run("IPAM Isolation", func(t *testing.T) {
		// User A creates a network
		netA, err := networkService.CreateNetwork(ctx, &domain.CreateNetworkRequest{
			Name:       "Network A IPAM",
			CIDR:       "10.1.0.0/24",
			Visibility: "private",
			JoinPolicy: "open",
		}, userA.ID, tenantA, "")
		if err != nil {
			t.Fatalf("Failed to create network for Tenant A: %v", err)
		}

		// User A joins network
		_, _, err = membershipService.JoinNetwork(ctx, netA.ID, userA.ID, tenantA, "idem-join-a")
		if err != nil {
			t.Fatalf("Failed to join network for User A: %v", err)
		}

		// User A allocates IP
		allocA, err := ipamService.AllocateIP(ctx, netA.ID, userA.ID, tenantA)
		if err != nil {
			t.Fatalf("Failed to allocate IP for User A: %v", err)
		}
		if allocA.IP == "" {
			t.Errorf("Expected IP allocation for User A")
		}

		// User B tries to allocate IP in Tenant A's network
		_, err = ipamService.AllocateIP(ctx, netA.ID, userB.ID, tenantB)
		if err == nil {
			t.Errorf("Tenant B should not be able to allocate IP in Tenant A's network")
		}

		// User B tries to list allocations
		_, err = ipamService.ListAllocations(ctx, netA.ID, userB.ID, tenantB)
		if err == nil {
			t.Errorf("Tenant B should not be able to list allocations in Tenant A's network")
		}
	})
}
