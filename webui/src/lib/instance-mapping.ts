import type { DeviceReturningValues } from "@/types/db";

export function normalizeDeviceFields(device?: DeviceReturningValues | null) {
  if (!device) {
    return null;
  }

  const othersObject = normalizeDeviceOthers(device.others);
  return {
    id: device.id,
    device_id: device.device_id ?? device.id,
    name: device.name,
    brand: device.brand,
    serial_number: device.serial_number,
    connection_method: device.connection_method ?? device.connection,
    ip_address: device.ip_address,
    location: device.location,
    description: device.description,
    ...othersObject,
    others: othersObject,
  };
}

export function normalizeDeviceOthers(
  others: DeviceReturningValues["others"],
): Record<string, unknown> {
  if (!Array.isArray(others)) {
    return {};
  }

  return others.reduce<Record<string, unknown>>((acc, item) => {
    const key = typeof item?.key === "string" ? item.key : "";
    if (!key) {
      return acc;
    }
    acc[key] = item.value;
    return acc;
  }, {});
}

export function buildInstanceRuntimePreview(args: {
  instanceName?: string;
  instanceId?: string;
  deviceId?: string | null;
  selectedDevice?: DeviceReturningValues | null;
  includeDeviceData?: boolean;
  scriptOutput?: Record<string, unknown>;
  destinations?: Array<{ serverId?: string | null; resourceOperationId?: string }>;
}) {
  const device = normalizeDeviceFields(args.selectedDevice);

  return {
    device,
    device_data: device,
    instance: {
      id: args.instanceId ?? "",
      name: args.instanceName || "Nova Instancia",
      device_id: args.deviceId ?? "",
    },
    system: {
      now: new Date().toISOString(),
      instance_name: args.instanceName || "Nova Instancia",
      instance_id: args.instanceId ?? "",
      device_id: args.deviceId ?? "",
      include_device_data: !!args.includeDeviceData,
      destinations: (args.destinations || []).map((d, index) => ({
        index: index + 1,
        server_id: d.serverId || "",
        operation_id: d.resourceOperationId || "",
      })),
    },
    script: args.scriptOutput || {},
    output: args.scriptOutput || {},
  };
}
