/**
 * Infere um esquema JSON a partir de um objeto de dados
 */
export function inferSchema(data: any): any {
  if (data === null || data === undefined) {
    return { type: "null" };
  }

  const type = typeof data;

  if (type === "string") {
    return { type: "string" };
  }

  if (type === "number") {
    return { type: "number" };
  }

  if (type === "boolean") {
    return { type: "boolean" };
  }

  if (Array.isArray(data)) {
    const items = data.length > 0 ? inferSchema(data[0]) : { type: "string" };
    return {
      type: "array",
      items,
    };
  }

  if (type === "object") {
    const properties: Record<string, any> = {};
    const required: string[] = [];

    Object.entries(data).forEach(([key, value]) => {
      properties[key] = inferSchema(value);
      required.push(key);
    });

    return {
      type: "object",
      properties,
      required,
    };
  }

  return { type: "string" };
}

