package repository

import (
	"context"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemoryIPAM(t *testing.T) {
	repo := NewInMemoryIPAM()

	assert.NotNil(t, repo)
}

func TestIPAMRepository_GetOrAllocate_FirstAllocation(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	alloc, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")

	require.NoError(t, err)
	require.NotNil(t, alloc)
	assert.Equal(t, "network-1", alloc.NetworkID)
	// subjectID is now stored in DeviceID field (device-based allocation)
	assert.Equal(t, "user-1", alloc.DeviceID)
	assert.NotEmpty(t, alloc.IP)
	assert.Equal(t, "10.0.0.1", alloc.IP) // First usable IP
	assert.Equal(t, uint32(1), alloc.Offset)
}

func TestIPAMRepository_GetOrAllocate_MultipleUsers(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	// Allocate for three users
	alloc1, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)

	alloc2, err := repo.GetOrAllocate(ctx, "network-1", "user-2", "10.0.0.0/24")
	require.NoError(t, err)

	alloc3, err := repo.GetOrAllocate(ctx, "network-1", "user-3", "10.0.0.0/24")
	require.NoError(t, err)

	// All should have different IPs
	assert.NotEqual(t, alloc1.IP, alloc2.IP)
	assert.NotEqual(t, alloc2.IP, alloc3.IP)
	assert.NotEqual(t, alloc1.IP, alloc3.IP)

	// Sequential IPs
	assert.Equal(t, "10.0.0.1", alloc1.IP)
	assert.Equal(t, "10.0.0.2", alloc2.IP)
	assert.Equal(t, "10.0.0.3", alloc3.IP)
}

func TestIPAMRepository_GetOrAllocate_Idempotent(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	// First allocation
	alloc1, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)

	// Second call should return same allocation
	alloc2, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)

	assert.Equal(t, alloc1.IP, alloc2.IP)
	assert.Equal(t, alloc1.Offset, alloc2.Offset)
}

func TestIPAMRepository_GetOrAllocate_DifferentNetworks(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	// Same user, different networks
	alloc1, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)

	alloc2, err := repo.GetOrAllocate(ctx, "network-2", "user-1", "10.0.1.0/24")
	require.NoError(t, err)

	assert.Equal(t, "network-1", alloc1.NetworkID)
	assert.Equal(t, "network-2", alloc2.NetworkID)
	assert.Equal(t, "10.0.0.1", alloc1.IP)
	assert.Equal(t, "10.0.1.1", alloc2.IP)
}

func TestIPAMRepository_List_EmptyNetwork(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	list, err := repo.List(ctx, "network-1")

	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestIPAMRepository_List_WithAllocations(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	// Create allocations
	repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
	repo.GetOrAllocate(ctx, "network-1", "user-2", "10.0.0.0/24")
	repo.GetOrAllocate(ctx, "network-1", "user-3", "10.0.0.0/24")

	list, err := repo.List(ctx, "network-1")

	require.NoError(t, err)
	assert.Len(t, list, 3)
}

func TestIPAMRepository_List_FilterByNetwork(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	// Allocations in different networks
	repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
	repo.GetOrAllocate(ctx, "network-1", "user-2", "10.0.0.0/24")
	repo.GetOrAllocate(ctx, "network-2", "user-3", "10.0.1.0/24")

	list1, err := repo.List(ctx, "network-1")
	require.NoError(t, err)
	assert.Len(t, list1, 2)

	list2, err := repo.List(ctx, "network-2")
	require.NoError(t, err)
	assert.Len(t, list2, 1)
}

func TestIPAMRepository_Release_Success(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	// Allocate
	alloc, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1", alloc.IP)

	// Release
	err = repo.Release(ctx, "network-1", "user-1")
	require.NoError(t, err)

	// List should be empty
	list, _ := repo.List(ctx, "network-1")
	assert.Empty(t, list)
}

func TestIPAMRepository_Release_Idempotent(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	// Allocate
	repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")

	// Release multiple times
	err1 := repo.Release(ctx, "network-1", "user-1")
	require.NoError(t, err1)

	err2 := repo.Release(ctx, "network-1", "user-1")
	require.NoError(t, err2)

	err3 := repo.Release(ctx, "network-1", "user-1")
	require.NoError(t, err3)
}

func TestIPAMRepository_Release_NonExistentNetwork(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	// Release from non-existent network (should be idempotent)
	err := repo.Release(ctx, "non-existent", "user-1")

	require.NoError(t, err)
}

func TestIPAMRepository_Release_NonExistentUser(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	// Create network with one user
	repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")

	// Release non-existent user (should be idempotent)
	err := repo.Release(ctx, "network-1", "non-existent")

	require.NoError(t, err)

	// Original allocation should still exist
	list, _ := repo.List(ctx, "network-1")
	assert.Len(t, list, 1)
}

func TestIPAMRepository_ReleaseAndReuse(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	// Allocate for user-1
	alloc1, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	ip1 := alloc1.IP

	// Allocate for user-2
	alloc2, err := repo.GetOrAllocate(ctx, "network-1", "user-2", "10.0.0.0/24")
	require.NoError(t, err)
	ip2 := alloc2.IP

	// Release user-1
	repo.Release(ctx, "network-1", "user-1")

	// Allocate for user-3 - should reuse user-1's IP (LIFO)
	alloc3, err := repo.GetOrAllocate(ctx, "network-1", "user-3", "10.0.0.0/24")
	require.NoError(t, err)

	assert.Equal(t, ip1, alloc3.IP) // Reused IP
	assert.NotEqual(t, ip2, alloc3.IP)
}

func TestIPAMRepository_MultipleReleaseAndReuse(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	// Allocate 5 IPs
	users := []string{"user-a", "user-b", "user-c", "user-d", "user-e"}
	for _, user := range users {
		repo.GetOrAllocate(ctx, "network-1", user, "10.0.0.0/24")
	}

	list, _ := repo.List(ctx, "network-1")
	assert.Len(t, list, 5)

	// Release user-b, user-c, user-d
	repo.Release(ctx, "network-1", "user-b")
	repo.Release(ctx, "network-1", "user-c")
	repo.Release(ctx, "network-1", "user-d")

	list, _ = repo.List(ctx, "network-1")
	assert.Len(t, list, 2) // Only user-a and user-e remain

	// Allocate new users - should reuse in LIFO order (d, c, b)
	alloc1, _ := repo.GetOrAllocate(ctx, "network-1", "new-1", "10.0.0.0/24")
	alloc2, _ := repo.GetOrAllocate(ctx, "network-1", "new-2", "10.0.0.0/24")
	alloc3, _ := repo.GetOrAllocate(ctx, "network-1", "new-3", "10.0.0.0/24")

	// Should reuse the released IPs
	ips := []string{alloc1.IP, alloc2.IP, alloc3.IP}
	assert.Contains(t, ips, "10.0.0.2") // user-b's IP
	assert.Contains(t, ips, "10.0.0.3") // user-c's IP
	assert.Contains(t, ips, "10.0.0.4") // user-d's IP
}

func TestIPAMRepository_SmallCIDR(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	// /30 network has only 2 usable IPs
	alloc1, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/30")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1", alloc1.IP)

	alloc2, err := repo.GetOrAllocate(ctx, "network-1", "user-2", "10.0.0.0/30")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.2", alloc2.IP)
}

func TestIPAMRepository_IPExhaustion(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	// /30 network has only 2 usable IPs
	repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/30")
	repo.GetOrAllocate(ctx, "network-1", "user-2", "10.0.0.0/30")

	// Third allocation should fail
	_, err := repo.GetOrAllocate(ctx, "network-1", "user-3", "10.0.0.0/30")

	require.Error(t, err)
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrIPExhausted, domainErr.Code)
	assert.Contains(t, domainErr.Message, "No available IP")
}

func TestIPAMRepository_LargeCIDR(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	// /16 network has many IPs
	for i := 1; i <= 10; i++ {
		alloc, err := repo.GetOrAllocate(ctx, "network-1", "user-"+string(rune('a'+i)), "10.0.0.0/16")
		require.NoError(t, err)
		assert.NotEmpty(t, alloc.IP)
	}

	list, _ := repo.List(ctx, "network-1")
	assert.Len(t, list, 10)
}

func TestIPAMRepository_ConcurrentAllocations(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	// Pre-allocate some IPs
	repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
	repo.GetOrAllocate(ctx, "network-1", "user-2", "10.0.0.0/24")

	// Concurrent reads (List) should be safe
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			list, err := repo.List(ctx, "network-1")
			assert.NoError(t, err)
			assert.NotNil(t, list)
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestIPAMRepository_FullCycle(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	// Allocate
	alloc, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1", alloc.IP)

	// List
	list, err := repo.List(ctx, "network-1")
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// Idempotent get
	alloc2, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, alloc.IP, alloc2.IP)

	// Release
	err = repo.Release(ctx, "network-1", "user-1")
	require.NoError(t, err)

	// List should be empty
	list, _ = repo.List(ctx, "network-1")
	assert.Empty(t, list)

	// Reallocate - should get same IP (reuse)
	alloc3, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1", alloc3.IP)
}

func TestIPAMRepository_OffsetTracking(t *testing.T) {
	repo := NewInMemoryIPAM()
	ctx := context.Background()

	alloc1, _ := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
	alloc2, _ := repo.GetOrAllocate(ctx, "network-1", "user-2", "10.0.0.0/24")
	alloc3, _ := repo.GetOrAllocate(ctx, "network-1", "user-3", "10.0.0.0/24")

	assert.Equal(t, uint32(1), alloc1.Offset)
	assert.Equal(t, uint32(2), alloc2.Offset)
	assert.Equal(t, uint32(3), alloc3.Offset)
}
