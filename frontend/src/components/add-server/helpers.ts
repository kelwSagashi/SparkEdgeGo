import { z } from "zod";
import type { ServerHeaderSchema } from "./schemas";

export function parseHeaderValue(
  value: string,
  type: z.infer<typeof ServerHeaderSchema.shape.type>
) {
  switch (type) {
    case "text":
      return String(value);

    case "number": {
      const num = Number(value);
      if (Number.isNaN(num)) {
        throw new Error(`Valor inválido para number: ${value}`);
      }
      return num;
    }
    default:
      return String(value);
  }
}
export function headersArrayToRecord(
  headers?: z.infer<typeof ServerHeaderSchema>[]
) {
  if (!headers || headers.length === 0) return {};

  return headers.reduce<Record<string, any>>((acc, cur) => {
    acc[cur.key] = parseHeaderValue(cur.value, cur.type);
    return acc;
  }, {});
}

export function recordToHeadersArray(headersObj?: Record<string, any>) {
  if (!headersObj) return [];
  return Object.entries(headersObj).map(([key, value]) => ({
    key,
    value: String(value),
    type: (typeof value === "number" ? "number" : "text") as "text" | "number",
  }));
}

