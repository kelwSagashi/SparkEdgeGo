package devices

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

func TestDevicesServiceLifecycle(t *testing.T) {
	ctx := context.Background()
	store := sqlite.NewStore()
	store.Path = filepath.Join(t.TempDir(), "sparkedge-test.db")
	if err := store.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	service := NewService(store.Devices)
	device, err := service.Upsert(ctx, UpsertRequest{
		DeviceID:         "edge-device-1",
		Name:             "Pump Sensor",
		Brand:            "Acme",
		ConnectionMethod: domain.DeviceConnectionMQTT,
		IPAddress:        "192.168.1.10",
		Others: []domain.DeviceOtherField{
			{Key: "floor", Value: "2", Type: "number"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if device.ID == "" || device.DeviceID != "edge-device-1" {
		t.Fatalf("unexpected device identifiers: %#v", device)
	}

	found, err := service.FindByDeviceID(ctx, "edge-device-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(found.Others) != 1 || found.Others[0].Key != "floor" {
		t.Fatalf("expected JSON others to round-trip, got %#v", found.Others)
	}

	updated, err := service.Upsert(ctx, UpsertRequest{
		ID:               device.ID,
		DeviceID:         device.DeviceID,
		Name:             "Pump Sensor Updated",
		Brand:            "Acme",
		ConnectionMethod: domain.DeviceConnectionHTTP,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Pump Sensor Updated" || updated.ConnectionMethod != domain.DeviceConnectionHTTP {
		t.Fatalf("unexpected updated device %#v", updated)
	}

	items, err := service.ListAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one device, got %d", len(items))
	}

	if err := service.Delete(ctx, device.ID); err != nil {
		t.Fatal(err)
	}
	items, err = service.ListAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no devices after delete, got %d", len(items))
	}
}
