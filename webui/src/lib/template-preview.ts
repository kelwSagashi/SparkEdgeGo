const templatePattern = /\{\{([^}]+)\}\}/g;

function getPathValue(context: Record<string, unknown>, path: string): unknown {
  const normalized = path.trim().replace(/^\$\./, "");
  if (!normalized) {
    return context;
  }

  return normalized.split(".").reduce<unknown>((acc, key) => {
    if (acc && typeof acc === "object" && !Array.isArray(acc)) {
      return (acc as Record<string, unknown>)[key];
    }
    if (Array.isArray(acc)) {
      const index = Number.parseInt(key, 10);
      return Number.isNaN(index) ? undefined : acc[index];
    }
    return undefined;
  }, context);
}

function splitArguments(value: string): string[] {
  const parts: string[] = [];
  let depth = 0;
  let quote: string | null = null;
  let start = 0;

  for (let index = 0; index < value.length; index += 1) {
    const char = value[index];
    if ((char === "'" || char === '"') && value[index - 1] !== "\\") {
      quote = quote === char ? null : quote ?? char;
    } else if (!quote && char === "(") {
      depth += 1;
    } else if (!quote && char === ")") {
      depth -= 1;
    } else if (!quote && depth === 0 && char === ",") {
      parts.push(value.slice(start, index).trim());
      start = index + 1;
    }
  }

  parts.push(value.slice(start).trim());
  return parts.filter(Boolean);
}

function parseLiteral(value: string): unknown {
  const trimmed = value.trim();
  if (
    (trimmed.startsWith("'") && trimmed.endsWith("'")) ||
    (trimmed.startsWith('"') && trimmed.endsWith('"'))
  ) {
    return trimmed.slice(1, -1);
  }
  if (trimmed === "true") return true;
  if (trimmed === "false") return false;
  if (trimmed === "null") return null;
  const numeric = Number(trimmed);
  if (!Number.isNaN(numeric) && trimmed !== "") {
    return numeric;
  }
  return undefined;
}

function evaluateFunction(
  name: string,
  args: string[],
  context: Record<string, unknown>,
): unknown {
  const values = args.map((arg) => evaluateExpression(arg, context));
  switch (name) {
    case "concat":
      return values.map((item) => `${item ?? ""}`).join("");
    case "upper":
      return `${values[0] ?? ""}`.toUpperCase();
    case "lower":
      return `${values[0] ?? ""}`.toLowerCase();
    case "trim":
      return `${values[0] ?? ""}`.trim();
    case "default":
    case "coalesce":
      return values.find((item) => item !== undefined && item !== null && item !== "");
    case "if":
      return values[0] ? values[1] : values[2];
    case "add":
      return values.reduce((sum, current) => sum + Number(current ?? 0), 0);
    case "string":
      return `${values[0] ?? ""}`;
    case "number":
      return Number(values[0] ?? 0);
    case "bool":
      return Boolean(values[0]);
    default:
      return undefined;
  }
}

export function evaluateExpression(
  expression: string,
  context: Record<string, unknown>,
): unknown {
  const trimmed = expression.trim();

  const literal = parseLiteral(trimmed);
  if (literal !== undefined || trimmed === "null") {
    return literal;
  }

  const ternaryIndex = trimmed.indexOf("?");
  if (ternaryIndex !== -1) {
    const colonIndex = trimmed.indexOf(":", ternaryIndex);
    if (colonIndex !== -1) {
      const condition = trimmed.slice(0, ternaryIndex).trim();
      const truthy = trimmed.slice(ternaryIndex + 1, colonIndex).trim();
      const falsy = trimmed.slice(colonIndex + 1).trim();
      return evaluateExpression(condition, context)
        ? evaluateExpression(truthy, context)
        : evaluateExpression(falsy, context);
    }
  }

  const fallbackParts = trimmed.split("||").map((part) => part.trim());
  if (fallbackParts.length > 1) {
    for (const part of fallbackParts) {
      const resolved = evaluateExpression(part, context);
      if (resolved !== undefined && resolved !== null && resolved !== "") {
        return resolved;
      }
    }
    return undefined;
  }

  const functionMatch = trimmed.match(/^([a-zA-Z_][a-zA-Z0-9_]*)\((.*)\)$/);
  if (functionMatch) {
    return evaluateFunction(functionMatch[1], splitArguments(functionMatch[2]), context);
  }

  return getPathValue(context, trimmed);
}

export function resolveTemplatePreview(
  input: unknown,
  context: Record<string, unknown>,
): unknown {
  if (typeof input === "string") {
    const trimmed = input.trim();
    if (trimmed.startsWith("{{") && trimmed.endsWith("}}")) {
      return evaluateExpression(trimmed.slice(2, -2), context);
    }

    return input.replace(templatePattern, (_match, expression) => {
      const value = evaluateExpression(expression, context);
      return value === undefined || value === null ? "" : `${value}`;
    });
  }

  if (Array.isArray(input)) {
    return input.map((item) => resolveTemplatePreview(item, context));
  }

  if (input && typeof input === "object") {
    return Object.fromEntries(
      Object.entries(input as Record<string, unknown>).map(([key, value]) => [
        key,
        resolveTemplatePreview(value, context),
      ]),
    );
  }

  return input;
}

export function flattenPreviewPaths(
  input: unknown,
  prefix = "$",
): Array<{ label: string; path: string }> {
  if (Array.isArray(input)) {
    return input.flatMap((item, index) =>
      flattenPreviewPaths(item, `${prefix}.${index}`),
    );
  }
  if (input && typeof input === "object") {
    return Object.entries(input as Record<string, unknown>).flatMap(([key, value]) => {
      const nextPath = `${prefix}.${key}`;
      const nested = flattenPreviewPaths(value, nextPath);
      return nested.length > 0 ? nested : [{ label: nextPath, path: `{{${nextPath}}}` }];
    });
  }
  return prefix === "$" ? [] : [{ label: prefix, path: `{{${prefix}}}` }];
}
