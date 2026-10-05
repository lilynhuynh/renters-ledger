import { z } from "zod";

const PolicySchema = z.object({
    id: z.string(),
    customer_id: z.string(),
    unit: z.string(),
    status: z.enum(["quoted", "pending", "active", "past_due", "cancelled", "lapsed", "reinstated"]),
    premium_cents: z.number().int(),
    effective_date: z.string().refine((date) => !isNaN(Date.parse(date)), {
        message: "Invalid date format",
    }),
    version: z.number().int(),
    created_at: z.string().refine((date) => !isNaN(Date.parse(date)), {
        message: "Invalid date format",
    }),
    updated_at: z.string().refine((date) => !isNaN(Date.parse(date)), {
        message: "Invalid date format",
    }),
});
export { PolicySchema };

type Policy = z.infer<typeof PolicySchema>; // the TS type comes from the schema
export type { Policy };

const API_URL = process.env.API_URL ?? "http://localhost:8080";

export async function getPolicies(): Promise<Policy[]> {
    const res = await fetch(`${API_URL}/policies`, { cache: "no-store" });
    if (!res.ok) {
        throw new Error(`GET /policies failed: ${res.status}`);
    }
    return z.array(PolicySchema).parse(await res.json());
}
// Like completable future, gets list of policies
// fetch - wait for policies to return
// cache no store - get fresh data, not cache
// !res.ok - status 200
// z.array(schema).parse - Jackson into list of policies, validate against schema

export async function getPolicy(id: string): Promise<Policy | null> {
    const res = await fetch(`${API_URL}/policies/${id}`, { cache: "no-store" });
    if (res.status === 404) {
        return null;
    }
    if (!res.ok) {
        throw new Error(`GET /policies/${id} failed: ${res.status}`);
    }
    return PolicySchema.parse(await res.json()); // only need 1 object
}
